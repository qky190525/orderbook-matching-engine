package engine

import (
	"orderbook-matching-engine/orderbook"
	"testing"
)

func TestMatchingEngine_BatchMatching(t *testing.T) {
	me := NewMatchingEngine()

	// 1. Setup OrderBook with multiple orders at same price level
	// Sell orders (Asks) at price 100
	asks := []*orderbook.Order{
		{ID: 1, Price: 100, Size: 10, Side: orderbook.Sell},
		{ID: 2, Price: 100, Size: 10, Side: orderbook.Sell},
		{ID: 3, Price: 100, Size: 10, Side: orderbook.Sell},
		// Another price level
		{ID: 4, Price: 101, Size: 20, Side: orderbook.Sell},
	}

	for _, o := range asks {
		me.OrderBook.AddMakerOrder(o)
	}

	// 2. Place a large Buy order that consumes multiple orders at 100 and part of 101
	// Total Ask at 100 is 30. Order size 35. Should consume all 30 at 100, and 5 at 101.
	takerOrder := &orderbook.Order{
		ID:        100, // ID will be generated
		Timestamp: 1000,
		Price:     102, // Willing to buy up to 102
		Size:      35,
		Side:      orderbook.Buy,
	}

	events, err := me.PlaceOrder(takerOrder)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// 3. Verify Matches
	if len(events) != 4 {
		t.Errorf("Expected 4 match events, got %d", len(events))
	}

	totalMatched := int64(0)
	for _, e := range events {
		totalMatched += e.Size
		t.Logf("Match: Price %d, Size %d, MakerID %d", e.Price, e.Size, e.MakerOrderID)
	}

	if totalMatched != 35 {
		t.Errorf("Expected total matched 35, got %d", totalMatched)
	}

	// Verify specific matches
	// We expect ID 1, 2, 3 to be fully matched at 100
	// We expect ID 4 to be partially matched (5) at 101

	// Since we use FIFO, order should be 1, 2, 3, then 4
	if events[0].MakerOrderID != 1 || events[0].Price != 100 {
		t.Errorf("Event 0 mismatch")
	}
	if events[3].MakerOrderID != 4 || events[3].Price != 101 || events[3].Size != 5 {
		t.Errorf("Event 3 mismatch")
	}

	// 4. Verify OrderBook State
	// ID 1, 2, 3 should be gone
	if _, ok := me.OrderBook.GetOrder(1); ok {
		t.Errorf("Order 1 should be removed")
	}
	// ID 4 should remain with Size 15
	o4, ok := me.OrderBook.GetOrder(4)
	if !ok {
		t.Errorf("Order 4 should exist")
	} else if o4.Size != 15 {
		t.Errorf("Order 4 size should be 15, got %d", o4.Size)
	}

	// Verify Depth
	depth := me.GetDepth(10)
	if len(depth.Asks) != 1 {
		t.Errorf("Expected 1 Ask level, got %d", len(depth.Asks))
	} else {
		if depth.Asks[0].Price != 101 || depth.Asks[0].Size != 15 {
			t.Errorf("Ask level incorrect: %v", depth.Asks[0])
		}
	}
}

func TestMatchingEngine_PartialLevelFill(t *testing.T) {
	me := NewMatchingEngine()
	// Sell orders at 100: ID 1 (10), ID 2 (10)
	me.OrderBook.AddMakerOrder(&orderbook.Order{ID: 1, Price: 100, Size: 10, Side: orderbook.Sell})
	me.OrderBook.AddMakerOrder(&orderbook.Order{ID: 2, Price: 100, Size: 10, Side: orderbook.Sell})

	// Buy 15. Should fill ID 1 (10) and ID 2 (5).
	taker := &orderbook.Order{ID: 100, Price: 100, Size: 15, Side: orderbook.Buy, Timestamp: 1000}
	events, _ := me.PlaceOrder(taker)

	if len(events) != 2 {
		t.Errorf("Expected 2 events")
	}

	// ID 1 gone
	if _, ok := me.OrderBook.GetOrder(1); ok {
		t.Errorf("Order 1 should be gone")
	}
	// ID 2 remains (5)
	o2, ok := me.OrderBook.GetOrder(2)
	if !ok || o2.Size != 5 {
		t.Errorf("Order 2 should remain with size 5")
	}

	// Verify SkipMap points to ID 2
	best := me.OrderBook.GetBestAsk()
	if best.ID != 2 {
		t.Errorf("Best Ask should be ID 2, got %d", best.ID)
	}
}

func TestMatchingEngine_MarketOrder(t *testing.T) {
	me := NewMatchingEngine()

	// Setup OrderBook
	// Sell orders at 100, 101, 102
	me.OrderBook.AddMakerOrder(&orderbook.Order{ID: 1, Price: 100, Size: 10, Side: orderbook.Sell})
	me.OrderBook.AddMakerOrder(&orderbook.Order{ID: 2, Price: 101, Size: 10, Side: orderbook.Sell})
	me.OrderBook.AddMakerOrder(&orderbook.Order{ID: 3, Price: 102, Size: 10, Side: orderbook.Sell})

	// 1. Market Buy Order - Full Fill
	// Buy 15. Should fill 10 @ 100 and 5 @ 101.
	marketBuy := &orderbook.Order{
		ID: 99, Type: orderbook.Market, Size: 15, Side: orderbook.Buy, Timestamp: 1000,
	}
	events, err := me.PlaceOrder(marketBuy)
	if err != nil {
		t.Fatalf("Market Buy failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 match events, got %d", len(events))
	}
	if events[0].Price != 100 || events[0].Size != 10 {
		t.Errorf("Event 0 mismatch: %v", events[0])
	}
	if events[1].Price != 101 || events[1].Size != 5 {
		t.Errorf("Event 1 mismatch: %v", events[1])
	}

	// Verify ID 1 is gone, ID 2 has 5 left
	if _, ok := me.OrderBook.GetOrder(1); ok {
		t.Errorf("Order 1 should be gone")
	}
	if o2, ok := me.OrderBook.GetOrder(2); !ok || o2.Size != 5 {
		t.Errorf("Order 2 should have size 5")
	}

	// 2. Market Buy Order - Partial Fill (Liquidity Exhausted)
	// Current Book: ID 2 (5 @ 101), ID 3 (10 @ 102) -> Total 15
	// Buy 20. Should fill 15 and CANCEL the remaining 5.
	marketBuyHuge := &orderbook.Order{
		ID: 100, Type: orderbook.Market, Size: 20, Side: orderbook.Buy, Timestamp: 1000,
	}
	events2, err := me.PlaceOrder(marketBuyHuge)
	if err != nil {
		t.Fatalf("Market Buy Huge failed: %v", err)
	}

	if len(events2) != 2 {
		t.Errorf("Expected 2 match events (exhausted book), got %d", len(events2))
	}

	// Verify Book is Empty
	if me.OrderBook.GetBestAsk() != nil {
		t.Errorf("Book should be empty of asks")
	}

	// Verify Remainder is NOT in book
	if _, ok := me.OrderBook.GetOrder(100); ok {
		t.Errorf("Market order remainder should NOT be in book")
	}
}

func TestMatchingEngine_Validation(t *testing.T) {
	me := NewMatchingEngine()

	// Limit Order with Price 0 -> Error
	limitZero := &orderbook.Order{ID: 100, Type: orderbook.Limit, Price: 0, Size: 10, Side: orderbook.Buy, Timestamp: 1000}
	if _, err := me.PlaceOrder(limitZero); err == nil {
		t.Errorf("Expected error for Limit Order with Price 0")
	}

	// Market Order with Price 0 -> OK
	marketZero := &orderbook.Order{ID: 101, Type: orderbook.Market, Price: 0, Size: 10, Side: orderbook.Buy, Timestamp: 1000}
	if _, err := me.PlaceOrder(marketZero); err != nil {
		t.Errorf("Market Order with Price 0 should be allowed, got error: %v", err)
	}
}
