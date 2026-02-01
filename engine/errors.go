package engine

import "errors"

var (
	// ErrOrderIDNotSet returned when order ID is not set
	ErrOrderIDNotSet = errors.New("order ID must be set")
	// ErrInvalidOrderSize returned when order size is invalid
	ErrInvalidOrderSize = errors.New("invalid order size")
	// ErrInvalidLimitOrderPrice returned when limit order price is invalid
	ErrInvalidLimitOrderPrice = errors.New("invalid limit order price")
	// ErrOrderNotFound returned when order is not found
	ErrOrderNotFound = errors.New("order not found")
	// ErrOrderDuplicate returned when order already exists
	ErrOrderDuplicate = errors.New("order hash already exists")
	// ErrTimestampRequired returned when timestamp is not set (Web3 deterministic requirement)
	ErrTimestampRequired = errors.New("timestamp is required for deterministic execution")
)
