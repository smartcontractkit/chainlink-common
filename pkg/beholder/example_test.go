package beholder_test

import (
	"context"
	"log"
	"math/rand"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	// chainlink-common
	beholder "github.com/smartcontractkit/chainlink-common/pkg/beholder/global"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

func ExampleBeholderCustomMessage() {
	beholderConfig := beholder.NewConfig()

	// Bootstrap Beholder Client
	err := beholder.Bootstrap(beholderConfig, errorHandler)
	if err != nil {
		log.Fatalf("Error bootstrapping Beholder: %v", err)
	}

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

	// Initialise custom message to wrap the payload
	customMessage := beholder.NewMessage(payloadBytes,
		"beholder_data_schema", "/custom-message/versions/1", // required
		"beholder_data_type", "custom_message",
		"foo", "bar",
	)

	// Emit the custom message anywhere from application logic
	for range 10 {
		err := beholder.EmitMessage(context.Background(), customMessage)
		if err != nil {
			log.Printf("Error emitting message: %v", err)
		}
	}
	// Output:
}

func ExampleBeholderMetricTraces() {
	beholderConfig := beholder.NewConfig()

	// Bootstrap Beholder Client
	err := beholder.Bootstrap(beholderConfig, errorHandler)
	if err != nil {
		log.Fatalf("Error bootstrapping Beholder: %v", err)
	}

	// Define a new counter
	counter, err := beholder.Meter().Int64Counter("custom_message.count")
	if err != nil {
		log.Fatalf("failed to create new counter")
	}

	// Define a new gauge
	gauge, err := beholder.Meter().Int64Gauge("custom_message.gauge")
	if err != nil {
		log.Fatalf("failed to create new gauge")
	}

	// Use the counter and gauge for metrics within application logic
	counter.Add(context.Background(), 1)
	gauge.Record(context.Background(), rand.Int63n(101))

	// Create a new trace span
	_, rootSpan := beholder.Tracer().Start(context.Background(), "foo", trace.WithAttributes(
		attribute.String("app_name", "beholderdemo"),
	))
	defer rootSpan.End()
	// Output:
}

func errorHandler(e error) {
	if e != nil {
		log.Printf("otel error: %v", e)
	}
}
