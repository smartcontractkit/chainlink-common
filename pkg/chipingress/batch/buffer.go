// Package batch provides a thread-safe batching client for chip ingress messages.
package batch

import (
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// messageBatch is a thread-safe buffer for accumulating messages before sending as a batch
type messageBatch struct {
	messages []*messageWithCallback
	mu       sync.Mutex
}

type messageWithCallback struct {
	event    *chipingress.CloudEventPb
	callback func(error)
}

// newMessageBatch creates a new messageBatch with the given initial capacity
func newMessageBatch(capacity int) *messageBatch {
	return &messageBatch{
		messages: make([]*messageWithCallback, 0, capacity),
	}
}

// Add appends a message to the batch
func (b *messageBatch) Add(msg *messageWithCallback) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
}

// Len returns the current number of messages in the batch
func (b *messageBatch) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.messages)
}

// Clear removes all messages from the batch and returns a copy of them
func (b *messageBatch) Clear() []*messageWithCallback {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.messages) == 0 {
		return nil
	}

	// Make a copy
	result := make([]*messageWithCallback, len(b.messages))
	copy(result, b.messages)

	// Reset the internal slice
	b.messages = b.messages[:0]

	return result
}

// Values returns a copy of the current messages without clearing them
func (b *messageBatch) Values() []*messageWithCallback {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.messages) == 0 {
		return nil
	}

	result := make([]*messageWithCallback, len(b.messages))
	copy(result, b.messages)
	return result
}
