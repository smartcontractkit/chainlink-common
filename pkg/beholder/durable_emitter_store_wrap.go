package beholder

import (
	"context"
	"errors"
	"time"
)

// metricsInstrumentedStore wraps DurableEventStore to record store operation metrics.
type metricsInstrumentedStore struct {
	inner DurableEventStore
	m     *durableEmitterMetrics
}

var _ DurableEventStore = (*metricsInstrumentedStore)(nil)
var _ DurableQueueObserver = (*metricsInstrumentedStore)(nil)

func newMetricsInstrumentedStore(inner DurableEventStore, m *durableEmitterMetrics) DurableEventStore {
	if m == nil {
		return inner
	}
	return &metricsInstrumentedStore{inner: inner, m: m}
}

func (s *metricsInstrumentedStore) Insert(ctx context.Context, payload []byte) (int64, error) {
	t0 := time.Now()
	id, err := s.inner.Insert(ctx, payload)
	s.m.recordStoreOp(ctx, "insert", time.Since(t0), err)
	return id, err
}

func (s *metricsInstrumentedStore) Delete(ctx context.Context, id int64) error {
	t0 := time.Now()
	err := s.inner.Delete(ctx, id)
	s.m.recordStoreOp(ctx, "delete", time.Since(t0), err)
	return err
}

func (s *metricsInstrumentedStore) MarkDelivered(ctx context.Context, id int64) error {
	t0 := time.Now()
	err := s.inner.MarkDelivered(ctx, id)
	s.m.recordStoreOp(ctx, "mark_delivered", time.Since(t0), err)
	return err
}

func (s *metricsInstrumentedStore) MarkDeliveredBatch(ctx context.Context, ids []int64) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.MarkDeliveredBatch(ctx, ids)
	s.m.recordStoreOp(ctx, "mark_delivered_batch", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) PurgeDelivered(ctx context.Context, batchLimit int) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.PurgeDelivered(ctx, batchLimit)
	s.m.recordStoreOp(ctx, "purge_delivered", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error) {
	t0 := time.Now()
	evs, err := s.inner.ListPending(ctx, createdBefore, limit)
	s.m.recordStoreOp(ctx, "list_pending", time.Since(t0), err)
	return evs, err
}

func (s *metricsInstrumentedStore) DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.DeleteExpired(ctx, ttl)
	s.m.recordStoreOp(ctx, "delete_expired", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) ObserveDurableQueue(ctx context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error) {
	o, ok := s.inner.(DurableQueueObserver)
	if !ok {
		return DurableQueueStats{}, errors.New("inner DurableEventStore does not implement DurableQueueObserver")
	}
	return o.ObserveDurableQueue(ctx, eventTTL, nearExpiryLead)
}
