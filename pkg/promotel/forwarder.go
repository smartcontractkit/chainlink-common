package promotel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

const (
	heartbeatMetricName = "promotel_heartbeat"
	hearbeatInterval    = 15 * time.Second
	scopeName           = "PromOTELForwarder"
)

type ForwarderOptions struct {
	Endpoint    string
	TLSInsecure bool
	AuthHeaders map[string]string
	Interval    time.Duration
	Verbose     bool
}

type Forwarder struct {
	lggr      logger.Logger
	heartbeat prometheus.Counter
	exporter  internal.MetricExporter
	receiver  internal.MetricReceiver
	closeOnce sync.Once
	startOnce sync.Once
	stopCh    services.StopChan
}

func NewForwarder(g prometheus.Gatherer, r prometheus.Registerer, lggr logger.Logger, opts ForwarderOptions) (*Forwarder, error) {
	exporter, err := newMetricExporter(opts, lggr)
	if err != nil {
		return nil, err
	}
	receiver, err := newMetricReceiver(g, r, opts.Interval, lggr, func(ctx context.Context, md pmetric.Metrics) error {
		if opts.Verbose {
			logOtelMetric(md, lggr)
		}
		return exporter.Export(ctx, md)
	})
	if err != nil {
		return nil, err
	}
	return &Forwarder{
		lggr: logger.Named(lggr, scopeName),
		heartbeat: promauto.With(r).NewCounter(prometheus.CounterOpts{
			Name: heartbeatMetricName,
			ConstLabels: prometheus.Labels{
				"source": scopeName,
			},
		}),
		exporter: exporter,
		receiver: receiver,
		stopCh:   make(chan struct{}),
	}, nil
}

func (f *Forwarder) Start(ctx context.Context) error {
	f.startOnce.Do(func() {
		go f.run(ctx)
	})
	return nil
}

func (f *Forwarder) run(ctx context.Context) {
	newCtx, _ := f.stopCh.Ctx(ctx)
	go f.reportHeartbeatMetric(newCtx)
	go f.startMetricExporter(newCtx)
	go f.startMetricReceiver(newCtx)
	<-newCtx.Done()
}

func (f *Forwarder) startMetricReceiver(ctx context.Context) {
	f.lggr.Debug("Starting promotel metric receiver")
	if err := f.receiver.Start(ctx); err != nil {
		f.lggr.Errorw("Failed to start promotel metric receiver, closing forwarder", "error", err)
		f.Close()
	}
	select {
	case <-ctx.Done():
		f.lggr.Debug("Context done, closing receiver")
	case <-f.stopCh:
		f.lggr.Debug("Stop channel closed, closing receiver")
	}
	if err := f.receiver.Close(); err != nil {
		f.lggr.Errorw("Failed to close receiver", "error", err)
	}
}

func (f *Forwarder) startMetricExporter(ctx context.Context) {
	f.lggr.Debug("Starting promotel metric exporter")
	if err := f.exporter.Start(ctx); err != nil {
		f.lggr.Error("Failed to start exporter, closing forwarder", err)
		f.Close()
		return
	}
	select {
	case <-ctx.Done():
		f.lggr.Debug("Context done, closing exporter")
	case <-f.stopCh:
		f.lggr.Debug("Stop channel closed, closing exporter")
	}
	if err := f.exporter.Close(); err != nil {
		f.lggr.Errorw("Failed to close exporter", "error", err)
	}
}

func (f *Forwarder) reportHeartbeatMetric(ctx context.Context) {
	ticker := time.NewTicker(hearbeatInterval)
	defer ticker.Stop()
	for {
		f.heartbeat.Inc()
		f.lggr.Debug("Heartbeat promotel")
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (f *Forwarder) Close() error {
	f.closeOnce.Do(func() {
		close(f.stopCh)
	})
	return nil
}

func DefaultForwarderOptions() ForwarderOptions {
	return ForwarderOptions{
		Endpoint:    "localhost:4317",
		TLSInsecure: true,
		AuthHeaders: nil,
		Verbose:     false,
		Interval:    15 * time.Second,
	}
}

// type ExporterConfig configgrpc.ClientConfig
func newExporterConfig(opts ForwarderOptions) (*internal.ExporterConfig, error) {
	return internal.NewMetricExporterConfig(opts.Endpoint, opts.TLSInsecure, opts.AuthHeaders)
}

func newMetricExporter(opts ForwarderOptions, lggr logger.Logger) (internal.MetricExporter, error) {
	expConfig, err := newExporterConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter config %w", err)
	}
	// Sends metrics data in OTLP format to otel-collector endpoint
	exporter, err := internal.NewMetricExporter(expConfig, lggr)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter %w", err)
	}
	return exporter, nil
}

func newMetricReceiver(g prometheus.Gatherer, r prometheus.Registerer, interval time.Duration, lggr logger.Logger, next internal.NextFunc) (internal.MetricReceiver, error) {
	receiverConfig, err := internal.NewReceiverConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create config %w", err)
	}
	receiver, err := internal.NewMetricReceiver(receiverConfig, g, r, interval, lggr, next)
	if err != nil {
		return nil, fmt.Errorf("failed to create debug metric receiver %w", err)
	}
	return receiver, nil
}

func logOtelMetric(md pmetric.Metrics, lggr logger.Logger) {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				lggr.Debugw("Exporting OTel metric ", "name", metric.Name())
			}
		}
	}
}
