package beholder

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
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

	beholderLogger "github.com/smartcontractkit/chainlink-common/pkg/beholder/logger"
)

var logger = beholderLogger.New()

type EventEmitter interface {
	// Send log record with body and attributes asynchronously, errors are handled separately by otel.ErrorHandler
	Emit(ctx context.Context, body []byte, attrs map[string]any) error
	EmitEvent(ctx context.Context, event Event) error
	// Send log record with body and attributes synchronously and returns error if any
	Send(ctx context.Context, body []byte, attrs map[string]any) error
	SendEvent(ctx context.Context, event Event) error
}
type Client interface {
	Logger() otellog.Logger
	Tracer() oteltrace.Tracer
	Meter() otelmetric.Meter
	EventEmitter() EventEmitter
	Close() error
}

var _ Client = (*BeholderClient)(nil)

type eventEmitter struct {
	exporter    sdklog.Exporter
	eventLogger otellog.Logger
	retryCount  uint
	retryDelay  time.Duration
}

type BeholderClient struct {
	config Config
	// Logger
	logger otellog.Logger
	// Tracer
	tracer oteltrace.Tracer
	// Meter
	meter otelmetric.Meter
	// EventEmitter
	eventEmitter EventEmitter

	// Graceful shutdown for tracer, meter, logger providers
	closeFunc func() error
}

func NewClient(
	config Config,
	logger otellog.Logger,
	tracer oteltrace.Tracer,
	meter otelmetric.Meter,
	eventEmitter EventEmitter,
	onClose func() error,
) *BeholderClient {
	return &BeholderClient{
		config:       config,
		logger:       logger,
		tracer:       tracer,
		meter:        meter,
		eventEmitter: eventEmitter,
		closeFunc:    onClose,
	}
}

// NewOtelClient creates a new BeholderClient with OpenTelemetry otlploggrpc Exporter

func NewOtelClient(cfg Config, errorHandler errorHandlerFunc) (*BeholderClient, error) {
	factory := func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error) {
		return otlploggrpc.New(ctx, options...)
	}
	return newOtelClient(cfg, errorHandler, factory)
}

// Used for testing to override the default exporter
type otlploggrpcFactory func(ctx context.Context, options ...otlploggrpc.Option) (sdklog.Exporter, error)

func newOtelClient(cfg Config, errorHandler errorHandlerFunc, otlploggrpcNew otlploggrpcFactory) (*BeholderClient, error) {
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

	// EventEmitter
	eventLogProcessor := sdklog.NewBatchProcessor(
		sharedLogExporter,
		sdklog.WithExportTimeout(1*time.Second), // Default is 30s
	)
	eventAttributes := []attribute.KeyValue{
		attribute.String("beholder_data_type", "custom_event"),
	}
	eventLoggerResource, err := sdkresource.Merge(
		sdkresource.NewSchemaless(eventAttributes...),
		baseResource,
	)
	if err != nil {
		return nil, err
	}
	eventLoggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(eventLoggerResource),
		sdklog.WithProcessor(eventLogProcessor),
	)
	eventLogger := eventLoggerProvider.Logger(cfg.PackageName)
	eventEmitter := newEventEmitter(sharedLogExporter, eventLogger, cfg)

	setOtelErrorHandler(errorHandler)

	onClose := closeFunc(ctx, loggerProvider, eventLoggerProvider, tracerProvider, meterProvider)

	client := NewClient(cfg, logger, tracer, meter, eventEmitter, onClose)

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
		sdkresource.WithProcess(),
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

func newEventEmitter(
	exporter sdklog.Exporter,
	eventLogger otellog.Logger,
	config Config,
) EventEmitter {
	return eventEmitter{
		exporter:    exporter,
		eventLogger: eventLogger,
		retryCount:  config.EventEmitterRetryCount,
		retryDelay:  config.EventEmitterRetryDelay,
	}
}

// Emits logs the event, but does not wait for the event to be processed.
// Open question: what are pros/cons for using use map[]any vs use otellog.KeyValue
func (e eventEmitter) Emit(ctx context.Context, body []byte, attrs map[string]any) error {
	event := NewEvent(body, attrs)
	if err := event.Validate(); err != nil {
		return err
	}
	e.eventLogger.Emit(ctx, event.OtelRecord())
	return nil
}

func (e eventEmitter) EmitEvent(ctx context.Context, event Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	e.eventLogger.Emit(ctx, event.OtelRecord())
	return nil
}

// Sends log record with body and attributes synchronously and returns error if any
func (e eventEmitter) Send(ctx context.Context, body []byte, attrs map[string]any) error {
	event := NewEvent(body, attrs)
	if err := event.Validate(); err != nil {
		return err
	}
	// NOTE: String attributes will be dropped due to a limitation in sdklog.Record
	// Will be fixed as part of INFOPLAT-811
	logger.Warn("Use Emit instead of Send. See INFOPLAT-811")
	return retry.Do(
		func() error {
			return e.exporter.Export(ctx, []sdklog.Record{event.SdkOtelRecord()})
		},
		retry.Attempts(e.retryCount),
		retry.Delay(e.retryDelay),
	)
}

// Sends log record synchronously and returns error if any
func (e eventEmitter) SendEvent(ctx context.Context, event Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	// NOTE: String attributes will be dropped due to a limitation in sdklog.Record
	// Will be fixed as part of INFOPLAT-811
	logger.Warn("Use EmitEvent instead of SendEvent. See INFOPLAT-811")
	return retry.Do(
		func() error {
			return e.exporter.Export(ctx, []sdklog.Record{event.SdkOtelRecord()})
		},
		retry.Attempts(e.retryCount),
		retry.Delay(e.retryDelay),
	)
}

func (b *BeholderClient) Logger() otellog.Logger {
	return b.logger
}

func (b *BeholderClient) Tracer() oteltrace.Tracer {
	return b.tracer
}

func (b *BeholderClient) Meter() otelmetric.Meter {
	return b.meter
}
func (b *BeholderClient) EventEmitter() EventEmitter {
	return b.eventEmitter
}

func (b *BeholderClient) Close() error {
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
