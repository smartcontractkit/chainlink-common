package limits

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// QueueLimiter is a limiter for queues.
type QueueLimiter[T any] interface {
	Limiter[int]
	// Len returns the current size of the queue.
	Len(context.Context) (int, error)
	// Put queues the value, or returns ErrorQueueFull.
	Put(context.Context, T) error
	// Get returns the next value if available, otherwise ErrQueueEmpty.
	Get(context.Context) (T, error)
	// Wait gets the next value, waiting up until context cancellation.
	Wait(context.Context) (T, error)
}

// NewQueueLimiter returns a simple static QueueLimiter.
func NewQueueLimiter[T any](capacity int) QueueLimiter[T] {
	q := &queue[T]{
		cap:          capacity,
		recordLimit:  func(context.Context, int) {},
		recordUsage:  func(context.Context, int) {},
		recordDenied: func(context.Context, int) {},
	}
	q.cond.L = &q.mu
	return q
}

type queue[T any] struct {
	*updater[int]

	key    string         // optional
	scope  settings.Scope // optional
	tenant string         // optional

	recordLimit  func(context.Context, int)
	recordUsage  func(context.Context, int)
	recordDenied func(context.Context, int)

	cap  int
	list list.List
	mu   sync.Mutex
	cond sync.Cond
}

func (q *queue[T]) Limit(context.Context) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.cap, nil
}

func (q *queue[T]) Len(context.Context) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.list.Len(), nil
}

func (q *queue[T]) setCap(ctx context.Context, c int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cap = c
	q.recordLimit(ctx, c)
}

func (q *queue[T]) record(ctx context.Context) {
	q.recordUsage(ctx, q.list.Len())
	q.recordLimit(ctx, q.cap)
}

func (q *queue[T]) Put(ctx context.Context, t T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if c := q.cap; q.list.Len() >= c {
		q.recordDenied(ctx, 1)
		return ErrorQueueFull{Key: q.key, Scope: q.scope, Tenant: q.tenant, Limit: c}
	}
	q.list.PushBack(t)
	q.cond.Signal()
	q.record(ctx)
	return nil
}

func (q *queue[T]) Get(ctx context.Context) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.list.Len() == 0 {
		var zero T
		return zero, ErrQueueEmpty
	}
	t := q.list.Front()
	q.list.Remove(t)
	q.record(ctx)
	return t.Value.(T), nil
}

func (q *queue[T]) Wait(ctx context.Context) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.list.Len() == 0 {
		// Ensure cond.Wait() yields to context expiration
		stop := context.AfterFunc(ctx, func() {
			q.mu.Lock()
			defer q.mu.Unlock()
			q.cond.Broadcast()
		})
		defer stop()
		for q.list.Len() == 0 {
			q.cond.Wait()
			if ctx.Err() != nil {
				var zero T
				return zero, ctx.Err()
			}
		}
	}
	t := q.list.Front()
	q.list.Remove(t)
	q.record(ctx)
	return t.Value.(T), nil
}

type unscopedQueue[T any] struct {
	*queue[T]
}

func newUnscopedQueue[T any](f Factory, limit settings.Setting[int]) (QueueLimiter[T], error) {
	q := &queue[T]{
		key:     limit.Key,
		cap:     limit.DefaultValue,
		updater: newUpdater[int](nil, func(d context.Context) (int, error) { return limit.DefaultValue, nil }, nil),
	}
	q.cond.L = &q.mu
	q.updater.recordLimit = q.setCap

	if f.Logger == nil {
		q.lggr = logger.Nop()
	} else {
		q.lggr = logger.Sugared(f.Logger).Named("QueueLimiter").With("key", limit.Key)
	}

	if f.Meter != nil {
		limitGauge, err := f.Meter.Int64Gauge("queue."+limit.Key+".limit", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		usageGauge, err := f.Meter.Int64Gauge("queue."+limit.Key+".usage", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		deniedHist, err := f.Meter.Int64Histogram("queue."+limit.Key+".denied", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		q.recordLimit = func(ctx context.Context, i int) {
			limitGauge.Record(ctx, int64(i))
		}
		q.recordUsage = func(ctx context.Context, i int) {
			usageGauge.Record(ctx, int64(i))
		}
		q.recordDenied = func(ctx context.Context, i int) {
			deniedHist.Record(ctx, int64(i))
		}
	} else {
		q.recordLimit = func(context.Context, int) {}
		q.recordUsage = func(context.Context, int) {}
		q.recordDenied = func(context.Context, int) {}
	}

	if f.Settings != nil {
		q.getLimitFn = func(ctx context.Context) (int, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if registry, ok := f.Settings.(settings.Registry); ok {
			q.subFn = func(ctx context.Context) (updates <-chan settings.Update[int], cancelSub func()) {
				return limit.Subscribe(ctx, registry)
			}
		}
	}

	go q.updateLoop(context.Background())

	return unscopedQueue[T]{q}, nil
}

func newScopedQueue[T any](f Factory, limit settings.Setting[int]) (QueueLimiter[T], error) {
	q := &scopedQueue[T]{
		key:        limit.Key,
		scope:      limit.Scope,
		defaultCap: limit.DefaultValue,
	}

	if f.Logger == nil {
		q.lggr = logger.Nop()
	} else {
		q.lggr = logger.Sugared(f.Logger).Named("QueueLimiter").With("key", limit.Key, "scope", limit.Scope)
	}

	if f.Meter != nil {
		var err error
		q.limitGauge, err = f.Meter.Int64Gauge("queue."+limit.Key+".limit", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		q.usageGauge, err = f.Meter.Int64Gauge("queue."+limit.Key+".usage", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		q.deniedHist, err = f.Meter.Int64Histogram("queue."+limit.Key+".denied", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
	}

	if f.Settings != nil {
		q.capFn = func(ctx context.Context) (int, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if r, ok := f.Settings.(settings.Registry); ok {
			q.subFn = func(ctx context.Context) (<-chan settings.Update[int], func()) {
				return limit.Subscribe(ctx, r)
			}
		}
	} else {
		q.capFn = func(ctx context.Context) (int, error) {
			return limit.DefaultValue, nil
		}
	}

	return q, nil
}

type scopedQueue[T any] struct {
	lggr       logger.Logger
	scope      settings.Scope
	defaultCap int

	capFn func(context.Context) (int, error)
	subFn func(ctx context.Context) (<-chan settings.Update[int], func()) // optional

	key string // optional

	limitGauge metric.Int64Gauge     // optional
	usageGauge metric.Int64Gauge     // optional
	deniedHist metric.Int64Histogram // optional

	// opt: reap after period of non-use
	queues sync.Map           // map[string]*queue
	wg     services.WaitGroup // tracks and blocks limiters background routines
}

func (s *scopedQueue[T]) Close() (err error) {
	s.wg.Wait()

	// cleanup
	s.queues.Range(func(tenant, value any) bool {
		// opt: parallelize
		err = errors.Join(err, value.(*queue[T]).Close())
		return true
	})
	return
}

func (s *scopedQueue[T]) Limit(ctx context.Context) (int, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return -1, err
	}
	defer done()
	return l.Limit(ctx)
}

func (s *scopedQueue[T]) Len(ctx context.Context) (int, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return -1, err
	}
	defer done()
	return l.Len(ctx)
}

func (s *scopedQueue[T]) Put(ctx context.Context, t T) error {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return l.Put(ctx, t)
}

func (s *scopedQueue[T]) Get(ctx context.Context) (T, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	defer done()
	return l.Get(ctx)
}

func (s *scopedQueue[T]) Wait(ctx context.Context) (T, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	defer done()
	return l.Wait(ctx)
}

func (s *scopedQueue[T]) getOrCreate(ctx context.Context) (*queue[T], func(), error) {
	if err := s.wg.TryAdd(1); err != nil {
		return nil, nil, err
	}
	tenant := s.scope.Value(ctx)
	if tenant == "" {
		s.wg.Done()
		return nil, nil, fmt.Errorf("failed to get queue: missing tenant for scope: %s", s.scope)
	}

	q := s.newQueue(tenant)
	actual, loaded := s.queues.LoadOrStore(tenant, q)
	creCtx := contexts.WithCRE(ctx, s.scope.RoundCRE(contexts.CREValue(ctx)))
	if !loaded {
		go q.updateLoop(creCtx)
	} else {
		q = actual.(*queue[T])
		q.updateCtx(creCtx)
	}
	return q, s.wg.Done, nil
}

func (s *scopedQueue[T]) newQueue(tenant string) *queue[T] {
	q := &queue[T]{
		key:     s.key,
		scope:   s.scope,
		tenant:  tenant,
		updater: newUpdater[int](logger.With(s.lggr, s.scope.String(), tenant), s.capFn, s.subFn),
		cap:     s.defaultCap,
		recordLimit: func(ctx context.Context, c int) {
			if s.limitGauge != nil {
				s.limitGauge.Record(ctx, int64(c), withScope(ctx, s.scope))
			}
		},
		recordUsage: func(ctx context.Context, c int) {
			if s.usageGauge != nil {
				s.usageGauge.Record(ctx, int64(c), withScope(ctx, s.scope))
			}
		},
		recordDenied: func(ctx context.Context, c int) {
			if s.deniedHist != nil {
				s.deniedHist.Record(ctx, int64(c), withScope(ctx, s.scope))
			}
		},
	}
	q.cond.L = &q.mu
	q.updater.recordLimit = q.setCap
	return q
}
