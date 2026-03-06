package beholder

import (
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
)

// chipIngressEmitterWorker wraps a batch.Client for a single (domain, entity) pair.
// The batch.Client handles buffering, periodic flushing, concurrent sends, and graceful shutdown.
type chipIngressEmitterWorker struct {
	batchClient *batch.Client
	metricAttrs otelmetric.MeasurementOption
}

func newChipIngressEmitterWorker(
	batchClient *batch.Client,
	domain string,
	entity string,
) *chipIngressEmitterWorker {
	return &chipIngressEmitterWorker{
		batchClient: batchClient,
		metricAttrs: otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("domain", domain),
			attribute.String("entity", entity),
		)),
	}
}
