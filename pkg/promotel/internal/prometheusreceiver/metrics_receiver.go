package prometheusreceiver

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheus/scrape"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheusreceiver/internal"
)

const (
	defaultGCInterval = 2 * time.Minute
	gcIntervalDelta   = 1 * time.Minute
)

// pReceiver is the type that provides Prometheus scraper/receiver functionality.
type pReceiver struct {
	cfg            *Config
	consumer       consumer.Metrics
	cancelFunc     context.CancelFunc
	configLoaded   chan struct{}
	loadConfigOnce sync.Once

	settings          receiver.Settings
	registerer        prometheus.Registerer
	gatherer          prometheus.Gatherer
	unregisterMetrics func()
}

func NewPrometheusReceiver(set receiver.Settings, cfg *Config, next consumer.Metrics) *pReceiver {
	return newPrometheusReceiver(set, cfg, next)
}

// New creates a new prometheus.Receiver reference.
func newPrometheusReceiver(set receiver.Settings, cfg *Config, next consumer.Metrics) *pReceiver {
	var (
		registerer prometheus.Registerer
		gatherer   prometheus.Gatherer
	)
	if cfg.Registry != nil {
		registerer = cfg.Registry
		gatherer = cfg.Registry
	} else {
		registerer = prometheus.DefaultRegisterer
		gatherer = prometheus.DefaultGatherer
	}

	pr := &pReceiver{
		cfg:          cfg,
		consumer:     next,
		settings:     set,
		configLoaded: make(chan struct{}),
		registerer: prometheus.WrapRegistererWith(
			prometheus.Labels{"receiver": set.ID.String()},
			registerer),
		gatherer: gatherer,
	}
	return pr
}

// Start is the method that starts Prometheus scraping. It
// is controlled by having previously defined a Configuration using perhaps New.
func (r *pReceiver) Start(ctx context.Context, host component.Host) error {
	discoveryCtx, cancel := context.WithCancel(context.Background())
	r.cancelFunc = cancel

	logger := internal.NewZapToGokitLogAdapter(r.settings.Logger)

	err := r.initPrometheusComponents(discoveryCtx, logger, host)
	if err != nil {
		r.settings.Logger.Error("Failed to initPrometheusComponents Prometheus components", zap.Error(err))
		return err
	}

	r.loadConfigOnce.Do(func() {
		close(r.configLoaded)
	})

	return nil
}

func (r *pReceiver) initPrometheusComponents(ctx context.Context, logger log.Logger, host component.Host) error {
	var startTimeMetricRegex *regexp.Regexp
	var err error
	if r.cfg.StartTimeMetricRegex != "" {
		startTimeMetricRegex, err = regexp.Compile(r.cfg.StartTimeMetricRegex)
		if err != nil {
			return err
		}
	}

	store, err := internal.NewAppendable(
		r.consumer,
		r.settings,
		gcInterval(r.cfg.PrometheusConfig),
		r.cfg.UseStartTimeMetric,
		startTimeMetricRegex,
		false,
		false,
		r.cfg.PrometheusConfig.GlobalConfig.ExternalLabels,
		r.cfg.TrimMetricSuffixes,
	)
	if err != nil {
		return err
	}

	loop, err := scrape.NewGathererLoop(ctx, nil, store, r.registerer, r.gatherer, 10*time.Millisecond)
	if err != nil {
		return err
	}

	r.unregisterMetrics = func() {
		loop.UnregisterMetrics()
	}

	go func() {
		<-r.configLoaded
		r.settings.Logger.Info("Starting gatherer loop")
		// Run loop directly instead of scrape manager
		loop.Run(nil)
		// if err := r.scrapeManager.Run(r.discoveryManager.SyncCh()); err != nil {
		// 	r.settings.Logger.Error("Scrape manager failed", zap.Error(err))
		// 	componentstatus.ReportStatus(host, componentstatus.NewFatalErrorEvent(err))
		// }
	}()
	return nil
}

// gcInterval returns the longest scrape interval used by a scrape config,
// plus a delta to prevent race conditions.
// This ensures jobs are not garbage collected between scrapes.
func gcInterval(cfg *PromConfig) time.Duration {
	gcInterval := defaultGCInterval
	if time.Duration(cfg.GlobalConfig.ScrapeInterval)+gcIntervalDelta > gcInterval {
		gcInterval = time.Duration(cfg.GlobalConfig.ScrapeInterval) + gcIntervalDelta
	}
	for _, scrapeConfig := range cfg.ScrapeConfigs {
		if time.Duration(scrapeConfig.ScrapeInterval)+gcIntervalDelta > gcInterval {
			gcInterval = time.Duration(scrapeConfig.ScrapeInterval) + gcIntervalDelta
		}
	}
	return gcInterval
}

// Shutdown stops and cancels the underlying Prometheus scrapers.
func (r *pReceiver) Shutdown(context.Context) error {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	if r.unregisterMetrics != nil {
		r.unregisterMetrics()
	}
	return nil
}
