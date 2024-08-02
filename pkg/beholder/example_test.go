package beholder_test

import (
	"context"
	"fmt"

	otelattribute "go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/global"
)

// This Testable Example demonstrates how to create a new event, add attributes to it and send it to Otel Collector.
// For more details see Testable Examples in Go https://go.dev/blog/examples
func ExampleClient() {
	// TODO(gg) fill in based on the existing example
}

func asseetNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func ExampleGlobalClient() {
	ctx := context.Background()
	// Initialize beholder client
	client, _ := beholder.NewOtelClient(beholder.DefaultBeholderConfig(), asseetNoError)
	// Set global client so it will be accessible from anywhere through beholder/global functions
	global.SetClient(client)
	// After that you can use global functions to get logger, tracer, meter, eventEmitter
	logger, tracer, meter, eventEmitter := global.Logger(), global.Tracer(), global.Meter(), global.EventEmitter()

	fmt.Println("Emit otel log record")
	logger.Emit(ctx, otellog.Record{})

	fmt.Println("Create trace span")
	ctx, span := tracer.Start(ctx, "ExampleGlobalClient", oteltrace.WithAttributes(otelattribute.String("key", "value")))
	defer span.End()

	fmt.Println("Create metric counter")
	counter, _ := meter.Int64Counter("global_counter")
	counter.Add(context.Background(), 1)

	fmt.Println("Emit custom event")
	eventEmitter.Emit(ctx, []byte("test"), beholder.Attributes{"key": "value"})

	// Output:
	// Emit otel log record
	// Create trace span
	// Create metric counter
	// Emit custom event
}
