package beholder_test

import (
	"context"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestNoopClient(t *testing.T) {
	noopClient := beholder.NewNoopClient()
	assert.NotNil(t, noopClient)

	// Message Emitter
	err := noopClient.Emitter.Emit(t.Context(), []byte("test"),
		"key1", "value1",
	)
	require.NoError(t, err)

	// Logger
	noopClient.Logger.Emit(t.Context(), otellog.Record{})

	// Define a new counter
	counter, err := noopClient.Meter.Int64Counter("custom_message.count")
	if err != nil {
		log.Fatalf("failed to create new counter")
	}

	// Define a new gauge
	gauge, err := noopClient.Meter.Int64Gauge("custom_message.gauge")
	if err != nil {
		log.Fatalf("failed to create new gauge")
	}
	require.NoError(t, err)

	// Use the counter and gauge for metrics within application logic
	counter.Add(t.Context(), 1)
	gauge.Record(t.Context(), rand.Int63n(101))

	// Create a new trace span
	_, rootSpan := noopClient.Tracer.Start(context.Background(), "foo", trace.WithAttributes(
		attribute.String("app_name", "beholderdemo"),
	))
	rootSpan.End()

	err = noopClient.Close()
	assert.NoError(t, err)
}
