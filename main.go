package main

import (
	"fmt"
	"time"

	"orderbook-matching-engine/engine"
	"orderbook-matching-engine/orderbook"
)

func main() {
	me := engine.NewMatchingEngine()
	defer me.Stop()

	// Initialize Idempotency Manager
	idempotencyMgr := engine.NewDefaultInMemoryIdempotencyManager()

	// 1. Add some Asks (Sells)
	fmt.Println("\n--- Placing Ask Orders ---")
	asks := []*orderbook.Order{
		{ID: 1, Price: 50000 * 1e8, Size: 1 * 1e8, Side: orderbook.Sell, Timestamp: time.Now().UnixNano(), OrderHash: "hash_ask_1"},
		{ID: 2, Price: 50100 * 1e8, Size: 0.5 * 1e8, Side: orderbook.Sell, Timestamp: time.Now().UnixNano(), OrderHash: "hash_ask_2"},
		{ID: 3, Price: 49000 * 1e8, Size: 0.1 * 1e8, Side: orderbook.Sell, Timestamp: time.Now().UnixNano(), OrderHash: "hash_ask_3"},
		{ID: 1, Price: 50000 * 1e8, Size: 1 * 1e8, Side: orderbook.Sell, Timestamp: time.Now().UnixNano(), OrderHash: "hash_ask_1"}, // Duplicate order
	}

	for _, order := range asks {
		// Idempotency Check
		if idempotencyMgr.Contains(order.OrderHash) {
			fmt.Printf("Skipping duplicate order %d (Hash: %s)\n", order.ID, order.OrderHash)
			continue
		}

		events, err := me.PlaceOrder(order)
		if err != nil {
			fmt.Printf("Error placing order %d: %v\n", order.ID, err)
		} else {
			// Record success
			idempotencyMgr.Add(order.OrderHash)

			fmt.Printf("Placed Ask Order %d: Price=%.2f, Size=%.2f\n",
				order.ID, float64(order.Price)/1e8, float64(order.Size)/1e8)
			printEvents(events)
		}
	}

	printDepth(me)

	// 2. Place a Bid (Buy) that matches partially
	fmt.Println("\n--- Placing Aggressive Bid Order ---")
	bidOrder := &orderbook.Order{
		ID:        4,
		Price:     50200 * 1e8,
		Size:      1.2 * 1e8,
		Side:      orderbook.Buy,
		Timestamp: time.Now().UnixNano(),
		OrderHash: "hash_bid_4",
	}
	fmt.Printf("Placing Bid Order %d: Price=%.2f, Size=%.2f\n",
		bidOrder.ID, float64(bidOrder.Price)/1e8, float64(bidOrder.Size)/1e8)

	// Idempotency Check for Bid
	if idempotencyMgr.Contains(bidOrder.OrderHash) {
		fmt.Printf("Skipping duplicate bid order %d (Hash: %s)\n", bidOrder.ID, bidOrder.OrderHash)
	} else {
		events, err := me.PlaceOrder(bidOrder)
		if err != nil {
			fmt.Printf("Error placing bid: %v\n", err)
		} else {
			idempotencyMgr.Add(bidOrder.OrderHash)
			printEvents(events)
		}
	}
	printDepth(me)

	// 3. Cancel remaining part of Order 2 (if any) - Actually Order 2 is partially filled?

	fmt.Println("\n--- Canceling Remaining Order 2 ---")
	err := me.CancelOrder(2)
	if err != nil {
		fmt.Printf("Cancel failed: %v\n", err)
	} else {
		fmt.Println("Order 2 canceled successfully")
	}
	printDepth(me)
}

func printEvents(events []orderbook.MatchEvent) {
	if len(events) == 0 {
		return
	}
	fmt.Println("  -> Match Events:")
	for _, e := range events {
		fmt.Printf("     Maker:%d Taker:%d Price:%.2f Size:%.2f\n",
			e.MakerOrderID, e.TakerOrderID, float64(e.Price)/1e8, float64(e.Size)/1e8)
	}
}

func printDepth(me *engine.MatchingEngine) {
	depth := me.GetDepth(5)
	fmt.Println("  -> Current Depth:")
	fmt.Println("     ASKS (Sells):")
	// Print Asks in reverse (high to low) usually for UI, but here standard order
	for i := len(depth.Asks) - 1; i >= 0; i-- {
		l := depth.Asks[i]
		fmt.Printf("       Price: %.2f | Size: %.2f\n", float64(l.Price)/1e8, float64(l.Size)/1e8)
	}
	fmt.Println("     BIDS (Buys):")
	for _, l := range depth.Bids {
		fmt.Printf("       Price: %.2f | Size: %.2f\n", float64(l.Price)/1e8, float64(l.Size)/1e8)
	}
}
