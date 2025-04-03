package beholder

import (
	"context"

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

// Emits logs the message, but does not wait for the message to be processed.
// Open question: what are pros/cons for using use map[]any vs use otellog.KeyValue
func (e messageEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	message := NewMessage(body, attrKVs...)
	if err := message.Validate(); err != nil {
		return err
	}
	e.messageLogger.Emit(ctx, message.OtelRecord())
	return nil
}
