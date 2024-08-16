package beholder

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelglobal "go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
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
type Client interface {
	Close() error
}

type messageEmitter struct {
	messageLogger otellog.Logger
}

type OtelClient struct {
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

// NewOtelClient creates a new Client with OTel exporter
func NewOtelClient(cfg Config, errorHandler errorHandlerFunc) (OtelClient, error) {
	factory := func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error) {
		return otlploggrpc.New(ctx, options...)
	}
	return newOtelClient(cfg, errorHandler, factory)
}

// Used for testing to override the default exporter
type otlploggrpcFactory func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error)

func newOtelClient(cfg Config, errorHandler errorHandlerFunc, otlploggrpcNew otlploggrpcFactory) (OtelClient, error) {
	ctx := context.Background()
	baseResource, err := newOtelResource(cfg)
	noop := NewNoopClient()
	if err != nil {
		return noop, err
	}
	creds := insecure.NewCredentials()
	if !cfg.InsecureConnection && cfg.CACertFile != "" {
		creds, err = credentials.NewClientTLSFromFile(cfg.CACertFile, "")
		if err != nil {
			return noop, err
		}
	}
	sharedLogExporter, err := otlploggrpcNew(
		ctx,
		otlploggrpc.WithTLSCredentials(creds),
		otlploggrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return noop, err
	}

	// Logger
	loggerProcessor := sdklog.NewBatchProcessor(
		sharedLogExporter,
		sdklog.WithExportTimeout(cfg.LogExportTimeout), // Default is 30s
	)
	loggerAttributes := []attribute.KeyValue{
		attribute.String("beholder_data_type", "zap_log_message"),
	}
	loggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(loggerAttributes...),
		baseResource,
	)
	if err != nil {
		return noop, err
	}
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(loggerResource),
		sdklog.WithProcessor(loggerProcessor),
	)
	logger := loggerProvider.Logger(defaultPackageName)

	// Tracer
	tracerProvider, err := newTracerProvider(cfg, baseResource, creds)
	if err != nil {
		return noop, err
	}
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider, err := newMeterProvider(cfg, baseResource, creds)
	if err != nil {
		return noop, err
	}
	meter := meterProvider.Meter(defaultPackageName)

	// Message Emitter
	messageLogProcessor := sdklog.NewBatchProcessor(
		sharedLogExporter,
		sdklog.WithExportTimeout(cfg.EmitterExportTimeout), // Default is 30s
	)
	messageAttributes := []attribute.KeyValue{
		attribute.String("beholder_data_type", "custom_message"),
	}
	messageLoggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(messageAttributes...),
		baseResource,
	)
	if err != nil {
		return noop, err
	}
	messageLoggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(messageLoggerResource),
		sdklog.WithProcessor(messageLogProcessor),
	)
	messageLogger := messageLoggerProvider.Logger(defaultPackageName)

	messageEmitter := &messageEmitter{
		messageLogger: messageLogger,
	}

	setOtelErrorHandler(errorHandler)

	onClose := func() (err error) {
		for _, provider := range []shutdowner{messageLoggerProvider, loggerProvider, tracerProvider, meterProvider, messageLoggerProvider} {
			err = errors.Join(err, provider.Shutdown(context.Background()))
		}
		return
	}
	client := OtelClient{cfg, logger, tracer, meter, messageEmitter, loggerProvider, tracerProvider, meterProvider, messageLoggerProvider, onClose}

	return client, nil
}

// Sets the global OTel logger, tracer, meter providers.
// Makes them accessible from anywhere in the code via global otel getters:
// - otelglobal.GetLoggerProvider()
// - otel.GetTracerProvider()
// - otel.GetTextMapPropagator()
// - otel.GetMeterProvider()
// Any package that relies on go.opentelemetry.io will be able to pick up configured global providers
// e.g [otelgrpc](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc#example-NewServerHandler)
func (c OtelClient) SetGlobals() {
	// Logger
	otelglobal.SetLoggerProvider(c.LoggerProvider)
	// Tracer
	otel.SetTracerProvider(c.TracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// Meter
	otel.SetMeterProvider(c.MeterProvider)
}

// Closes all providers, flushes all data and stops all background processes
func (c OtelClient) Close() (err error) {
	if c.OnClose != nil {
		return c.OnClose()
	}
	return
}

// Returns a new OtelClient with the same configuration but with a different package name
func (c OtelClient) ForPackage(name string) OtelClient {
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

type errorHandlerFunc func(err error)

// Sets the global error handler for OpenTelemetry
func setOtelErrorHandler(h errorHandlerFunc) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(h))
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

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			trace.WithBatchTimeout(config.TraceBatchTimeout)), // Default is 5s
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(config.TraceSampleRate),
			),
		),
	)
	return tp, nil
}

func newMeterProvider(config Config, resource *sdkresource.Resource, creds credentials.TransportCredentials) (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()

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
