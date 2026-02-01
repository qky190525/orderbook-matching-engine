package orderbook

import (
	"github.com/bytedance/gopkg/collection/skipmap"
)

type OrderQueue struct {
	Head *Order
	Tail *Order
}

type OrderBook struct {
	Asks     *skipmap.Int64Map  // Selling: Price -> *OrderQueue (Ascending)
	Bids     *skipmap.Int64Map  // Buying: -Price -> *OrderQueue (Ascending -Price = Descending Price)
	OrderMap *skipmap.Uint64Map // OrderID -> *Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:     skipmap.NewInt64(),
		Bids:     skipmap.NewInt64(),
		OrderMap: skipmap.NewUint64(),
	}
}

// AddMakerOrder adds a liquidity providing order to the book
func (ob *OrderBook) AddMakerOrder(order *Order) {
	ob.OrderMap.Store(order.ID, order)
	ob.addOrderToSkipMap(order)
}

func (ob *OrderBook) addOrderToSkipMap(order *Order) {
	var sm *skipmap.Int64Map
	var key int64

	if order.Side == Buy {
		sm = ob.Bids
		key = -order.Price // Negate price for descending order
	} else {
		sm = ob.Asks
		key = order.Price
	}

	// Try to store as new queue
	newQ := &OrderQueue{Head: order, Tail: order}
	actual, loaded := sm.LoadOrStore(key, newQ)

	if loaded {
		// List already exists, append to tail using O(1) access
		q := actual.(*OrderQueue)
		q.Tail.Next = order
		q.Tail = order
	}
}

// RemoveOrder removes an order by ID
func (ob *OrderBook) RemoveOrder(orderID uint64) (*Order, bool) {
	val, exists := ob.OrderMap.LoadAndDelete(orderID)
	if !exists {
		return nil, false
	}
	order := val.(*Order)

	ob.removeOrderFromSkipMap(order)
	return order, true
}

func (ob *OrderBook) removeOrderFromSkipMap(order *Order) {
	var sm *skipmap.Int64Map
	var key int64

	if order.Side == Buy {
		sm = ob.Bids
		key = -order.Price
	} else {
		sm = ob.Asks
		key = order.Price
	}

	val, ok := sm.Load(key)
	if !ok {
		return
	}
	q := val.(*OrderQueue)
	head := q.Head

	if head.ID == order.ID {
		// Removing head
		if head.Next == nil {
			// Queue becomes empty
			sm.Delete(key)
		} else {
			q.Head = head.Next
			// If we removed the head and it was the only element, we deleted the key above.
			// If we removed the head and there are more elements, Tail remains the same (unless only 1 left? No, Tail points to end).
			// If head.Next != nil, then Tail is still valid (it's either head.Next or further down).
		}
	} else {
		// Remove from middle
		curr := head
		for curr.Next != nil {
			if curr.Next.ID == order.ID {
				// Found it
				if curr.Next == q.Tail {
					// Removing tail, update Tail pointer
					q.Tail = curr
				}
				curr.Next = curr.Next.Next
				break
			}
			curr = curr.Next
		}
	}
	// Clear next pointer of removed order to avoid memory leaks/dangling pointers
	order.Next = nil
}

// GetOrder returns an order by ID (Lock-free read)
func (ob *OrderBook) GetOrder(orderID uint64) (*Order, bool) {
	val, ok := ob.OrderMap.Load(orderID)
	if !ok {
		return nil, false
	}
	return val.(*Order), true
}

// BatchAddMakerOrders adds multiple orders to the book with a single lock acquisition
func (ob *OrderBook) BatchAddMakerOrders(orders []*Order) {
	for _, order := range orders {
		ob.OrderMap.Store(order.ID, order)
		ob.addOrderToSkipMap(order)
	}
}

// GetBestBid returns the highest buy order
func (ob *OrderBook) GetBestBid() *Order {
	var best *Order
	ob.Bids.Range(func(key int64, value interface{}) bool {
		q := value.(*OrderQueue)
		best = q.Head
		return false // Stop after first
	})
	return best
}

// GetBestAsk returns the lowest sell order
func (ob *OrderBook) GetBestAsk() *Order {
	var best *Order
	ob.Asks.Range(func(key int64, value interface{}) bool {
		q := value.(*OrderQueue)
		best = q.Head
		return false // Stop after first
	})
	return best
}

// GetDepth returns the snapshot of the order book
func (ob *OrderBook) GetDepth(limit int) *DepthSnapshot {
	snapshot := &DepthSnapshot{
		Asks: make([]PriceLevel, 0, limit),
		Bids: make([]PriceLevel, 0, limit),
	}

	// Traverse Asks
	count := 0
	ob.Asks.Range(func(key int64, value interface{}) bool {
		if count >= limit {
			return false
		}
		totalSize := int64(0)
		q := value.(*OrderQueue)
		ord := q.Head
		for ord != nil {
			totalSize += ord.Size
			ord = ord.Next
		}
		snapshot.Asks = append(snapshot.Asks, PriceLevel{Price: key, Size: totalSize})
		count++
		return true
	})

	// Traverse Bids
	count = 0
	ob.Bids.Range(func(key int64, value interface{}) bool {
		if count >= limit {
			return false
		}
		totalSize := int64(0)
		q := value.(*OrderQueue)
		ord := q.Head
		for ord != nil {
			totalSize += ord.Size
			ord = ord.Next
		}
		// Key is -Price, so we negate it back
		snapshot.Bids = append(snapshot.Bids, PriceLevel{Price: -key, Size: totalSize})
		count++
		return true
	})

	return snapshot
}
