package telemetry

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	prombridge "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promutil"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

func (c TelemetryConfig) tracingConfig(lggr logger.Logger, appID string) TracingConfig {
	attrs := make(map[string]string, len(c.TracingAttributes)+1)
	for k, v := range c.TracingAttributes {
		attrs[k] = v
	}
	if appID != "" {
		attrs[string(semconv.ServiceInstanceIDKey)] = appID
	}
	return TracingConfig{
		Enabled:         c.TracingEnabled,
		CollectorTarget: c.TracingCollectorTarget,
		SamplingRatio:   c.TracingSamplingRatio,
		TLSCertPath:     c.TracingTLSCertPath,
		NodeAttributes:  attrs,
		// NewSpanExporter's dialer calls this unchecked, so nil would turn an unreachable collector
		// into a nil dereference on a background goroutine.
		OnDialError: func(err error) {
			logger.Sugared(lggr).Errorw("Failed to dial the tracing collector", "err", err, "target", c.TracingCollectorTarget)
		},
	}
}

func (c TelemetryConfig) beholderConfig(lggr logger.Logger, tracing TracingConfig, otelViews []sdkmetric.View) (beholder.Config, error) {
	var attributes []attribute.KeyValue
	if tracing.Enabled {
		attributes = tracing.Attributes()
	}
	cardinalityLimit := beholder.DefaultMetricCardinalityLimit
	if c.TelemetryMetricCardinalityLimit != nil {
		cardinalityLimit = *c.TelemetryMetricCardinalityLimit
	}
	cfg := beholder.Config{
		InsecureConnection:             c.TelemetryInsecureConnection,
		CACertFile:                     c.TelemetryCACertFile,
		OtelExporterGRPCEndpoint:       c.TelemetryEndpoint,
		ResourceAttributes:             append(attributes, c.TelemetryAttributes.AsStringAttributes()...),
		TraceSampleRatio:               c.TelemetryTraceSampleRatio,
		TraceCompressor:                c.TelemetryTraceCompressor,
		EmitterBatchProcessor:          c.TelemetryEmitterBatchProcessor,
		EmitterExportTimeout:           c.TelemetryEmitterExportTimeout,
		EmitterExportInterval:          c.TelemetryEmitterExportInterval,
		EmitterExportMaxBatchSize:      c.TelemetryEmitterExportMaxBatchSize,
		EmitterMaxQueueSize:            c.TelemetryEmitterMaxQueueSize,
		LogStreamingEnabled:            c.TelemetryLogStreamingEnabled,
		LogLevel:                       c.TelemetryLogLevel,
		LogBatchProcessor:              c.TelemetryLogBatchProcessor,
		LogExportTimeout:               c.TelemetryLogExportTimeout,
		LogExportMaxBatchSize:          c.TelemetryLogExportMaxBatchSize,
		LogExportInterval:              c.TelemetryLogExportInterval,
		LogMaxQueueSize:                c.TelemetryLogMaxQueueSize,
		LogCompressor:                  c.TelemetryLogCompressor,
		ChipIngressEmitterEnabled:      c.ChipIngressEndpoint != "",
		ChipIngressEmitterGRPCEndpoint: c.ChipIngressEndpoint,
		ChipIngressInsecureConnection:  c.ChipIngressInsecureConnection,
		ChipIngressBatchEmitterEnabled: c.ChipIngressBatchEmitterEnabled,
		ChipIngressBufferSize:          c.ChipIngressBufferSize,
		ChipIngressMaxBatchSize:        c.ChipIngressMaxBatchSize,
		ChipIngressMaxConcurrentSends:  c.ChipIngressMaxConcurrentSends,
		ChipIngressSendInterval:        c.ChipIngressSendInterval,
		ChipIngressSendTimeout:         c.ChipIngressSendTimeout,
		ChipIngressDrainTimeout:        c.ChipIngressDrainTimeout,
		ChipIngressMaxGRPCRequestSize:  c.ChipIngressMaxGRPCRequestSize,
		ChipIngressLogger:              lggr,
		MetricCompressor:               c.TelemetryMetricCompressor,
		MetricExportBatchSize:          c.TelemetryMetricExportBatchSize,
		MetricCardinalityLimit:         cardinalityLimit,
		MetricViewsDenyAttributes:      c.TelemetryMetricViewsDenyAttributes,
		// The client decides between rotating and static auth by whether AuthHeadersTTL is set.
		AuthHeaders:      c.TelemetryAuthHeaders,
		AuthPublicKeyHex: c.TelemetryAuthPubKeyHex,
		AuthHeadersTTL:   c.TelemetryAuthHeadersTTL,
		// OTEL requires all histogram buckets to be defined when the client is created.
		MetricViews: otelViews,
	}

	if c.TelemetryPrometheusBridgeEnabled {
		var opts []prombridge.Option
		if len(c.TelemetryPrometheusBridgePrefixes) > 0 {
			opts = append(opts, prombridge.WithGatherer(promutil.NewPrefixGatherer(prometheus.DefaultGatherer, c.TelemetryPrometheusBridgePrefixes)))
		}
		cfg.MetricProducers = append(cfg.MetricProducers, prombridge.NewMetricProducer(opts...))
	}

	if tracing.Enabled {
		if cfg.AuthHeaders != nil {
			tracing.AuthHeaders = cfg.AuthHeaders
		}
		exporter, err := tracing.NewSpanExporter()
		if err != nil {
			return beholder.Config{}, fmt.Errorf("failed to setup tracing exporter: %w", err)
		}
		cfg.TraceSpanExporter = exporter
	}

	return cfg, nil
}

// Telemetry owns the beholder client, if one is configured, and installs it as the process's global
// client once it has started.
type Telemetry struct {
	services.Service

	client   *beholder.Client
	logLevel *zapcore.Level // set when logs are streamed to the client
}

// NewTelemetry sets up tracing and beholder from cfg. appID identifies this process in traces, and
// otelViews are added to the beholder client's metric views.
func NewTelemetry(lggr logger.Logger, cfg TelemetryConfig, appID string, otelViews []sdkmetric.View) (*Telemetry, error) {
	tracing := cfg.tracingConfig(lggr, appID)

	if cfg.TelemetryEndpoint == "" {
		if err := SetupTracing(tracing); err != nil {
			return nil, fmt.Errorf("failed to setup tracing: %w", err)
		}
		t := &Telemetry{}
		t.Service, _ = services.Config{Name: "Telemetry"}.NewServiceEngine(lggr)
		return t, nil
	}

	bcfg, err := cfg.beholderConfig(lggr, tracing, otelViews)
	if err != nil {
		return nil, err
	}

	client, err := beholder.NewClient(bcfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create beholder client: %w", err)
	}

	t := newTelemetry(lggr, client)
	if bcfg.LogStreamingEnabled {
		t.logLevel = &bcfg.LogLevel
	}
	return t, nil
}

func newTelemetry(lggr logger.Logger, client *beholder.Client) *Telemetry {
	t := &Telemetry{client: client}
	t.Service, _ = services.Config{
		Name: "Telemetry",
		NewSubServices: func(logger.Logger) []services.Service {
			return []services.Service{client}
		},
		// Sub-services start first, so the client is running by the time it's installed.
		Start: func(context.Context) error {
			beholder.SetClient(client)
			beholder.SetGlobalOtelProviders()
			return nil
		},
	}.NewServiceEngine(lggr)
	return t
}

// Client is the beholder client this owns, or nil when no endpoint was configured.
func (t *Telemetry) Client() *beholder.Client { return t.client }

// LogStreamingLevel is the level logs are streamed to Client at, and false when they aren't.
func (t *Telemetry) LogStreamingLevel() (zapcore.Level, bool) {
	if t.client == nil || t.logLevel == nil {
		return 0, false
	}
	return *t.logLevel, true
}
