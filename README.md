# OrderBook

Matching Engine for OrderBook written in Go.

## Feature:
- Standard price-time priority matching
- Supports both market and limit orders
- Supports order cancelling and getting order depth
- Batch matching by price level
- Memory allocation optimization

## Usage

Take a look at example：([https://github.com/qky190525/orderbook-matching-engine/blob/master/main.go](https://github.com/qky190525/orderbook-matching-engine/blob/master/main.go))

Primary functions:
> func (me *MatchingEngine) PlaceOrder(order *orderbook.Order) ([]orderbook.MatchEvent, error) {...}
> 
> func (me *MatchingEngine) CancelOrder(orderID uint64) error {...}
> 
> func (me *MatchingEngine) GetDepth(limit int) *orderbook.DepthSnapshot {...}

```
--- Placing Ask Orders ---
Placed Ask Order 1: Price=50000.00, Size=1.00
Placed Ask Order 2: Price=50100.00, Size=0.50
Placed Ask Order 3: Price=49000.00, Size=0.10
Skipping duplicate order 1 (Hash: hash_ask_1)
  -> Current Depth:
     ASKS (Sells):
       Price: 50100.00 | Size: 0.50
       Price: 50000.00 | Size: 1.00
       Price: 49000.00 | Size: 0.10
     BIDS (Buys):

--- Placing Aggressive Bid Order ---
Placing Bid Order 4: Price=50200.00, Size=1.20
  -> Match Events:
     Maker:3 Taker:4 Price:49000.00 Size:0.10
     Maker:1 Taker:4 Price:50000.00 Size:1.00
     Maker:2 Taker:4 Price:50100.00 Size:0.10
  -> Current Depth:
     ASKS (Sells):
       Price: 50100.00 | Size: 0.40
     BIDS (Buys):

--- Canceling Remaining Order 2 ---
Order 2 canceled successfully
  -> Current Depth:
     ASKS (Sells):
     BIDS (Buys):
```