package orderbook

import (
	"sync"
)

// OrderPool manages a pool of Order objects to reduce GC pressure
var OrderPool = sync.Pool{
	New: func() interface{} {
		return &Order{}
	},
}

// GetOrder retrieves an order from the pool
func GetOrder() *Order {
	return OrderPool.Get().(*Order)
}

// PutOrder returns an order to the pool after resetting it
func PutOrder(o *Order) {
	o.Reset()
	OrderPool.Put(o)
}

// Reset clears the order fields for reuse
func (o *Order) Reset() {
	o.ID = 0
	o.UserID = ""
	o.OrderHash = ""
	o.Type = Limit // Default
	o.Price = 0
	o.Size = 0
	o.Side = Buy // Default
	o.Timestamp = 0
	o.Next = nil
}

// MatchEventPool manages a pool of MatchEvent objects
// Note: Since MatchEvents are typically returned in a slice, handling individual
// struct pointers might be overhead. However, if we change MatchEvent to be a pointer
// or just reuse the slice backing array, that's another optimization.
// For now, let's keep MatchEvent as a struct (value type) in the slice.
// Optimizing slice allocation is more important.
// But if the user strictly wants object pooling for MatchEvent, we can provide it.
// Given MatchEvent is small (40 bytes), value semantics are often fine.
// But let's provide a pool for []MatchEvent slices if needed, or just follow the instruction.

// The instruction says "reuse Order and MatchEvent objects".
// Since processPlaceOrder returns []MatchEvent, reusing the slice buffer is key.
var MatchEventSlicePool = sync.Pool{
	New: func() interface{} {
		// Default capacity 16 as seen in matching.go
		return make([]MatchEvent, 0, 16)
	},
}

func GetMatchEventSlice() []MatchEvent {
	return MatchEventSlicePool.Get().([]MatchEvent)
}

func PutMatchEventSlice(events []MatchEvent) {
	events = events[:0] // Reset length
	MatchEventSlicePool.Put(events)
}
