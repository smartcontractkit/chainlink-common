package beholder

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// testChipClient is a minimal chipingress.Client for tests.
type testChipClient struct {
	chipingress.NoopClient

	mu           sync.Mutex
	publishErr   error
	batchErr     error
	publishCount atomic.Int64
	batchCount   atomic.Int64
}

func (c *testChipClient) Publish(_ context.Context, _ *chipingress.CloudEventPb, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.publishCount.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	return &chipingress.PublishResponse{}, c.publishErr
}

func (c *testChipClient) PublishBatch(_ context.Context, _ *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.batchCount.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	return &chipingress.PublishResponse{}, c.batchErr
}

func (c *testChipClient) setPublishErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishErr = err
}

func (c *testChipClient) setBatchErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batchErr = err
}

func testEmitAttrs() []any {
	return []any{"source", "test-source", "type", "test-type"}
}

func newTestDurableEmitter(t *testing.T, store DurableEventStore, client chipingress.Client, cfgOverride *DurableEmitterConfig) *DurableEmitter {
	t.Helper()
	cfg := DefaultDurableEmitterConfig()
	if cfgOverride != nil {
		cfg = *cfgOverride
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)
	return em
}

func TestDurableEmitter_HooksImmediatePath(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(1), pubCalls.Load())
	assert.Equal(t, int32(1), delCalls.Load())
}

func TestDurableEmitter_HooksPublishFailureSkipsDeleteHook(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("down"))
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return pubCalls.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(0), delCalls.Load())
}

func TestDurableEmitter_EmitPersistsAndPublishes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err)

	// Immediate async publish should fire and delete the record.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_EmitReturnSuccessEvenWhenPublishFails(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))

	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err, "Emit must succeed once the DB insert succeeds")

	// Wait for the async publish attempt to complete.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	// Event must remain in the store for retransmit.
	assert.Equal(t, 1, store.Len())
}

func TestDurableEmitter_RetransmitLoopDeliversFailedEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))
	client.setBatchErr(errors.New("connection refused"))

	cfg := DefaultDurableEmitterConfig()
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("retry-me"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	// Fix the batch client so retransmit succeeds.
	client.setBatchErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should eventually deliver and delete the event")

	assert.GreaterOrEqual(t, client.batchCount.Load(), int64(1))
}

func TestDurableEmitter_ExpiryLoopDeletesOldEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("always fail"))

	cfg := DefaultDurableEmitterConfig()
	cfg.ExpiryInterval = 100 * time.Millisecond
	cfg.EventTTL = 50 * time.Millisecond
	cfg.RetransmitInterval = 10 * time.Minute // effectively disable retransmit

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("will-expire"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "expiry loop should purge the event")
}

func TestDurableEmitter_EmitRejectsInvalidAttributes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	err := em.Emit(context.Background(), []byte("no-attrs"))
	require.Error(t, err)
	assert.Equal(t, 0, store.Len(), "nothing should be persisted when attributes are invalid")
}

func TestDurableEmitter_MultipleEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	const n = 50
	for i := 0; i < n; i++ {
		err := em.Emit(ctx, []byte("event"), testEmitAttrs()...)
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		return client.publishCount.Load() == int64(n)
	}, 5*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 10*time.Millisecond, "all events should be delivered and deleted")
}

func TestNewDurableEmitter_ValidationErrors(t *testing.T) {
	log := logger.Test(t)
	cfg := DefaultDurableEmitterConfig()

	_, err := NewDurableEmitter(nil, &testChipClient{}, cfg, log)
	assert.ErrorContains(t, err, "store")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), nil, cfg, log)
	assert.ErrorContains(t, err, "client")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), &testChipClient{}, cfg, nil)
	assert.ErrorContains(t, err, "logger")
}
