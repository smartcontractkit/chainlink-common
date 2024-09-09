package beholder

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Default client to fallback when is is not initialized properly
func NewNoopClient() *Client {
	cfg := DefaultConfig()
	// Logger
	loggerProvider := otellognoop.NewLoggerProvider()
	logger := loggerProvider.Logger(defaultPackageName)
	// Tracer
	tracerProvider := oteltracenoop.NewTracerProvider()
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider := otelmetricnoop.NewMeterProvider()
	meter := meterProvider.Meter(defaultPackageName)

	// MessageEmitter
	messageEmitter := noopMessageEmitter{}

	return &Client{cfg, logger, tracer, meter, messageEmitter, loggerProvider, tracerProvider, meterProvider, loggerProvider, noopOnClose}
}

// NewStdoutClient creates a new Client with exporters which send telemetry data to standard output
// Used for testing and debugging
func NewStdoutClient() (*Client, error) {
	return NewWriterClient(os.Stdout)
}

// NewWriterClient creates a new Client with otel exporters which send telemetry data to custom io.Writer
func NewWriterClient(w io.Writer) (*Client, error) {
	cfg := DefaultWriterClientConfig()
	cfg.WithWriter(w)

	// Logger
	loggerExporter, err := stdoutlog.New(
		append([]stdoutlog.Option{stdoutlog.WithoutTimestamps()}, cfg.LogOptions...)...,
	)
	if err != nil {
		return NewNoopClient(), err
	}
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(loggerExporter)))
	logger := loggerProvider.Logger(defaultPackageName)

	// Tracer
	traceExporter, err := stdouttrace.New(cfg.TraceOptions...)
	if err != nil {
		return NewNoopClient(), err
	}

	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(
		sdktrace.NewSimpleSpanProcessor(traceExporter),
	))
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	metricExporter, err := stdoutmetric.New(cfg.MetricOptions...)
	if err != nil {
		return NewNoopClient(), err
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(100*time.Millisecond), // Default is 10s
			)),
	)
	meter := meterProvider.Meter(defaultPackageName)

	// MessageEmitter
	emitter := messageEmitter{messageLogger: logger}

	onClose := func() (err error) {
		for _, provider := range []shutdowner{loggerProvider, tracerProvider, meterProvider} {
			err = errors.Join(err, provider.Shutdown(context.Background()))
		}
		return
	}

	return &Client{Config: cfg.Config, Logger: logger, Tracer: tracer, Meter: meter, Emitter: emitter, LoggerProvider: loggerProvider, TracerProvider: tracerProvider, MeterProvider: meterProvider, MessageLoggerProvider: loggerProvider, OnClose: onClose}, nil
}

type noopMessageEmitter struct{}

func (noopMessageEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return nil
}
func (noopMessageEmitter) EmitMessage(ctx context.Context, message Message) error {
	return nil
}

func noopOnClose() error {
	return nil
}

type writerClientConfig struct {
	Config
	LogOptions    []stdoutlog.Option
	TraceOptions  []stdouttrace.Option
	MetricOptions []stdoutmetric.Option
}

func DefaultWriterClientConfig() writerClientConfig {
	return writerClientConfig{
		Config: DefaultConfig(),
	}
}

func (cfg *writerClientConfig) WithWriter(w io.Writer) {
	cfg.LogOptions = append(cfg.LogOptions, stdoutlog.WithWriter(w))
	cfg.TraceOptions = append(cfg.TraceOptions, stdouttrace.WithWriter(w))
	cfg.MetricOptions = append(cfg.MetricOptions, stdoutmetric.WithWriter(w))
}
