package beholder

import (
	"context"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Emitter interface {
	// Sends message with bytes and attributes to OTel Collector
	Emit(ctx context.Context, body []byte, attrKVs ...any) error
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

	// Providers
	LoggerProvider        otellog.LoggerProvider
	TracerProvider        oteltrace.TracerProvider
	MeterProvider         otelmetric.MeterProvider
	MessageLoggerProvider otellog.LoggerProvider

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
	opts := []otlploggrpc.Option{
		otlploggrpc.WithTLSCredentials(creds),
		otlploggrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
		otlploggrpc.WithHeaders(cfg.AuthHeaders),
	}
	if cfg.LogRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
			Enabled:         cfg.LogRetryConfig.Enabled(),
			InitialInterval: cfg.LogRetryConfig.GetInitialInterval(),
			MaxInterval:     cfg.LogRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  cfg.LogRetryConfig.GetMaxElapsedTime(),
		}))
	}
	sharedLogExporter, err := otlploggrpcNew(opts...)
	if err != nil {
		return nil, err
	}

	// Logger
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
		attribute.String("beholder_data_type", "zap_log_message"),
	}
	loggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(loggerAttributes...),
		baseResource,
	)
	if err != nil {
		return nil, err
	}
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(loggerResource),
		sdklog.WithProcessor(loggerProcessor),
	)
	logger := loggerProvider.Logger(defaultPackageName)

	// Tracer
	tracerProvider, err := newTracerProvider(cfg, baseResource, creds)
	if err != nil {
		return nil, err
	}
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider, err := newMeterProvider(cfg, baseResource, creds)
	if err != nil {
		return nil, err
	}
	meter := meterProvider.Meter(defaultPackageName)

	// Message Emitter
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
		attribute.String("beholder_data_type", "custom_message"),
	}
	messageLoggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(messageAttributes...),
		baseResource,
	)
	if err != nil {
		return nil, err
	}

	messageLoggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(messageLoggerResource),
		sdklog.WithProcessor(messageLogProcessor),
	)
	messageLogger := messageLoggerProvider.Logger(defaultPackageName)

	// Use the messageEmitter by default
	// This will eventually be removed in favor of chip-ingress emitter
	// and logs will be sent via OTLP using the regular Logger instead of calling Emit
	emitter := NewMessageEmitter(messageLogger)
	// if chip ingress is enabled, create dual source emitter that sends to both otel collector and chip ingress
	// eventually we will remove the dual source emitter and just use chip ingress
	if cfg.ChipIngressEmitterEnabled {

		chipIngressClient, err := chipingress.NewChipIngressClient(
			cfg.ChipIngressEmitterGRPCEndpoint,
			chipingress.WithTransportCredentials(insecure.NewCredentials()),
		)
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
	return &Client{cfg, logger, tracer, meter, emitter, loggerProvider, tracerProvider, meterProvider, messageLoggerProvider, onClose}, nil
}

// Closes all providers, flushes all data and stops all background processes
func (c Client) Close() (err error) {
	if c.OnClose != nil {
		return c.OnClose()
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

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

func newTracerProvider(config Config, resource *sdkresource.Resource, creds credentials.TransportCredentials) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
		otlptracegrpc.WithHeaders(config.AuthHeaders),
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

func newMeterProvider(config Config, resource *sdkresource.Resource, creds credentials.TransportCredentials) (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTLSCredentials(creds),
		otlpmetricgrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
		otlpmetricgrpc.WithHeaders(config.AuthHeaders),
	}
	if config.MetricRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         config.MetricRetryConfig.Enabled(),
			InitialInterval: config.MetricRetryConfig.GetInitialInterval(),
			MaxInterval:     config.MetricRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  config.MetricRetryConfig.GetMaxElapsedTime(),
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
				sdkmetric.WithInterval(config.MetricReaderInterval), // Default is 10s
			)),
		sdkmetric.WithResource(resource),
		sdkmetric.WithView(config.MetricViews...),
	)
	return mp, nil
}
