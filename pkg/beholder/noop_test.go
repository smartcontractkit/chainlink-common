package beholder

import (
	"context"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestNoopOtelClient(t *testing.T) {
	noopOtelClient := NewNoopClient()
	assert.NotNil(t, noopOtelClient)

	// Message Emitter
	err := noopOtelClient.Emitter.Emit(tests.Context(t), []byte("test"),
		"key1", "value1",
	)
	assert.NoError(t, err)

	// Logger
	noopOtelClient.Logger.Emit(tests.Context(t), otellog.Record{})

	// Define a new counter
	counter, err := noopOtelClient.Meter.Int64Counter("custom_message.count")
	if err != nil {
		log.Fatalf("failed to create new counter")
	}

	// Define a new gauge
	gauge, err := noopOtelClient.Meter.Int64Gauge("custom_message.gauge")
	if err != nil {
		log.Fatalf("failed to create new gauge")
	}
	assert.NoError(t, err)

	// Use the counter and gauge for metrics within application logic
	counter.Add(tests.Context(t), 1)
	gauge.Record(tests.Context(t), rand.Int63n(101))

	// Create a new trace span
	_, rootSpan := noopOtelClient.Tracer.Start(context.Background(), "foo", trace.WithAttributes(
		attribute.String("app_name", "beholderdemo"),
	))
	rootSpan.End()

	err = noopOtelClient.Close()
	assert.NoError(t, err)
}
