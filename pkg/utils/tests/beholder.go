package tests

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	packageNameBeholder = "test_beholder"
)

// BeholderTester is a test helper that provides assertion methods on received messages within the beholder client.
type BeholderTester struct {
	emitter *assertMessageEmitter
}

// Len returns the total number of messages received that match the provided attribute key/value pairs.
func (b BeholderTester) Len(t *testing.T, attrKVs ...any) int {
	t.Helper()

	found := b.msgsForKVs(t, attrKVs...)

	return len(found)
}

func (b BeholderTester) msgsForKVs(t *testing.T, attrKVs ...any) []beholder.Message {
	t.Helper()

	b.emitter.mu.RLock()
	defer b.emitter.mu.RUnlock()

	found := []beholder.Message{}

	for _, eMsg := range b.emitter.msgs {
		var i, j int

		for i < len(attrKVs)-1 {
			j = i + 1

			key, ok := attrKVs[i].(string)
			require.True(t, ok)

			value := attrKVs[j]
			val, ok := eMsg.Attrs[key]

			if ok && reflect.DeepEqual(value, val) {
				found = append(found, eMsg)
			}

			i = i + 2
		}
	}

	return found
}

// Beholder sets the global beholder client as a message collector and returns a tester that provides helper assertion
// functions on received messages.
func Beholder(t *testing.T) BeholderTester {
	t.Helper()

	cfg := beholder.DefaultConfig()

	// Logger
	loggerProvider := otellognoop.NewLoggerProvider()
	logger := loggerProvider.Logger(packageNameBeholder)

	// Tracer
	tracerProvider := oteltracenoop.NewTracerProvider()
	tracer := tracerProvider.Tracer(packageNameBeholder)

	// Meter
	meterProvider := otelmetricnoop.NewMeterProvider()
	meter := meterProvider.Meter(packageNameBeholder)

	// MessageEmitter
	messageEmitter := &assertMessageEmitter{t: t}

	client := &beholder.Client{
		Config:                cfg,
		Logger:                logger,
		Tracer:                tracer,
		Meter:                 meter,
		Emitter:               messageEmitter,
		LoggerProvider:        loggerProvider,
		TracerProvider:        tracerProvider,
		MeterProvider:         meterProvider,
		MessageLoggerProvider: loggerProvider,
		OnClose:               func() error { return nil },
	}

	//reset Beholder state after the test
	prevClient := beholder.GetClient()
	t.Cleanup(func() {
		beholder.SetClient(prevClient)
	})
	beholder.SetClient(client)

	return BeholderTester{emitter: messageEmitter}
}

// assertMessageEmitter is implemented with the same interface as the noopMessageEmitter in pkg/beholder/noop.go
// it is unknown at this time whether EmitMessage is needed, but it exists in the case that it is needed
type assertMessageEmitter struct {
	t    *testing.T
	mu   sync.RWMutex
	msgs []beholder.Message
}

func (e *assertMessageEmitter) Emit(_ context.Context, msg []byte, attrKVs ...any) error {
	e.t.Helper()

	e.mu.Lock()
	defer e.mu.Unlock()

	e.msgs = append(e.msgs, beholder.NewMessage(msg, attrKVs...))

	return nil
}

func (e *assertMessageEmitter) EmitMessage(_ context.Context, msg beholder.Message) error {
	e.t.Helper()

	e.mu.Lock()
	defer e.mu.Unlock()

	e.msgs = append(e.msgs, msg)

	return nil
}
