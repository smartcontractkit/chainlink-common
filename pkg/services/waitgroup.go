package services

import (
	"errors"
	"sync"
)

// WaitGroup is like [sync.WaitGroup], but TryAdd may be called after Wait and will return error rather than cause a race.
type WaitGroup struct {
	mu      sync.Mutex
	cond    sync.Cond
	count   int
	waiting bool
}

func (t *WaitGroup) ensureCond() { t.cond.L = &t.mu }

// TryAdd increments the internal count by n, or returns an error if Wait has already been called.
// If successful, Done must be called n times when complete.
func (t *WaitGroup) TryAdd(n int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.waiting {
		return errors.New("stopped")
	}
	t.count += n

	return nil
}

// Done decrements the internal count.
func (t *WaitGroup) Done() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.count--
	if t.count == 0 {
		t.ensureCond()
		t.cond.Signal()
	}
}

// Wait halts future additions and then blocks until all Done calls complete.
func (t *WaitGroup) Wait() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.waiting = true
	if t.count != 0 {
		t.ensureCond()
		t.cond.Wait()
	}
}
