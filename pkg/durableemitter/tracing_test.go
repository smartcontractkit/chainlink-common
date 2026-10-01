package durableemitter

import (
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
	assert.False(t, deliver.Parent().IsValid(), "delivery is linked, not parented")
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

func TestSetup_BuildChipOptsIncludesTracerProvider(t *testing.T) {
	tp, _ := newTestTracerProvider()
	assert.Len(t, buildChipOpts(SetupConfig{}, nil), 2) // TLS + client name
	assert.Len(t, buildChipOpts(SetupConfig{TracerProvider: tp}, nil), 3)
}
