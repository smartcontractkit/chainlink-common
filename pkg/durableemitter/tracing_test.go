package durableemitter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func newTestTracerProvider() (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	rec := tracetest.NewSpanRecorder()
	return sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)), rec
}

func endedSpan(rec *tracetest.SpanRecorder, name string) sdktrace.ReadOnlySpan {
	for _, s := range rec.Ended() {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func TestDurableEmitter_TracingEmitAndLinkedDeliver(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return endedSpan(rec, spanDeliver) != nil }, 2*time.Second, 5*time.Millisecond)

	emit := endedSpan(rec, spanEmit)
	require.NotNil(t, emit)
	assert.Equal(t, otelcodes.Unset, emit.Status().Code)

	deliver := endedSpan(rec, spanDeliver)
	require.Len(t, deliver.Links(), 1)
	assert.Equal(t, emit.SpanContext().SpanID(), deliver.Links()[0].SpanContext.SpanID())
	assert.Equal(t, emit.SpanContext().SpanID(), deliver.Parent().SpanID(), "emit is the remote parent so sampling is consistent")
	assert.True(t, deliver.Parent().IsRemote())
	assert.Equal(t, emit.SpanContext().TraceID(), deliver.SpanContext().TraceID())
}

func TestDurableEmitter_TracingDeliverFailureRecorded(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp

	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))
	em := newTestDurableEmitter(t, NewMemDurableEventStore(), be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return endedSpan(rec, spanDeliver) != nil }, 2*time.Second, 5*time.Millisecond)
	assert.Equal(t, otelcodes.Error, endedSpan(rec, spanDeliver).Status().Code)
}

func TestDurableEmitter_TracingEmitErrorSetsStatus(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp

	em := newTestDurableEmitter(t, NewMemDurableEventStore(), newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)

	require.Error(t, em.Emit(t.Context(), []byte("hello")))
	emit := endedSpan(rec, spanEmit)
	require.NotNil(t, emit)
	assert.Equal(t, otelcodes.Error, emit.Status().Code)
}

func TestDurableEmitter_TracingRetransmitTick(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.RetransmitAfter = 10 * time.Millisecond

	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))
	em := newTestDurableEmitter(t, NewMemDurableEventStore(), be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return endedSpan(rec, spanRetransmitTick) != nil }, 3*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_TracingRetransmitDeliverLinksToTick(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.RetransmitAfter = 10 * time.Millisecond

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("hello"), testEmitAttrs()...))
	retransmitDeliver := func() sdktrace.ReadOnlySpan {
		for _, s := range rec.Ended() {
			if s.Name() != spanDeliver {
				continue
			}
			for _, a := range s.Attributes() {
				if a.Key == "phase" && a.Value.AsString() == publishPhaseRetransmit.String() {
					return s
				}
			}
		}
		return nil
	}
	require.Eventually(t, func() bool { return retransmitDeliver() != nil }, 3*time.Second, 10*time.Millisecond)
	retransmit := retransmitDeliver()
	require.NotNil(t, retransmit, "retransmit-phase deliver span")
	require.Len(t, retransmit.Links(), 1)
	assert.Equal(t, retransmit.Links()[0].SpanContext.SpanID(), retransmit.Parent().SpanID())
	tick := endedSpan(rec, spanRetransmitTick)
	require.NotNil(t, tick)
}

func TestDurableEmitter_TracingExpiryTick(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp
	cfg.ExpiryInterval = 50 * time.Millisecond
	cfg.EventTTL = 10 * time.Millisecond
	cfg.RetransmitInterval = 10 * time.Minute

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("always fail"))
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("will-expire"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return endedSpan(rec, spanExpiryTick) != nil }, 3*time.Second, 10*time.Millisecond)
	assert.Equal(t, otelcodes.Unset, endedSpan(rec, spanExpiryTick).Status().Code)
}

type failingStore struct {
	*MemDurableEventStore
	listErr   error
	expireErr error
}

func (s *failingStore) ListPending(ctx context.Context, createdBefore, afterCreatedAt time.Time, afterID int64, limit int) ([]DurableEvent, error) {
	return nil, s.listErr
}

func (s *failingStore) DeleteExpiredBatch(context.Context, time.Duration, int) ([][]byte, error) {
	return nil, s.expireErr
}

func TestDurableEmitter_TracingBackgroundErrorsSetStatus(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.ExpiryInterval = 50 * time.Millisecond

	store := &failingStore{MemDurableEventStore: NewMemDurableEventStore(), listErr: errors.New("list failed"), expireErr: errors.New("expire failed")}
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)

	require.Eventually(t, func() bool {
		return endedSpan(rec, spanRetransmitTick) != nil && endedSpan(rec, spanExpiryTick) != nil
	}, 3*time.Second, 10*time.Millisecond)
	assert.Equal(t, otelcodes.Error, endedSpan(rec, spanRetransmitTick).Status().Code)
	assert.Equal(t, otelcodes.Error, endedSpan(rec, spanExpiryTick).Status().Code)
}

func TestDurableEmitter_TracingIdleRetransmitTicksAreNotTraced(t *testing.T) {
	tp, rec := newTestTracerProvider()
	cfg := DefaultConfig()
	cfg.TracerProvider = tp
	cfg.RetransmitInterval = 10 * time.Millisecond

	em := newTestDurableEmitter(t, NewMemDurableEventStore(), newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)

	time.Sleep(200 * time.Millisecond) // ~20 retransmit ticks with nothing pending
	assert.Nil(t, endedSpan(rec, spanRetransmitTick))
}

func TestDurableEmitter_NilTracerProviderUsesGlobal(t *testing.T) {
	em := newTestDurableEmitter(t, NewMemDurableEventStore(), newTestBatchEmitter(), nil)
	require.NotNil(t, em.tracer)
	servicetest.Run(t, em)
	require.NoError(t, em.Emit(t.Context(), []byte("hello"), testEmitAttrs()...))
}

func TestSetup_BuildChipOptsTracerProvider(t *testing.T) {
	tp, _ := newTestTracerProvider()
	without := len(buildChipOpts(SetupConfig{}, nil))
	with := len(buildChipOpts(SetupConfig{TracerProvider: tp}, nil))
	assert.Equal(t, without+1, with, "a tracer provider adds exactly one chipingress option")
}
