package monitoring

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

// MultiFeedMonitor manages the flow of data from multiple sources to
// multiple exporters for each feed in the configuration.
type MultiFeedMonitor interface {
	Run(ctx context.Context, data RDDData)
}

func NewMultiFeedMonitor(
	chainConfig ChainConfig,
	log Logger,

	sourceFactories []SourceFactory,
	exporterFactories []ExporterFactory,

	bufferCapacity uint32,
) MultiFeedMonitor {
	return &multiFeedMonitor{
		chainConfig,
		log,

		sourceFactories,
		exporterFactories,

		bufferCapacity,
	}
}

type multiFeedMonitor struct {
	chainConfig ChainConfig

	log               Logger
	sourceFactories   []SourceFactory
	exporterFactories []ExporterFactory

	bufferCapacity uint32
}

// Run should be executed as a goroutine.
func (m *multiFeedMonitor) Run(ctx context.Context, data RDDData) {
	var subs utils.Subprocesses
	defer subs.Wait()

	// setup node only monitors
	m.createMonitor(
		ctx,
		logger.With(m.log, "network", m.chainConfig.GetNetworkName()),
		&subs,
		nil,
		data.Nodes,
		true,
	)

	// setup monitors for each feed
	for _, feedConfig := range data.Feeds {
		feedLogger := logger.With(m.log,
			"feed_name", feedConfig.GetName(),
			"feed_id", feedConfig.GetID(),
			"network", m.chainConfig.GetNetworkName(),
		)
		m.createMonitor(ctx, feedLogger, &subs, feedConfig, data.Nodes, false)
	}
}

// createDataSource reusable internal method to create data sources
func (m *multiFeedMonitor) createMonitor(ctx context.Context, lgr Logger, subs *utils.Subprocesses, feedConfig FeedConfig, nodes []NodeConfig, nodeOnly bool) {
	pollers := []Poller{}
	for _, sourceFactory := range m.sourceFactories {
		if IsNodesOnly(sourceFactory.GetType()) != nodeOnly {
			// if factory is node only + running nodeOnly = keep going (true + true)
			// if factory is node only + NOT running nodeOnly = skip factory (true + false)
			// if factory is NOT node only + running nodeOnly = skip factory (false + true)
			// if factory is NOT node only + NOT running nodeOnly = keep going (false + false)
			continue // skip factory
		}
		source, err := sourceFactory.NewSource(SourceParams{
			ChainConfig: m.chainConfig,
			FeedConfig:  feedConfig,
			Nodes:       nodes,
		})
		if err != nil {
			lgr.Errorw("failed to create source", "error", err, "source-type", fmt.Sprintf("%T", sourceFactory))
			continue
		}
		poller := NewSourcePoller(
			source,
			logger.With(m.log, "component", "chain-poller", "source", sourceFactory.GetType()),
			m.chainConfig.GetPollInterval(),
			m.chainConfig.GetReadTimeout(),
			m.bufferCapacity,
		)
		pollers = append(pollers, poller)
	}
	if len(pollers) == 0 {
		lgr.Errorw("not tracking feed because all sources failed to initialize")
		return
	}
	// Create exporters
	exporters := []Exporter{}
	for _, exporterFactory := range m.exporterFactories {
		if IsNodesOnly(exporterFactory.GetType()) != nodeOnly {
			continue // see above for notes
		}

		exporter, err := exporterFactory.NewExporter(ExporterParams{
			m.chainConfig,
			feedConfig,
			nodes,
		})
		if err != nil {
			lgr.Errorw("failed to create new exporter", "error", err, "exporter-type", fmt.Sprintf("%T", exporterFactory))
			continue
		}
		exporters = append(exporters, exporter)
	}
	if len(exporters) == 0 {
		lgr.Errorw("not tracking feed because all exporters failed to initialize")
		return
	}
	// Run poller goroutines.
	for _, poller := range pollers {
		poller := poller
		subs.Go(func() {
			poller.Run(ctx)
		})
	}

	componentName := "feed"
	if nodeOnly {
		componentName = "node"
	}
	// Run feed monitor.
	feedMonitor := NewFeedMonitor(
		logger.With(lgr, "component", componentName+"-monitor"),
		pollers,
		exporters,
	)
	subs.Go(func() {
		feedMonitor.Run(ctx)
	})
}
