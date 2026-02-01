package orderbook

import (
	"testing"
)

func TestOrderBook(t *testing.T) {
	ob := NewOrderBook()

	// Test AddMakerOrder
	o1 := &Order{ID: 1, Price: 100, Size: 10, Side: Buy}
	ob.AddMakerOrder(o1)

	// Test GetOrder
	if got, exists := ob.GetOrder(1); !exists || got != o1 {
		t.Errorf("GetOrder(1) failed")
	}

	// Test RemoveOrder
	if _, ok := ob.RemoveOrder(1); !ok {
		t.Errorf("RemoveOrder(1) failed")
	}
	if _, exists := ob.GetOrder(1); exists {
		t.Errorf("Order 1 should be removed")
	}

	// Test BatchAddMakerOrders
	orders := []*Order{
		{ID: 2, Price: 90, Size: 5, Side: Buy},
		{ID: 3, Price: 110, Size: 5, Side: Sell},
		{ID: 4, Price: 90, Size: 10, Side: Buy}, // Same price as ID 2
	}
	ob.BatchAddMakerOrders(orders)

	// Verify Depth
	depth := ob.GetDepth(10)
	if len(depth.Bids) != 1 || depth.Bids[0].Price != 90 || depth.Bids[0].Size != 15 {
		t.Errorf("Bids depth incorrect: %v", depth.Bids)
	}
	if len(depth.Asks) != 1 || depth.Asks[0].Price != 110 {
		t.Errorf("Asks depth incorrect")
	}

	// Test GetBestBid/Ask
	if ob.GetBestBid().ID != 2 { // First inserted at 90
		t.Errorf("BestBid ID expected 2, got %d", ob.GetBestBid().ID)
	}
	if ob.GetBestAsk().ID != 3 {
		t.Errorf("BestAsk ID expected 3, got %d", ob.GetBestAsk().ID)
	}
}
