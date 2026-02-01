package engine

import (
	"orderbook-matching-engine/orderbook"
	"time"
)

const ()

type MatchingEngine struct {
	OrderBook *orderbook.OrderBook
}

// Option defines a functional option for configuring MatchingEngine
type Option func(*MatchingEngine)

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine() *MatchingEngine {
	me := &MatchingEngine{
		OrderBook: orderbook.NewOrderBook(),
	}
	return me
}

// Stop gracefully shuts down the engine (no-op in synchronous mode)
func (me *MatchingEngine) Stop() {
	// No background routines to stop
}

func (me *MatchingEngine) PlaceOrder(order *orderbook.Order) ([]orderbook.MatchEvent, error) {
	// Validation
	if order.ID == 0 {
		return nil, ErrOrderIDNotSet
	}
	if order.Size <= 0 {
		return nil, ErrInvalidOrderSize
	}
	if order.Type == orderbook.Limit && order.Price <= 0 {
		return nil, ErrInvalidLimitOrderPrice
	}

	// Web3 deterministic requirement: Timestamp must be provided (e.g. block time)
	// For off-chain matching, we allow flexible timestamp.
	// If not provided (0), use local time (high performance, non-deterministic).
	// If provided (e.g. from sequencer/block), use it (deterministic).
	if order.Timestamp == 0 {
		order.Timestamp = time.Now().UnixNano()
	}

	return me.processPlaceOrder(order)
}

// CancelOrder executes the cancel logic directly
func (me *MatchingEngine) CancelOrder(orderID uint64) error {
	return me.processCancelOrder(orderID)
}

// GetDepth executes the depth retrieval directly
func (me *MatchingEngine) GetDepth(limit int) *orderbook.DepthSnapshot {
	return me.OrderBook.GetDepth(limit)
}

func (me *MatchingEngine) processPlaceOrder(order *orderbook.Order) ([]orderbook.MatchEvent, error) {
	// In Web3 context, this should come from the block timestamp, not system time
	matchTime := order.Timestamp

	// Pre-allocate slice
	events := orderbook.GetMatchEventSlice()
	matchCount := 0

	// Matching Logic
	for order.Size > 0 {
		var bestLevelQueue *orderbook.OrderQueue
		var priceKey int64

		// Find best price level Price priority
		if order.Side == orderbook.Buy {
			// Buying: Look for lowest Sell (Ask)
			me.OrderBook.Asks.Range(func(key int64, value interface{}) bool {
				bestLevelQueue = value.(*orderbook.OrderQueue)
				priceKey = key
				return false // Stop after first
			})
		} else {
			// Selling: Look for highest Buy (Bid)
			me.OrderBook.Bids.Range(func(key int64, value interface{}) bool {
				bestLevelQueue = value.(*orderbook.OrderQueue)
				priceKey = key
				return false // Stop after first
			})
		}

		if bestLevelQueue == nil {
			break
		}

		bestLevelHead := bestLevelQueue.Head
		if bestLevelHead == nil {
			// If the price level exists but has no orders (empty queue), it's a zombie level.
			// We remove this empty level from the counter-party book and continue to find the next best price.
			// This does NOT delete the incoming order.
			if order.Side == orderbook.Buy {
				me.OrderBook.Asks.Delete(priceKey)
			} else {
				me.OrderBook.Bids.Delete(priceKey)
			}
			continue
		}

		// Check price crossing
		if order.Type == orderbook.Limit {
			if order.Side == orderbook.Buy {
				if order.Price < bestLevelHead.Price {
					break
				}
			} else {
				if order.Price > bestLevelHead.Price {
					break
				}
			}
		}

		// Batch matching at this price level
		curr := bestLevelHead
		// Note: We might modify curr pointer (curr = next) inside the loop
		// So we need to handle linked list traversal and deletion correctly

		// Record if the head node of current level has changed
		levelHeadChanged := false

		for curr != nil && order.Size > 0 {
			matchSize := order.Size
			if curr.Size < matchSize {
				matchSize = curr.Size
			}

			events = append(events, orderbook.MatchEvent{
				MakerOrderID: curr.ID,
				TakerOrderID: order.ID,
				Price:        curr.Price,
				Size:         matchSize,
				Timestamp:    matchTime,
			})
			matchCount++

			// Update sizes
			order.Size -= matchSize
			curr.Size -= matchSize

			if curr.Size == 0 {
				// Maker order filled
				me.OrderBook.OrderMap.Delete(curr.ID)

				// Move to next
				next := curr.Next
				curr.Next = nil // Help GC

				// Recycle object
				orderbook.PutOrder(curr)

				curr = next

				levelHeadChanged = true
			} else {
				// Maker partial fill -> Taker must be filled
				break
			}
		}

		// Update SkipMap
		if curr == nil {
			// Level exhausted
			if order.Side == orderbook.Buy {
				me.OrderBook.Asks.Delete(priceKey)
			} else {
				me.OrderBook.Bids.Delete(priceKey)
			}
		} else if levelHeadChanged {
			// Head changed but level not exhausted
			// Update the head of the list in SkipMap
			// Since bestLevelQueue is a pointer to the value in the map, we can update it directly
			bestLevelQueue.Head = curr
		}
	}

	// If remainder exists
	if order.Size > 0 {
		if order.Type == orderbook.Limit {
			// Add to book
			me.OrderBook.OrderMap.Store(order.ID, order)
			me.OrderBook.AddMakerOrder(order)
		}
		// Market Order remainder is cancelled (IOC)
		// Do nothing, just return events
	}

	return events, nil
}

// processCancelOrder is the internal cancel logic
func (me *MatchingEngine) processCancelOrder(orderID uint64) error {
	_, found := me.OrderBook.RemoveOrder(orderID)
	if !found {
		return ErrOrderNotFound
	}
	return nil
}
