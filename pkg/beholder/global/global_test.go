package global_test

import (
	"context"
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
)

func TestGlobal(t *testing.T) {
	// Get global logger, tracer, meter, eventEmitter
	// If not initialized with global.SetClient will return noop client
	logger, tracer, meter, eventEmitter := global.Logger(), global.Tracer(), global.Meter(), global.EventEmitter()
	noopClient := beholder.NewNoopClient()
	assert.IsType(t, otellognoop.Logger{}, logger)
	assert.IsType(t, oteltracenoop.Tracer{}, tracer)
	assert.IsType(t, otelmetricnoop.Meter{}, meter)
	expectedEventEmitter := beholder.NewNoopClient().EventEmitter()
	assert.IsType(t, expectedEventEmitter, eventEmitter)

	assert.IsType(t, noopClient, global.GetClient())
	assert.NotSame(t, noopClient, global.GetClient())

	// Set global client so it will be accessible from anywhere through beholder/global functions
	global.SetClient(noopClient)
	assert.Same(t, noopClient, global.GetClient())

	// After that use global functions to get logger, tracer, meter, eventEmitter
	logger, tracer, meter, eventEmitter = global.Logger(), global.Tracer(), global.Meter(), global.EventEmitter()

	// Emit otel log record
	logger.Emit(context.Background(), otellog.Record{})

	// Create trace span
	ctx, span := tracer.Start(context.Background(), "ExampleGlobalClient", oteltrace.WithAttributes(otelattribute.String("key", "value")))
	defer span.End()

	// Create metric counter
	counter, _ := meter.Int64Counter("global_counter")
	counter.Add(context.Background(), 1)

	// Emit custom event
	eventEmitter.Emit(ctx, []byte("test"), beholder.Attributes{"key": "value"})
}
