package utils

import (
	"context"
	"fmt"
	"math"
	mrand "math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// WithJitter adds +/- 10% to a duration
func WithJitter(d time.Duration) time.Duration {
	// #nosec
	if d == 0 {
		return 0
	}
	// ensure non-zero arg to Intn to avoid panic
	max := math.Max(float64(d.Abs())/5.0, 1.)
	// #nosec - non critical randomness
	jitter := mrand.Intn(int(max))
	jitter = jitter - (jitter / 2)
	return time.Duration(int(d) + jitter)
}

// ContextFromChan creates a context that finishes when the provided channel
// receives or is closed.
// When channel closes, the ctx.Err() will always be context.Canceled
// NOTE: Spins up a goroutine that exits on cancellation.
// REMEMBER TO CALL CANCEL OTHERWISE IT CAN LEAD TO MEMORY LEAKS
func ContextFromChan(chStop <-chan struct{}) (context.Context, context.CancelFunc) {
	return services.StopRChan(chStop).NewCtx()
}

// ContextWithDeadlineFn returns a copy of the parent context with the deadline modified by deadlineFn.
// deadlineFn will only be called if the parent has a deadline.
// The new deadline must be sooner than the old to have an effect.
func ContextWithDeadlineFn(ctx context.Context, deadlineFn func(orig time.Time) time.Time) (context.Context, context.CancelFunc) {
	cancel := func() {}
	if d, ok := ctx.Deadline(); ok {
		if m := deadlineFn(d); m.Before(d) {
			ctx, cancel = context.WithDeadline(ctx, m)
		}
	}
	return ctx, cancel
}

func IsZero[C comparable](val C) bool {
	var zero C
	return zero == val
}

// EnsureHexPrefix adds the prefix (0x) to a given hex string.
func EnsureHexPrefix(str string) string {
	if !strings.HasPrefix(str, "0x") {
		str = "0x" + str
	}
	return str
}

// BoundedQueue is a FIFO queue that discards older items when it reaches its capacity.
type BoundedQueue[T any] struct {
	capacity int
	items    []T
	mu       sync.RWMutex
}

// NewBoundedQueue creates a new BoundedQueue instance
func NewBoundedQueue[T any](capacity int) *BoundedQueue[T] {
	var bq BoundedQueue[T]
	bq.capacity = capacity
	return &bq
}

// Add appends items to a BoundedQueue
func (q *BoundedQueue[T]) Add(x T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, x)
	if len(q.items) > q.capacity {
		excess := len(q.items) - q.capacity
		q.items = q.items[excess:]
	}
}

// Take pulls the first item from the array and removes it
func (q *BoundedQueue[T]) Take() (t T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return
	}
	t = q.items[0]
	q.items = q.items[1:]
	return
}

// Empty check is a BoundedQueue is empty
func (q *BoundedQueue[T]) Empty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items) == 0
}

// Full checks if a BoundedQueue is over capacity.
func (q *BoundedQueue[T]) Full() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items) >= q.capacity
}

// BoundedPriorityQueue stores a series of BoundedQueues
// with associated priorities and capacities
type BoundedPriorityQueue[T any] struct {
	queues     map[uint]*BoundedQueue[T]
	priorities []uint
	capacities map[uint]int
	mu         sync.RWMutex
}

// NewBoundedPriorityQueue creates a new BoundedPriorityQueue
func NewBoundedPriorityQueue[T any](capacities map[uint]int) *BoundedPriorityQueue[T] {
	queues := make(map[uint]*BoundedQueue[T])
	var priorities []uint
	for priority, capacity := range capacities {
		priorities = append(priorities, priority)
		queues[priority] = NewBoundedQueue[T](capacity)
	}
	sort.Slice(priorities, func(i, j int) bool { return priorities[i] < priorities[j] })
	bpq := BoundedPriorityQueue[T]{
		queues:     queues,
		priorities: priorities,
		capacities: capacities,
	}
	return &bpq
}

// Add pushes an item into a subque within a BoundedPriorityQueue
func (q *BoundedPriorityQueue[T]) Add(priority uint, x T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	subqueue, exists := q.queues[priority]
	if !exists {
		panic(fmt.Sprintf("nonexistent priority: %v", priority))
	}

	subqueue.Add(x)
}

// Take takes from the BoundedPriorityQueue's subque
func (q *BoundedPriorityQueue[T]) Take() (t T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, priority := range q.priorities {
		queue := q.queues[priority]
		if queue.Empty() {
			continue
		}
		return queue.Take()
	}
	return
}

// Empty checks the BoundedPriorityQueue
// if all subqueues are empty
func (q *BoundedPriorityQueue[T]) Empty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, priority := range q.priorities {
		queue := q.queues[priority]
		if !queue.Empty() {
			return false
		}
	}
	return true
}
