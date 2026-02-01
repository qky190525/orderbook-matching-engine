package engine

import "sync"

// IdempotencyManager defines the interface for order idempotency checking
type IdempotencyManager interface {
	// Contains checks if the order hash already exists
	Contains(hash string) bool
	// Add registers a new order hash
	Add(hash string)
}

// defaultInMemoryIdempotencyManager provides a thread-safe in-memory implementation
type defaultInMemoryIdempotencyManager struct {
	mu     sync.RWMutex
	hashes map[string]struct{}
}

// NewDefaultInMemoryIdempotencyManager provides a thread-safe in-memory implementation
func NewDefaultInMemoryIdempotencyManager() IdempotencyManager {
	return &defaultInMemoryIdempotencyManager{
		hashes: make(map[string]struct{}),
	}
}

func (m *defaultInMemoryIdempotencyManager) Contains(hash string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.hashes[hash]
	return exists
}

func (m *defaultInMemoryIdempotencyManager) Add(hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hashes[hash] = struct{}{}
}
