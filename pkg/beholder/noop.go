package beholder

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
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
	otelLogger := loggerProvider.Logger(defaultPackageName)
	// Tracer
	tracerProvider := oteltracenoop.NewTracerProvider()
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider := otelmetricnoop.NewMeterProvider()
	meter := meterProvider.Meter(defaultPackageName)

	// MessageEmitter
	messageEmitter := noopMessageEmitter{}

	// ChipIngress
	chipClient := &chipingress.NoopClient{}

	c := &Client{
		Config:                cfg,
		Logger:                otelLogger,
		Tracer:                tracer,
		Meter:                 meter,
		Emitter:               messageEmitter,
		Chip:                  chipClient,
		LoggerProvider:        loggerProvider,
		TracerProvider:        tracerProvider,
		MeterProvider:         meterProvider,
		MessageLoggerProvider: loggerProvider,
		OnClose:               noopOnClose,
	}
	c.Service, c.eng = services.Config{
		Name:  "BeholderClient",
		Start: c.start,
		Close: c.closeResources,
	}.NewServiceEngine(logger.Nop())
	return c
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
	otelLogger := loggerProvider.Logger(defaultPackageName)

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
	emitter := messageEmitter{messageLogger: otelLogger}

	onClose := func() (err error) {
		for _, provider := range []shutdowner{loggerProvider, tracerProvider, meterProvider} {
			err = errors.Join(err, provider.Shutdown(context.Background()))
		}
		return
	}

	c := &Client{
		Config:                cfg.Config,
		Logger:                otelLogger,
		Tracer:                tracer,
		Meter:                 meter,
		Emitter:               emitter,
		Chip:                  &chipingress.NoopClient{},
		LoggerProvider:        loggerProvider,
		TracerProvider:        tracerProvider,
		MeterProvider:         meterProvider,
		MessageLoggerProvider: loggerProvider,
		lazySigner:            nil,
		OnClose:               onClose,
	}
	c.Service, c.eng = services.Config{
		Name:  "BeholderClient",
		Start: c.start,
		Close: c.closeResources,
	}.NewServiceEngine(logger.Nop())
	return c, nil
}

type noopMessageEmitter struct{}

func (e noopMessageEmitter) Close() error { return nil }

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

type beholderNoopLogExporter struct{}

func (beholderNoopLogExporter) Export(ctx context.Context, records []sdklog.Record) error { return nil }
func (beholderNoopLogExporter) Shutdown(ctx context.Context) error                        { return nil }
func (beholderNoopLogExporter) ForceFlush(ctx context.Context) error                      { return nil }

// BeholderNoopLoggerProvider returns a *sdklog.LoggerProvider (the same type as sdklog.NewLoggerProvider) that drops all logs.
func BeholderNoopLoggerProvider() *sdklog.LoggerProvider {
	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(beholderNoopLogExporter{})),
	)
}
