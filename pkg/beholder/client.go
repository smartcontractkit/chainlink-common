package beholder

import (
	"context"
	"errors"

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
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Emitter interface {
	// Sends message with bytes and attributes to OTel Collector
	Emit(ctx context.Context, body []byte, attrKVs ...any) error
}

type messageEmitter struct {
	messageLogger otellog.Logger
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
		return newHTTPClient(cfg, factory)
	}

	factory := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
		// note: context is unused internally
		return otlploggrpc.New(context.Background(), options...) //nolint
	}
	return newGRPCClient(cfg, factory)
}

// Used for testing to override the default exporter
type otlploggrpcFactory func(options ...otlploggrpc.Option) (sdklog.Exporter, error)

func newGRPCClient(cfg Config, otlploggrpcNew otlploggrpcFactory) (*Client, error) {
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
	sharedLogExporter, err := otlploggrpcNew(
		otlploggrpc.WithTLSCredentials(creds),
		otlploggrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}

	// Logger
	var loggerProcessor sdklog.Processor
	if cfg.LogBatchProcessor {
		loggerProcessor = sdklog.NewBatchProcessor(
			sharedLogExporter,
			sdklog.WithExportTimeout(cfg.LogExportTimeout), // Default is 30s
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
		messageLogProcessor = sdklog.NewBatchProcessor(
			sharedLogExporter,
			sdklog.WithExportTimeout(cfg.EmitterExportTimeout), // Default is 30s
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

	emitter := messageEmitter{
		messageLogger: messageLogger,
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
func (c Client) ForPackage(name string) Client {
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

func (e messageEmitter) EmitMessage(ctx context.Context, message Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	e.messageLogger.Emit(ctx, message.OtelRecord())
	return nil
}

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

func newTracerProvider(config Config, resource *sdkresource.Resource, creds credentials.TransportCredentials) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// note: context is used internally
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter, trace.WithBatchTimeout(config.TraceBatchTimeout)), // Default is 5s
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

	// note: context is unused internally
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithTLSCredentials(creds),
		otlpmetricgrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	)
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
	)
	return mp, nil
}
