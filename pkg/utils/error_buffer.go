package utils

import (
	"errors"
	"sync"
)

// ErrorBuffer uses joinedErrors interface to join multiple errors into a single error.
// This is useful to track the most recent N errors in a service and flush them as a single error.
type ErrorBuffer struct {
	// buffer is a slice of errors
	buffer []error

	// cap is the maximum number of errors that the buffer can hold.
	// Exceeding the cap results in discarding the oldest error
	cap int

	mu sync.RWMutex
}

func (eb *ErrorBuffer) Flush() (err error) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	err = errors.Join(eb.buffer...)
	eb.buffer = nil
	return
}

func (eb *ErrorBuffer) Append(incoming error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.buffer) == eb.cap && eb.cap != 0 {
		eb.buffer = append(eb.buffer[1:], incoming)
		return
	}
	eb.buffer = append(eb.buffer, incoming)
}

func (eb *ErrorBuffer) SetCap(cap int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if len(eb.buffer) > cap {
		eb.buffer = eb.buffer[len(eb.buffer)-cap:]
	}
	eb.cap = cap
}
