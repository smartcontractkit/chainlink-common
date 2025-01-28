package internal

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type nonMutatingConsumer struct{}

// Capabilities returns the base consumer capabilities.
func (bc nonMutatingConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

type baseConsumer struct {
	nonMutatingConsumer
	consumer.ConsumeMetricsFunc
}

// NewNop returns a Consumer that just drops all received data and returns no error.
func NewNopConsumer() consumer.Metrics {
	return &baseConsumer{
		ConsumeMetricsFunc: func(context.Context, pmetric.Metrics) error { return nil },
	}
}

func NewConsumer(consumeFunc consumer.ConsumeMetricsFunc) consumer.Metrics {
	return &baseConsumer{
		ConsumeMetricsFunc: consumeFunc,
	}
}
