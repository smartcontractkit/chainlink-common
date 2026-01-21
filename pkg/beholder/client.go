package beholder

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

const defaultGRPCCompressor = "gzip"

type Emitter interface {
	// Sends message with bytes and attributes to OTel Collector
	Emit(ctx context.Context, body []byte, attrKVs ...any) error
	io.Closer
}

type Client struct {
	Config Config
	// Logger
	Logger otellog.Logger
	// Tracer
	Tracer oteltrace.Tracer
	// Meter
	Meter otelmetric.Meter
	// Message Emitter
	Emitter Emitter
	// Chip
	Chip chipingress.Client

	// Providers
	LoggerProvider        otellog.LoggerProvider
	TracerProvider        oteltrace.TracerProvider
	MeterProvider         otelmetric.MeterProvider
	MessageLoggerProvider otellog.LoggerProvider

	// lazySigner allows updating the keystore after client initialization.
	lazySigner *lazySigner

	// OnClose
	OnClose func() error
}

// NewClient creates a new Client with initialized OpenTelemetry components
// To handle OpenTelemetry errors use [otel.SetErrorHandler](https://pkg.go.dev/go.opentelemetry.io/otel#SetErrorHandler)
func NewClient(cfg Config) (*Client, error) {
	if cfg.OtelExporterGRPCEndpoint != "" && cfg.OtelExporterHTTPEndpoint != "" {
		return nil, errors.New("only one exporter endpoint should be set")
	}
	if cfg.OtelExporterGRPCEndpoint == "" && cfg.OtelExporterHTTPEndpoint == "" {
		return nil, errors.New("at least one exporter endpoint should be set")
	}
	if cfg.OtelExporterHTTPEndpoint != "" {
		factory := func(options ...otlploghttp.Option) (sdklog.Exporter, error) {
			// note: context is unused internally
			return otlploghttp.New(context.Background(), options...) //nolint
		}
		return NewHTTPClient(cfg, factory)
	}

	factory := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
		// note: context is unused internally
		return otlploggrpc.New(context.Background(), options...) //nolint
	}

	return NewGRPCClient(cfg, factory)
}

// Used for testing to override the default exporter
type otlploggrpcFactory func(options ...otlploggrpc.Option) (sdklog.Exporter, error)

// NewGRPCClient creates a GRPC based beholder Client. Use NewClient to create a client from a Config which will pick
// the best client type from the Config.
func NewGRPCClient(cfg Config, otlploggrpcNew otlploggrpcFactory) (*Client, error) {
	baseResource, err := newOtelResource(cfg)
	if err != nil {
		return nil, err
	}
	creds := insecure.NewCredentials()
	if !cfg.InsecureConnection && cfg.CACertFile != "" {
		creds, err = credentials.NewClientTLSFromFile(cfg.CACertFile, "")
		if err != nil {
			return nil, err
		}
	}

	// Initialize auth here for reuse with log, trace, and metric exporters
	// Two modes are supported:
	// 1. Static auth: If AuthHeadersTTL == 0, use AuthHeaders as-is and never change
	// 2. Rotating auth: If AuthHeadersTTL > 0, create lazySigner for deferred keystore injection
	var signer *lazySigner
	var auth Auth

	if cfg.AuthHeadersTTL > 0 {

		if cfg.AuthPublicKeyHex == "" {
			return nil, fmt.Errorf("auth: public key hex required for rotating auth (TTL > 0)")
		}

		// Clamp lowest possible value to 10mins
		if cfg.AuthHeadersTTL < 10*time.Minute {
			return nil, fmt.Errorf("auth: headers TTL must be at least 10 minutes")
		}

		key, err := hex.DecodeString(cfg.AuthPublicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("auth: failed to decode public key hex: %w", err)
		}

		// Optionally wrap the signer in a lazySigner if AuthKeySigner was provided
		// This allows the signer to be set both before and after client initialization
		signer = &lazySigner{}
		if cfg.AuthKeySigner != nil {
			signer.Set(cfg.AuthKeySigner)
		}

		auth = NewRotatingAuth(key, signer, cfg.AuthHeadersTTL, !cfg.InsecureConnection, cfg.AuthHeaders)
	}

	// Tracer
	tracerProvider, err := newTracerProvider(cfg, baseResource, auth, creds)
	if err != nil {
		return nil, err
	}
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider, err := newMeterProvider(cfg, baseResource, auth, creds)
	if err != nil {
		return nil, err
	}
	meter := meterProvider.Meter(defaultPackageName)

	// Shared log exporter for both logger and message emitter
	logOpts, err := newLoggerOpts(cfg, auth, creds, meterProvider, tracerProvider)
	if err != nil {
		return nil, err
	}
	sharedLogExporter, err := otlploggrpcNew(logOpts...)
	if err != nil {
		return nil, err
	}

	// Logger
	var loggerProvider *sdklog.LoggerProvider
	if !cfg.LogStreamingEnabled {
		loggerProvider = BeholderNoopLoggerProvider()
	} else {
		loggerOpts, err := newLoggerProviderOpts(cfg, baseResource, sharedLogExporter)
		if err != nil {
			return nil, err
		}
		loggerProvider = sdklog.NewLoggerProvider(loggerOpts...)
	}
	logger := loggerProvider.Logger(defaultPackageName)

	// Message emitter
	messageLoggerOpts, err := newMessageLoggerProviderOpts(cfg, baseResource, sharedLogExporter)
	if err != nil {
		return nil, err
	}
	messageLoggerProvider := sdklog.NewLoggerProvider(messageLoggerOpts...)
	messageLogger := messageLoggerProvider.Logger(defaultPackageName)

	// Use the messageEmitter by default
	// This will eventually be removed in favor of chip-ingress emitter
	// and logs will be sent via OTLP using the regular Logger instead of calling Emit
	emitter := NewMessageEmitter(messageLogger)

	var chipIngressClient chipingress.Client = &chipingress.NoopClient{}
	// if chip ingress is enabled, create dual source emitter that sends to both otel collector and chip ingress
	// eventually we will remove the dual source emitter and just use chip ingress
	if cfg.ChipIngressEmitterEnabled || cfg.ChipIngressEmitterGRPCEndpoint != "" {

		var opts []chipingress.Opt

		if cfg.ChipIngressInsecureConnection {
			opts = append(opts, chipingress.WithInsecureConnection())
		} else {
			opts = append(opts, chipingress.WithTLS())
		}
		// Chip ingress auth
		switch {
		// Rotating auth
		case auth != nil:
			opts = append(opts, chipingress.WithTokenAuth(auth))
		// Static auth
		case len(cfg.AuthHeaders) > 0:
			opts = append(opts, chipingress.WithTokenAuth(
				NewStaticAuth(cfg.AuthHeaders, !cfg.ChipIngressInsecureConnection),
			))
		// No auth
		default:
		}

		// Set OpenTelemetry providers
		opts = append(opts, chipingress.WithMeterProvider(meterProvider))
		opts = append(opts, chipingress.WithTracerProvider(tracerProvider))

		chipIngressClient, err = chipingress.NewClient(cfg.ChipIngressEmitterGRPCEndpoint, opts...)
		if err != nil {
			return nil, err
		}

		chipIngressEmitter, err := NewChipIngressEmitter(chipIngressClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create chip ingress emitter: %w", err)
		}

		emitter, err = NewDualSourceEmitter(chipIngressEmitter, emitter)
		if err != nil {
			return nil, fmt.Errorf("failed to create dual source emitter: %w", err)
		}
	}

	onClose := func() (err error) {
		for _, provider := range []shutdowner{messageLoggerProvider, loggerProvider, tracerProvider, meterProvider, messageLoggerProvider} {
			err = errors.Join(err, provider.Shutdown(context.Background()))
		}
		return
	}
	return &Client{cfg, logger, tracer, meter, emitter, chipIngressClient, loggerProvider, tracerProvider, meterProvider, messageLoggerProvider, signer, onClose}, nil
}

// Closes all providers, flushes all data and stops all background processes
func (c Client) Close() (err error) {
	if c.Chip != nil {
		err = errors.Join(err, c.Chip.Close())
	}
	if c.Emitter != nil {
		err = errors.Join(err, c.Emitter.Close())
	}
	if c.OnClose != nil {
		err = errors.Join(err, c.OnClose())
	}
	return
}

// Returns a new Client with the same configuration but with a different package name
// Deprecated: Use ForName
func (c Client) ForPackage(name string) Client {
	return c.ForName(name)
}

// ForName returns a new Client with the same configuration but with a different name.
// For global package-scoped telemetry, use the package name.
// For injected component-scoped telemetry, use a fully qualified name that uniquely identifies this instance.
func (c Client) ForName(name string) Client {
	// Logger
	logger := c.LoggerProvider.Logger(name)
	// Tracer
	tracer := c.TracerProvider.Tracer(name)
	// Meter
	meter := c.MeterProvider.Meter(name)
	// Message Emitter
	messageLogger := c.MessageLoggerProvider.Logger(name)
	messageEmitter := &messageEmitter{messageLogger: messageLogger}

	newClient := c // copy
	newClient.Logger = logger
	newClient.Tracer = tracer
	newClient.Meter = meter
	newClient.Emitter = messageEmitter
	return newClient
}

// SetSigner updates the signer in the lazy signer.
// This method enables setting the signer after the beholder client has been created, which is useful
// when the signer is not available at client initialization time but the client needs to be configured
// with rotating auth. The underlying lazy signer is thread-safe.
func (c *Client) SetSigner(signer Signer) {
	if c.lazySigner != nil {
		c.lazySigner.Set(signer)
	}
}

// IsSignerSet returns true if a signer has been set in the lazy signer.
func (c *Client) IsSignerSet() bool {
	return c.lazySigner != nil && c.lazySigner.IsSet()
}

func newOtelResource(cfg Config) (resource *sdkresource.Resource, err error) {
	extraResources, err := sdkresource.New(
		context.Background(),
		sdkresource.WithOS(),
		sdkresource.WithContainer(),
		sdkresource.WithHost(),
	)
	if err != nil {
		return nil, err
	}
	resource, err = sdkresource.Merge(
		sdkresource.Default(),
		extraResources,
	)
	if err != nil {
		return nil, err
	}

	// Add csa public key resource attribute
	csaPublicKeyHex := "not-configured"
	if len(cfg.AuthPublicKeyHex) > 0 {
		csaPublicKeyHex = cfg.AuthPublicKeyHex
	}
	csaPublicKeyAttr := attribute.String("csa_public_key", csaPublicKeyHex)
	resource, err = sdkresource.Merge(
		sdkresource.NewSchemaless(csaPublicKeyAttr),
		resource,
	)
	if err != nil {
		return nil, err
	}

	// Add custom resource attributes
	resource, err = sdkresource.Merge(
		sdkresource.NewSchemaless(cfg.ResourceAttributes...),
		resource,
	)
	if err != nil {
		return nil, err
	}
	return
}

// RecordConfig records the beholder config as a metric.
func (c *Client) RecordConfigMetric(ctx context.Context) error {
	configGauge, configAttrs, err := createConfigMetric(c.Meter, c.Config)
	if err != nil {
		return err
	}
	configGauge.Record(ctx, 1, otelmetric.WithAttributes(configAttrs...))
	return nil
}

// createConfigMetric creates a configuration info metric with Beholder settings as attributes.
func createConfigMetric(meter otelmetric.Meter, cfg Config) (otelmetric.Int64Gauge, []attribute.KeyValue, error) {
	configGauge, err := meter.Int64Gauge(
		"beholder.config.info",
		otelmetric.WithDescription("Beholder config info metric"),
		otelmetric.WithUnit("{info}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create beholder config info metric: %w", err)
	}

	configAttrs := []attribute.KeyValue{
		// Logging config
		attribute.Bool(
			"log_streaming_enabled", cfg.LogStreamingEnabled),
		attribute.String(
			"log_level", cfg.LogLevel.String()),
		attribute.Bool(
			"log_batch_processor", cfg.LogBatchProcessor),
		attribute.String(
			"log_export_interval", cfg.LogExportInterval.String()),
		attribute.Int(
			"log_export_max_batch_size", cfg.LogExportMaxBatchSize),
		attribute.Int(
			"log_max_queue_size", cfg.LogMaxQueueSize),
		attribute.String(
			"log_compressor", cfg.LogCompressor),

		// Message emitter config
		attribute.Bool(
			"chip_ingress_enabled", cfg.ChipIngressEmitterEnabled),
		attribute.Bool(
			"emitter_batch_processor", cfg.EmitterBatchProcessor),
		attribute.String(
			"emitter_export_interval", cfg.EmitterExportInterval.String()),
		attribute.Int(
			"emitter_export_max_batch_size", cfg.EmitterExportMaxBatchSize),
		attribute.Int(
			"emitter_max_queue_size", cfg.EmitterMaxQueueSize),

		// Tracing config
		attribute.Float64(
			"trace_sample_ratio", cfg.TraceSampleRatio),
		attribute.String(
			"trace_batch_timeout", cfg.TraceBatchTimeout.String()),
		attribute.String(
			"trace_compressor", cfg.TraceCompressor),

		// Metrics config
		attribute.String(
			"metric_reader_interval", cfg.MetricReaderInterval.String()),
		attribute.String(
			"metric_compressor", cfg.MetricCompressor),
	}

	return configGauge, configAttrs, nil
}

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

func newTracerProvider(config Config, resource *sdkresource.Resource, auth Auth, creds credentials.TransportCredentials) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	}
	switch {
	// Rotating auth
	case auth != nil:
		exporterOpts = append(exporterOpts, otlptracegrpc.WithDialOption(authDialOpt(auth)))
	// Static auth
	case len(config.AuthHeaders) > 0:
		exporterOpts = append(exporterOpts, otlptracegrpc.WithHeaders(config.AuthHeaders))
	// No auth
	default:
	}
	switch compressor := config.TraceCompressor; compressor {
	case "none":
	case "":
		exporterOpts = append(exporterOpts, otlptracegrpc.WithCompressor(defaultGRPCCompressor))
	default:
		exporterOpts = append(exporterOpts, otlptracegrpc.WithCompressor(compressor))
	}
	if config.TraceRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		exporterOpts = append(exporterOpts, otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         config.TraceRetryConfig.Enabled(),
			InitialInterval: config.TraceRetryConfig.GetInitialInterval(),
			MaxInterval:     config.TraceRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  config.TraceRetryConfig.GetMaxElapsedTime(),
		}))
	}
	// note: context is used internally
	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, err
	}
	batcherOpts := []sdktrace.BatchSpanProcessorOption{}
	if config.TraceBatchTimeout > 0 {
		batcherOpts = append(batcherOpts, sdktrace.WithBatchTimeout(config.TraceBatchTimeout)) // Default is 5s
	}
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter, batcherOpts...),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(config.TraceSampleRatio),
			),
		),
	}
	if config.TraceSpanExporter != nil {
		opts = append(opts, sdktrace.WithBatcher(config.TraceSpanExporter))
	}
	return sdktrace.NewTracerProvider(opts...), nil
}

func newMeterProvider(cfg Config, resource *sdkresource.Resource, auth Auth, creds credentials.TransportCredentials) (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTLSCredentials(creds),
		otlpmetricgrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
	}
	switch compressor := cfg.MetricCompressor; compressor {
	case "none":
	case "":
		opts = append(opts, otlpmetricgrpc.WithCompressor(defaultGRPCCompressor))
	default:
		opts = append(opts, otlpmetricgrpc.WithCompressor(compressor))
	}

	switch {
	// Rotating auth
	case auth != nil:
		opts = append(opts, otlpmetricgrpc.WithDialOption(authDialOpt(auth)))
	// Static auth
	case len(cfg.AuthHeaders) > 0:
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.AuthHeaders))
	// No auth
	default:
	}

	if cfg.MetricRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         cfg.MetricRetryConfig.Enabled(),
			InitialInterval: cfg.MetricRetryConfig.GetInitialInterval(),
			MaxInterval:     cfg.MetricRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  cfg.MetricRetryConfig.GetMaxElapsedTime(),
		}))
	}
	// note: context is unused internally
	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(cfg.MetricReaderInterval), // Default is 10s
			)),
		sdkmetric.WithResource(resource),
		sdkmetric.WithView(cfg.MetricViews...),
	)
	return mp, nil
}

// newLoggerOpts creates options for a logger exporter
func newLoggerOpts(cfg Config, auth Auth, creds credentials.TransportCredentials, meter *sdkmetric.MeterProvider, tracer *sdktrace.TracerProvider) ([]otlploggrpc.Option, error) {
	otelOpts := []otelgrpc.Option{
		otelgrpc.WithMeterProvider(meter),
		otelgrpc.WithTracerProvider(tracer),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelOpts...)),
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithTLSCredentials(creds),
		otlploggrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
	}
	switch compressor := cfg.LogCompressor; compressor {
	case "none":
	case "":
		opts = append(opts, otlploggrpc.WithCompressor(defaultGRPCCompressor))
	default:
		opts = append(opts, otlploggrpc.WithCompressor(compressor))
	}
	// Log exporter auth
	switch {
	// Rotating auth
	case auth != nil:
		dialOpts = append(dialOpts, authDialOpt(auth))
	// Static auth
	case len(cfg.AuthHeaders) > 0:
		opts = append(opts, otlploggrpc.WithHeaders(cfg.AuthHeaders))
	// No auth
	default:
	}

	opts = append(opts, otlploggrpc.WithDialOption(dialOpts...))

	if cfg.LogRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
			Enabled:         cfg.LogRetryConfig.Enabled(),
			InitialInterval: cfg.LogRetryConfig.GetInitialInterval(),
			MaxInterval:     cfg.LogRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  cfg.LogRetryConfig.GetMaxElapsedTime(),
		}))
	}
	return opts, nil
}

// newLoggerProviderOpts creates logger provider options for application logs
func newLoggerProviderOpts(cfg Config, baseResource *sdkresource.Resource, sharedLogExporter sdklog.Exporter) ([]sdklog.LoggerProviderOption, error) {
	var loggerProcessor sdklog.Processor
	if cfg.LogBatchProcessor {
		batchProcessorOpts := []sdklog.BatchProcessorOption{}
		if cfg.LogExportTimeout > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportTimeout(cfg.LogExportTimeout)) // Default is 30s
		}
		if cfg.LogExportMaxBatchSize > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportMaxBatchSize(cfg.LogExportMaxBatchSize)) // Default is 512, must be <= maxQueueSize
		}
		if cfg.LogExportInterval > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportInterval(cfg.LogExportInterval)) // Default is 1s
		}
		if cfg.LogMaxQueueSize > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithMaxQueueSize(cfg.LogMaxQueueSize)) // Default is 2048
		}
		loggerProcessor = sdklog.NewBatchProcessor(
			sharedLogExporter,
			batchProcessorOpts...,
		)
	} else {
		loggerProcessor = sdklog.NewSimpleProcessor(sharedLogExporter)
	}

	loggerAttributes := []attribute.KeyValue{
		attribute.String(AttrKeyDataType, "zap_log_message"),
	}
	loggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(loggerAttributes...),
		baseResource,
	)
	if err != nil {
		return nil, err
	}

	return []sdklog.LoggerProviderOption{
		sdklog.WithResource(loggerResource),
		sdklog.WithProcessor(loggerProcessor),
	}, nil
}

// newMessageLoggerProviderOpts creates logger provider options for custom message emitter
func newMessageLoggerProviderOpts(cfg Config, baseResource *sdkresource.Resource, sharedLogExporter sdklog.Exporter) ([]sdklog.LoggerProviderOption, error) {
	var messageLogProcessor sdklog.Processor
	if cfg.EmitterBatchProcessor {
		batchProcessorOpts := []sdklog.BatchProcessorOption{}
		if cfg.EmitterExportTimeout > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportTimeout(cfg.EmitterExportTimeout)) // Default is 30s
		}
		if cfg.EmitterExportMaxBatchSize > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportMaxBatchSize(cfg.EmitterExportMaxBatchSize)) // Default is 512, must be <= maxQueueSize
		}
		if cfg.EmitterExportInterval > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithExportInterval(cfg.EmitterExportInterval)) // Default is 1s
		}
		if cfg.EmitterMaxQueueSize > 0 {
			batchProcessorOpts = append(batchProcessorOpts, sdklog.WithMaxQueueSize(cfg.EmitterMaxQueueSize)) // Default is 2048
		}
		messageLogProcessor = sdklog.NewBatchProcessor(
			sharedLogExporter,
			batchProcessorOpts...,
		)
	} else {
		messageLogProcessor = sdklog.NewSimpleProcessor(sharedLogExporter)
	}

	messageAttributes := []attribute.KeyValue{
		attribute.String(AttrKeyDataType, "custom_message"),
	}
	messageLoggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(messageAttributes...),
		baseResource,
	)
	if err != nil {
		return nil, err
	}

	return []sdklog.LoggerProviderOption{
		sdklog.WithResource(messageLoggerResource),
		sdklog.WithProcessor(messageLogProcessor),
	}, nil
}
