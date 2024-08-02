package beholder_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	otelattribute "go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/global"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

func ExampleClient() {
	ctx := context.Background()

	client, err := beholder.NewOtelClient(beholder.DefaultConfig(), errorHandler)
	if err != nil {
		log.Fatalf("Error creating beholder client: %v", err)
	}
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Add(1)
		fmt.Printf("Emitting message %d\n", i)
		go func(i int) {
			// Create message metadata
			metadata := beholder.Metadata{
				DonID:              "test_don_id",
				NetworkName:        []string{"test_network"},
				NetworkChainID:     "test_chain_id",
				BeholderDataSchema: "/custom-message/versions/1",
			}
			// Create custom message
			customMessage := beholder.Message{
				// Set protobuf message bytes as body
				Body: newMessageBytes(i),
				// Set metadata attributes
				Attrs: metadata.Attributes().Add(
					// Add custom attributes
					"timestamp", time.Now().Unix(),
					"sender", "example-client",
				),
			}
			// Get message emitter
			em := client.Emitter()
			// Emit custom message
			err := em.EmitMessage(ctx, customMessage)
			if err != nil {
				log.Fatalf("Error emitting message: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	// Output:
	// Emitting message 0
	// Emitting message 1
	// Emitting message 2
}

func newMessageBytes(i int) []byte {
	// Create protobuf message
	customMessagePb := &pb.CustomMessage{}
	customMessagePb.BoolVal = true
	customMessagePb.IntVal = int64(i)
	customMessagePb.FloatVal = float32(i)
	customMessagePb.StringVal = fmt.Sprintf("string-value-%d", i)
	customMessagePb.BytesVal = []byte{byte(i)}
	// Encode protobuf message
	customMessageBytes, err := proto.Marshal(customMessagePb)
	if err != nil {
		log.Fatalf("Error encoding message: %v", err)
	}
	return customMessageBytes
}

func errorHandler(e error) {
	if e != nil {
		log.Fatalf("otel error: %v", e)
	}
}

func asseetNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func ExampleEmitter() {
	ctx := context.Background()
	// Initialize beholder client
	c, err := beholder.NewOtelClient(beholder.DefaultConfig(), asseetNoError)
	if err != nil {
		log.Fatalf("Error creating beholder client: %v", err)
	}
	var client beholder.Client = c

	// Set global client so it will be accessible from anywhere through beholder/global functions
	global.SetClient(&client)
	// After that you can use global functions to get logger, tracer, meter, messageEmitter
	logger, tracer, meter, messageEmitter := global.Logger(), global.Tracer(), global.Meter(), global.Emitter()

	fmt.Println("Emit otel log record")
	logger.Emit(ctx, otellog.Record{})

	fmt.Println("Create trace span")
	ctx, span := tracer.Start(ctx, "ExampleGlobalClient", oteltrace.WithAttributes(otelattribute.String("key", "value")))
	defer span.End()

	fmt.Println("Create metric counter")
	counter, _ := meter.Int64Counter("global_counter")
	counter.Add(ctx, 1)

	fmt.Println("Emit custom message")
	err = messageEmitter.Emit(ctx, []byte("test"), beholder.Attributes{
		"key":                  "value",
		"beholder_data_schema": "/test/versions/1",
	})
	if err != nil {
		log.Fatalf("Error emitting message: %v", err)
	}
	// Output:
	// Emit otel log record
	// Create trace span
	// Create metric counter
	// Emit custom message
}

func ExampleBootstrap() {
	beholderConfig := beholder.DefaultConfig()

	// Bootstrap Beholder Client
	err := global.Bootstrap(beholderConfig, errorHandler)
	if err != nil {
		log.Fatalf("Error bootstrapping Beholder: %v", err)
	}

	payloadBytes := newMessageBytes(0)

	// Emit custom message
	for range 3 {
		err := global.Emit(context.Background(), payloadBytes, beholder.Attributes{
			"beholder_data_type": "custom_message",
			"foo":                "bar",
		})
		if err != nil {
			log.Printf("Error emitting message: %v", err)
		}
	}
	// Output:
}
