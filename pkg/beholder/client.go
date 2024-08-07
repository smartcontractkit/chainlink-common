package beholder

import (
	"context"
	"errors"
	"time"

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
)

type Emitter interface {
	// Sends message with bytes and attributes to OTel Collector
	Emit(ctx context.Context, body []byte, attrs map[string]any) error
	// Sends message to OTel Collector
	EmitMessage(ctx context.Context, m Message) error
}
type Client interface {
	Logger() otellog.Logger
	Tracer() oteltrace.Tracer
	Meter() otelmetric.Meter
	Emitter() Emitter
	Close() error
}

var _ Client = (*beholderClient)(nil)

type messageEmitter struct {
	exporter      sdklog.Exporter
	messageLogger otellog.Logger
	retryCount    uint
	retryDelay    time.Duration
}

type beholderClient struct {
	config Config
	// Logger
	logger otellog.Logger
	// Tracer
	tracer oteltrace.Tracer
	// Meter
	meter otelmetric.Meter
	// MessageEmitter
	messageEmitter Emitter
	// Graceful shutdown for tracer, meter, logger providers
	closeFunc func() error
}

func NewClient(
	config Config,
	logger otellog.Logger,
	tracer oteltrace.Tracer,
	meter otelmetric.Meter,
	messageEmitter Emitter,
	onClose func() error,
) Client {
	return &beholderClient{
		config:         config,
		logger:         logger,
		tracer:         tracer,
		meter:          meter,
		messageEmitter: messageEmitter,
		closeFunc:      onClose,
	}
}

// NewOtelClient creates a new BeholderClient with OTel exporter
func NewOtelClient(cfg Config, errorHandler errorHandlerFunc) (Client, error) {
	factory := func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error) {
		return otlploggrpc.New(ctx, options...)
	}
	return newOtelClient(cfg, errorHandler, factory)
}

// Used for testing to override the default exporter
type otlploggrpcFactory func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error)

func newOtelClient(cfg Config, errorHandler errorHandlerFunc, otlploggrpcNew otlploggrpcFactory) (Client, error) {
	ctx := context.Background()
	baseResource, err := newOtelResource()
	if err != nil {
		return nil, err
	}
	sharedLogExporter, err := otlploggrpcNew(
		ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(cfg.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}

	// Logger
	loggerProcessor := sdklog.NewBatchProcessor(
		sharedLogExporter,
		sdklog.WithExportTimeout(1*time.Second), // Default is 30s
	)
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
	logger := loggerProvider.Logger(cfg.PackageName)

	// Set global logger provider
	otelglobal.SetLoggerProvider(loggerProvider)

	// Tracer
	tracerProvider, err := newTracerProvider(cfg, baseResource)
	if err != nil {
		return nil, err
	}
	tracer := tracerProvider.Tracer(cfg.PackageName)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Meter
	meterProvider, err := newMeterProvider(cfg, baseResource)
	if err != nil {
		return nil, err
	}
	meter := meterProvider.Meter(cfg.PackageName)
	otel.SetMeterProvider(meterProvider)

	// MessageEmitter
	messageLogProcessor := sdklog.NewBatchProcessor(
		sharedLogExporter,
		sdklog.WithExportTimeout(1*time.Second), // Default is 30s
	)
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
	messageLogger := messageLoggerProvider.Logger(cfg.PackageName)
	messageEmitter := newMessageEmitter(sharedLogExporter, messageLogger, cfg)

	setOtelErrorHandler(errorHandler)

	onClose := closeFunc(ctx, loggerProvider, messageLoggerProvider, tracerProvider, meterProvider)

	client := NewClient(cfg, logger, tracer, meter, messageEmitter, onClose)

	return client, nil
}

type errorHandlerFunc func(err error)

// Sets the global error handler for OpenTelemetry
func setOtelErrorHandler(h errorHandlerFunc) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(h))
}

func newOtelResource() (resource *sdkresource.Resource, err error) {
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
	return
}

func newMessageEmitter(
	exporter sdklog.Exporter,
	messageLogger otellog.Logger,
	config Config,
) Emitter {
	return messageEmitter{
		exporter:      exporter,
		messageLogger: messageLogger,
		retryCount:    config.MessageEmitterRetryCount,
		retryDelay:    config.MessageEmitterRetryDelay,
	}
}

// Emits logs the message, but does not wait for the message to be processed.
// Open question: what are pros/cons for using use map[]any vs use otellog.KeyValue
func (e messageEmitter) Emit(ctx context.Context, body []byte, attrs map[string]any) error {
	message := NewMessage(body, attrs)
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

func (b *beholderClient) Logger() otellog.Logger {
	return b.logger
}

func (b *beholderClient) Tracer() oteltrace.Tracer {
	return b.tracer
}

func (b *beholderClient) Meter() otelmetric.Meter {
	return b.meter
}
func (b *beholderClient) Emitter() Emitter {
	return b.messageEmitter
}

func (b *beholderClient) Close() error {
	if b.closeFunc != nil {
		return b.closeFunc()
	}
	return nil
}

type otelProvider interface {
	Shutdown(ctx context.Context) error
}

// Returns function that finalizes all providers
func closeFunc(ctx context.Context, providers ...otelProvider) func() error {
	return func() (err error) {
		for _, provider := range providers {
			err = errors.Join(err, provider.Shutdown(ctx))
		}
		return
	}
}

func newTracerProvider(config Config, resource *sdkresource.Resource) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			trace.WithBatchTimeout(time.Second)), // Default is 5s
		sdktrace.WithResource(resource),
	)
	return tp, nil
}

func newMeterProvider(config Config, resource *sdkresource.Resource) (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()

	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(config.OtelExporterGRPCEndpoint),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(time.Second), // Default is 10s
			)),
		sdkmetric.WithResource(resource),
	)
	return mp, nil
}
