# OrderBook

Matching Engine for OrderBook written in Go.

## Feature:
- Standard price-time priority matching
- Supports both market and limit orders
- Supports order cancelling and getting order depth
- Batch matching by price level
- Memory allocation optimization

## Usage

take a look at example：([https://github.com/qky190525/orderbook-matching-engine/blob/master/main.go](https://github.com/qky190525/orderbook-matching-engine/blob/master/main.go))

Primary functions:
> func (me *MatchingEngine) PlaceOrder(order *orderbook.Order) ([]orderbook.MatchEvent, error) {...}
> 
> func (me *MatchingEngine) CancelOrder(orderID uint64) error {...}
> 
> func (me *MatchingEngine) GetDepth(limit int) *orderbook.DepthSnapshot {...}
