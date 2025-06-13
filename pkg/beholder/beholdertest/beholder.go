package beholdertest

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

const (
	packageNameBeholder = "test_beholder"
)

// Observer is a test helper that provides assertion methods on received messages within the beholder client.
type Observer struct {
	emitter *assertMessageEmitter
}

// Len returns the total number of messages received that match the provided attribute key/value pairs.
func (o Observer) Len(t *testing.T, attrKVs ...any) int {
	t.Helper()

	found := o.msgsForKVs(t, attrKVs...)

	return len(found)
}

// Messages returns messages matching the provided keys and values.
func (o Observer) Messages(t *testing.T, attrKVs ...any) []beholder.Message {
	t.Helper()

	if attrKVs == nil {
		return o.emitter.msgs
	}

	return o.msgsForKVs(t, attrKVs...)
}
func (o Observer) msgsForKVs(t *testing.T, attrKVs ...any) []beholder.Message {
	t.Helper()

	o.emitter.mu.RLock()
	defer o.emitter.mu.RUnlock()

	found := []beholder.Message{}

	for _, eMsg := range o.emitter.msgs {
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

func (o Observer) BaseMessagesForLabels(t *testing.T, labels map[string]string) ([]*pb.BaseMessage, error) {
	t.Helper()

	o.emitter.mu.RLock()
	defer o.emitter.mu.RUnlock()

	var found []*pb.BaseMessage

messageLoop:
	for _, eMsg := range o.emitter.msgs {
		dataSchema, ok := eMsg.Attrs["beholder_entity"].(string)
		if !ok {
			continue
		}

		if dataSchema != "BaseMessage" {
			continue
		}

		payload := pb.BaseMessage{}
		err := proto.Unmarshal(eMsg.Body, &payload)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling base message: %v", err)
		}

		for k, v := range labels {
			if payload.Labels[k] != v {
				continue messageLoop
			}
		}

		found = append(found, &payload)
	}

	return found, nil
}

// NewObserver sets the global beholder client as a message collector and returns an Observer that provides helper assertion
// functions on received messages.
//
// NewObserver affects the whole process, it cannot be used in parallel tests or tests with parallel ancestors.
func NewObserver(t *testing.T) Observer {
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

	//reset NewObserver state after the test
	prevClient := beholder.GetClient()
	t.Cleanup(func() {
		beholder.SetClient(prevClient)
		t.Setenv(packageNameBeholder, packageNameBeholder)
	})
	beholder.SetClient(client)

	return Observer{emitter: messageEmitter}
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
