package beholder_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

type countingLogExporter struct {
	sdklog.Exporter
	shutdowns *atomic.Int64
}

func (e *countingLogExporter) Shutdown(ctx context.Context) error {
	e.shutdowns.Add(1)
	return e.Exporter.Shutdown(ctx)
}

func countingExporterFactory(created *atomic.Bool, shutdowns *atomic.Int64) func(...otlploggrpc.Option) (sdklog.Exporter, error) {
	return func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		created.Store(true)
		return &countingLogExporter{Exporter: &mockLogExporter{}, shutdowns: shutdowns}, nil
	}
}

func TestNewGRPCClient_DomainEmitterValidationBuildsNothing(t *testing.T) {
	created, shutdowns := &atomic.Bool{}, &atomic.Int64{}

	_, err := beholder.NewGRPCClient(beholder.Config{
		OtelExporterGRPCEndpoint:  "localhost:4317",
		ChipDomainEmittersEnabled: true,
	}, countingExporterFactory(created, shutdowns))
	require.ErrorContains(t, err, "ChipDomainEmittersEnabled requires chip ingress")

	cfg := domainTestConfig(t, "localhost:9999")
	cfg.ChipIngressLogger = nil
	_, err = beholder.NewGRPCClient(cfg, countingExporterFactory(created, shutdowns))
	require.ErrorContains(t, err, "ChipIngressLogger is required when ChipDomainEmittersEnabled is true")

	assert.False(t, created.Load(), "nothing is constructed for a config error")
	assert.Zero(t, shutdowns.Load())
}

func TestNewGRPCClient_ConstructionFailureShutsDownLogProviders(t *testing.T) {
	created, shutdowns := &atomic.Bool{}, &atomic.Int64{}

	cfg := domainTestConfig(t, "missing-port")
	cfg.LogStreamingEnabled = true // both log providers are real, so both shut down
	_, err := beholder.NewGRPCClient(cfg, countingExporterFactory(created, shutdowns))
	require.ErrorContains(t, err, `chip ingress client for domain "data-feeds"`)

	assert.True(t, created.Load(), "the log pipeline was built before the failure")
	assert.Equal(t, int64(2), shutdowns.Load(), "both log providers are shut down on failure")
}
