package beholder

import (
	"context"
	"errors"
	"fmt"
	"io"
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
func NewNoopClient() OtelClient {
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

	client := OtelClient{cfg, logger, tracer, meter, messageEmitter, loggerProvider, tracerProvider, meterProvider, loggerProvider, noopOnClose}

	return client
}

// NewStdoutClient creates a new Client with stdout exporters
// Use for testing and debugging
// Also this client is used as a noop client when otel exporter is not initialized properly
func NewStdoutClient(opts ...StddutClientOption) OtelClient {
	cfg := DefaultStdoutClientConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	// Logger
	loggerExporter, _ := stdoutlog.New(
		append([]stdoutlog.Option{stdoutlog.WithoutTimestamps()}, cfg.LogOptions...)...,
	) // stdoutlog.New() never returns an error
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(loggerExporter)))
	logger := loggerProvider.Logger(defaultPackageName)
	setOtelErrorHandler(func(err error) {
		fmt.Printf("OTel error %s", err)
	})

	// Tracer
	traceExporter, _ := stdouttrace.New(cfg.TraceOptions...) // stdouttrace.New() never returns an error

	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(
		sdktrace.NewSimpleSpanProcessor(traceExporter),
	))
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	metricExporter, _ := stdoutmetric.New(cfg.MetricOptions...) // stdoutmetric.New() never returns an error
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

	client := OtelClient{cfg.Config, logger, tracer, meter, emitter, loggerProvider, tracerProvider, meterProvider, loggerProvider, onClose}

	return client
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

type StddutClientOption func(*StdoutClientConfig)

type StdoutClientConfig struct {
	Config
	LogOptions    []stdoutlog.Option
	TraceOptions  []stdouttrace.Option
	MetricOptions []stdoutmetric.Option
}

func DefaultStdoutClientConfig() StdoutClientConfig {
	return StdoutClientConfig{
		Config: DefaultConfig(),
	}
}

func WithWriter(w io.Writer) StddutClientOption {
	return func(cfg *StdoutClientConfig) {
		cfg.LogOptions = append(cfg.LogOptions, stdoutlog.WithWriter(w))
		cfg.TraceOptions = append(cfg.TraceOptions, stdouttrace.WithWriter(w))
		cfg.MetricOptions = append(cfg.MetricOptions, stdoutmetric.WithWriter(w))
	}
}