package batch

import (
	"slices"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// buffer is a thread-safe buffer for accumulating messages before sending as a batch
type buffer struct {
	messages []*messageWithCallback
	capacity int
	mu       sync.RWMutex
}

type messageWithCallback struct {
	event    *chipingress.CloudEventPb
	callback func(error)
}

// newBuffer creates a buffer with the given capacity
func newBuffer(capacity int) *buffer {
	return &buffer{
		messages: make([]*messageWithCallback, 0, capacity),
		capacity: capacity,
	}
}

// Add appends a message to the batch
func (b *buffer) Add(msg *messageWithCallback) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
}

// Len returns the current number of messages in the batch
func (b *buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.messages)
}

// Values returns a copy of the current messages without clearing them
func (b *buffer) Values() []*messageWithCallback {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.messages)
}

// Clear removes all messages from the batch and returns a copy of them
func (b *buffer) Clear() []*messageWithCallback {
	b.mu.Lock()
	defer b.mu.Unlock()
	cloned := slices.Clone(b.messages)
	b.messages = make([]*messageWithCallback, 0, b.capacity)
	return cloned
}
