package orderbook

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Side represents the order side (Buy or Sell)
type Side int

const (
	Buy Side = iota
	Sell
)

func (s Side) String() string {
	if s == Buy {
		return "Buy"
	}
	return "Sell"
}

func (s Side) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Side) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "buy":
		*s = Buy
	case "sell":
		*s = Sell
	default:
		return fmt.Errorf("invalid side: %s", str)
	}
	return nil
}

// OrderType represents the type of order
type OrderType int

const (
	Limit OrderType = iota
	Market
)

func (t OrderType) String() string {
	if t == Market {
		return "Market"
	}
	return "Limit"
}

func (t OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *OrderType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "limit":
		*t = Limit
	case "market":
		*t = Market
	default:
		return fmt.Errorf("invalid order type: %s", str)
	}
	return nil
}

// Order represents an order in the system
type Order struct {
	ID        uint64    `json:"id"`
	UserID    string    `json:"user_id"`
	OrderHash string    `json:"order_hash"`
	Type      OrderType `json:"type"`
	Price     int64     `json:"price"` // Fixed-point representation (e.g., * 1e8)
	Size      int64     `json:"size"`  // Fixed-point representation
	Side      Side      `json:"side"`
	Timestamp int64     `json:"timestamp"` // Unix nanoseconds
	Next      *Order    `json:"-"`         // For SkipList/Linked List linking
}

// MatchEvent represents a trade execution
type MatchEvent struct {
	MakerOrderID uint64 `json:"maker_order_id"`
	TakerOrderID uint64 `json:"taker_order_id"`
	Price        int64  `json:"price"`
	Size         int64  `json:"size"`
	Timestamp    int64  `json:"timestamp"`
}

// DepthSnapshot represents the current state of the order book
type DepthSnapshot struct {
	Asks []PriceLevel `json:"asks"`
	Bids []PriceLevel `json:"bids"`
}

type PriceLevel struct {
	Price int64 `json:"price"`
	Size  int64 `json:"size"`
}
