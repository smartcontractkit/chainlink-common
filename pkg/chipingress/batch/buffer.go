package batch

import (
	"slices"
	"sync"
)

// buffer is a thread-safe buffer for accumulating messages before sending as a batch
type buffer[T any] struct {
	messages []T
	capacity int
	mu       sync.RWMutex
}

// newBuffer creates a buffer with the given capacity
func newBuffer[T any](capacity int) *buffer[T] {
	return &buffer[T]{
		messages: make([]T, 0, capacity),
		capacity: capacity,
	}
}

// Add appends a message to the batch
func (b *buffer[T]) Add(msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
}

// Len returns the current number of messages in the batch
func (b *buffer[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.messages)
}

// Values returns a copy of the current messages without clearing them
func (b *buffer[T]) Values() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.messages)
}

// Clear removes all messages from the batch and returns a copy of them
func (b *buffer[T]) Clear() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	cloned := slices.Clone(b.messages)
	b.messages = make([]T, 0, b.capacity)
	return cloned
}
