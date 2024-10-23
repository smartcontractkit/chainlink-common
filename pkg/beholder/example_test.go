package beholder_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

func ExampleNewClient() {
	config := beholder.DefaultConfig()

	// Initialize beholder otel client which sets up OTel components
	client, err := beholder.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Beholder client: %v", err)
	}
	// Handle OTel errors
	otel.SetErrorHandler(otelErrPrinter)
	// Set global client so it will be accessible from anywhere through beholder functions
	beholder.SetClient(client)

	// Define a custom protobuf payload to emit
	payload := &pb.TestCustomMessage{
		BoolVal:   true,
		IntVal:    42,
		FloatVal:  3.14,
		StringVal: "Hello, World!",
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal protobuf")
	}

	// Emit the custom message anywhere from application logic
	fmt.Println("Emit custom messages")
	for range 10 {
		err := beholder.GetEmitter().Emit(context.Background(), payloadBytes,
			"beholder_data_schema", "/custom-message/versions/1", // required
			"beholder_domain", "ExampleDomain", // required
			"beholder_entity", "ExampleEntity", // required
			"beholder_data_type", "custom_message",
			"foo", "bar",
		)
		if err != nil {
			log.Printf("Error emitting message: %v", err)
		}
	}
	// Output:
	// Emit custom messages
}

func ExampleTracer() {
	ctx := context.Background()

	config := beholder.DefaultConfig()

	// Initialize beholder otel client which sets up OTel components
	client, err := beholder.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Beholder client: %v", err)
	}
	// Handle OTel errors
	otel.SetErrorHandler(otelErrPrinter)
	// Set global client so it will be accessible from anywhere through beholder functions
	beholder.SetClient(client)

	// Define a new counter
	counter, err := beholder.GetMeter().Int64Counter("custom_message.count")
	if err != nil {
		log.Fatalf("failed to create new counter")
	}

	// Define a new gauge
	gauge, err := beholder.GetMeter().Int64Gauge("custom_message.gauge")
	if err != nil {
		log.Fatalf("failed to create new gauge")
	}

	// Use the counter and gauge for metrics within application logic
	fmt.Println("Update metrics")
	counter.Add(ctx, 1)
	gauge.Record(ctx, rand.Int63n(101))

	fmt.Println("Create new trace span")
	_, rootSpan := beholder.GetTracer().Start(ctx, "foo", trace.WithAttributes(
		attribute.String("app_name", "beholderdemo"),
	))
	defer rootSpan.End()
	// Output:
	// Update metrics
	// Create new trace span
}

func ExampleNewNoopClient() {
	fmt.Println("Beholder is not initialized. Fall back to Noop OTel Client")

	fmt.Println("Emitting custom message via noop otel client")

	err := beholder.GetEmitter().Emit(context.Background(), []byte("test message"),
		"beholder_data_schema", "/custom-message/versions/1", // required
		"beholder_domain", "ExampleDomain", // required
		"beholder_entity", "ExampleEntity", // required
	)
	if err != nil {
		log.Printf("Error emitting message: %v", err)
	}
	// Output:
	// Beholder is not initialized. Fall back to Noop OTel Client
	// Emitting custom message via noop otel client
}

var otelErrPrinter = otel.ErrorHandlerFunc(func(err error) {
	log.Printf("otel error: %v", err)
})
