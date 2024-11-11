package beholder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Used for testing to override the default exporter
type otlploghttpFactory func(options ...otlploghttp.Option) (sdklog.Exporter, error)

func newCertFromFile(certFile string) (*x509.CertPool, error) {
	b, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, errors.New("credentials: failed to append certificates")
	}
	return cp, nil
}

func newHTTPClient(cfg Config, otlploghttpNew otlploghttpFactory) (*Client, error) {
	baseResource, err := newOtelResource(cfg)
	if err != nil {
		return nil, err
	}
	var tlsConfig *tls.Config
	if !cfg.InsecureConnection {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if cfg.CACertFile != "" {
			rootCAs, e := newCertFromFile(cfg.CACertFile)
			if e != nil {
				return nil, e
			}
			tlsConfig.RootCAs = rootCAs
		}
	}
	tlsConfigOption := otlploghttp.WithInsecure()
	if tlsConfig != nil {
		tlsConfigOption = otlploghttp.WithTLSClientConfig(tlsConfig)
	}
	opts := []otlploghttp.Option{
		tlsConfigOption,
		otlploghttp.WithEndpoint(cfg.OtelExporterHTTPEndpoint),
		otlploghttp.WithHeaders(cfg.AuthHeaders),
	}
	if cfg.LogRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlploghttp.WithRetry(otlploghttp.RetryConfig{
			Enabled:         cfg.LogRetryConfig.Enabled(),
			InitialInterval: cfg.LogRetryConfig.GetInitialInterval(),
			MaxInterval:     cfg.LogRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  cfg.LogRetryConfig.GetMaxElapsedTime(),
		}))
	}
	sharedLogExporter, err := otlploghttpNew(opts...)
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
	tracerProvider, err := newHTTPTracerProvider(cfg, baseResource, tlsConfig)
	if err != nil {
		return nil, err
	}
	tracer := tracerProvider.Tracer(defaultPackageName)

	// Meter
	meterProvider, err := newHTTPMeterProvider(cfg, baseResource, tlsConfig)
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

func newHTTPTracerProvider(config Config, resource *sdkresource.Resource, tlsConfig *tls.Config) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	tlsConfigOption := otlptracehttp.WithInsecure()
	if tlsConfig != nil {
		tlsConfigOption = otlptracehttp.WithTLSClientConfig(tlsConfig)
	}
	exporterOpts := []otlptracehttp.Option{
		tlsConfigOption,
		otlptracehttp.WithEndpoint(config.OtelExporterHTTPEndpoint),
		otlptracehttp.WithHeaders(config.AuthHeaders),
	}
	if config.TraceRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		exporterOpts = append(exporterOpts, otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         config.TraceRetryConfig.Enabled(),
			InitialInterval: config.TraceRetryConfig.GetInitialInterval(),
			MaxInterval:     config.TraceRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  config.TraceRetryConfig.GetMaxElapsedTime(),
		}))
	}
	// note: context is unused internally
	exporter, err := otlptracehttp.New(ctx, exporterOpts...)
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

func newHTTPMeterProvider(config Config, resource *sdkresource.Resource, tlsConfig *tls.Config) (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()

	tlsConfigOption := otlpmetrichttp.WithInsecure()
	if tlsConfig != nil {
		tlsConfigOption = otlpmetrichttp.WithTLSClientConfig(tlsConfig)
	}
	opts := []otlpmetrichttp.Option{
		tlsConfigOption,
		otlpmetrichttp.WithEndpoint(config.OtelExporterHTTPEndpoint),
		otlpmetrichttp.WithHeaders(config.AuthHeaders),
	}
	if config.MetricRetryConfig != nil {
		// NOTE: By default, the retry is enabled in the OTel SDK
		opts = append(opts, otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			Enabled:         config.MetricRetryConfig.Enabled(),
			InitialInterval: config.MetricRetryConfig.GetInitialInterval(),
			MaxInterval:     config.MetricRetryConfig.GetMaxInterval(),
			MaxElapsedTime:  config.MetricRetryConfig.GetMaxElapsedTime(),
		}))
	}
	// note: context is unused internally
	exporter, err := otlpmetrichttp.New(ctx, opts...)
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
