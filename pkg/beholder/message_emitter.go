package beholder

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	otellog "go.opentelemetry.io/otel/log"
)

type messageEmitter struct {
	messageLogger otellog.Logger
}

// NewMessageEmitter creates a new message emitter that emits messages to the otel collector
func NewMessageEmitter(logger otellog.Logger) Emitter {
	return messageEmitter{
		messageLogger: logger,
	}
}

func (e messageEmitter) Close() error { return nil }

// Emits logs the message, but does not wait for the message to be processed.
// Open question: what are pros/cons for using use map[]any vs use otellog.KeyValue
func (e messageEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	_, err := e.BatchEmit(ctx, []Message{
		NewMessage(body, attrKVs...),
	})
	return err
}

func (e messageEmitter) BatchEmit(ctx context.Context, messages []Message, options ...BatchEmitOption) ([]*chipingress.PublishResult, error) {
	for _, message := range messages {
		if err := message.Validate(); err != nil {
			return nil, err
		}
		e.messageLogger.Emit(ctx, message.OtelRecord())
	}
	return nil, nil
}
