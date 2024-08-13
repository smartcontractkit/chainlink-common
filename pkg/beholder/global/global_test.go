package global_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	otelattribute "go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	oteltrace "go.opentelemetry.io/otel/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/global"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestGlobal(t *testing.T) {
	// Get global logger, tracer, meter, messageEmitter
	// If not initialized with global.SetClient will return noop client
	logger, tracer, meter, messageEmitter := global.Logger(), global.Tracer(), global.Meter(), global.Emitter()
	noopClient := beholder.NewNoopClient()
	assert.IsType(t, otellognoop.Logger{}, logger)
	assert.IsType(t, oteltracenoop.Tracer{}, tracer)
	assert.IsType(t, otelmetricnoop.Meter{}, meter)
	expectedMessageEmitter := beholder.NewNoopClient().Emitter
	assert.IsType(t, expectedMessageEmitter, messageEmitter)

	var noopClientPtr *beholder.OtelClient = &noopClient
	assert.IsType(t, noopClientPtr, global.GetClient())
	assert.NotSame(t, noopClientPtr, global.GetClient())

	// Set global client so it will be accessible from anywhere through beholder/global functions
	global.SetClient(noopClientPtr)
	assert.Same(t, noopClientPtr, global.GetClient())

	// After that use global functions to get logger, tracer, meter, messageEmitter
	logger, tracer, meter, messageEmitter = global.Logger(), global.Tracer(), global.Meter(), global.Emitter()

	// Emit otel log record
	logger.Emit(tests.Context(t), otellog.Record{})

	// Create trace span
	ctx, span := tracer.Start(tests.Context(t), "ExampleGlobalClient", oteltrace.WithAttributes(otelattribute.String("key", "value")))
	defer span.End()

	// Create metric counter
	counter, _ := meter.Int64Counter("global_counter")
	counter.Add(tests.Context(t), 1)

	// Emit custom message
	err := messageEmitter.Emit(ctx, []byte("test"), beholder.Attributes{"key": "value"})
	if err != nil {
		t.Fatalf("Error emitting message: %v", err)
	}
}
