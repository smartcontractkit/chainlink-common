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

	// iterate over each feed
	for _, feedConfig := range data.Feeds {
		feedLogger := logger.With(m.log,
			"feed_name", feedConfig.GetName(),
			"feed_id", feedConfig.GetID(),
			"network", m.chainConfig.GetNetworkName(),
		)

		// create sources outside of createMonitor
		sources := []Source{}
		sourceTypes := []string{}
		for _, sourceFactory := range m.sourceFactories {
			source, err := sourceFactory.NewSource(m.chainConfig, feedConfig)
			if err != nil {
				feedLogger.Errorw("failed to create source", "error", err, "source-type", fmt.Sprintf("%T", sourceFactory))
				continue
			}
			sources = append(sources, source)
			sourceTypes = append(sourceTypes, sourceFactory.GetType())
		}

		createMonitor(
			ctx,
			&subs,
			feedLogger,
			m.chainConfig,
			m.bufferCapacity,
			sources,
			sourceTypes,
			m.exporterFactories,
			ExporterParams{
				m.chainConfig,
				feedConfig,
				data.Nodes,
			},
		)
	}
}

// createMonitor is a reusable method for creating pollers, exporters, and monitors
// sources are created outside because they may require different input parameters
func createMonitor(
	ctx context.Context,
	subs *utils.Subprocesses,
	lgr Logger,
	chainConfig ChainConfig,
	bufferCapacity uint32,
	sources []Source,
	sourceTypes []string,
	exporterFactories []ExporterFactory,
	exporterParams ExporterParams,
) {
	// Create data sources
	pollers := []Poller{}
	for i, source := range sources {
		poller := NewSourcePoller(
			source,
			logger.With(lgr, "component", "chain-poller", "source", sourceTypes[i]),
			chainConfig.GetPollInterval(),
			chainConfig.GetReadTimeout(),
			bufferCapacity,
		)
		pollers = append(pollers, poller)
	}
	if len(pollers) == 0 {
		lgr.Errorw("not tracking feed because all sources failed to initialize")
		return
	}
	// Create exporters
	exporters := []Exporter{}
	for _, exporterFactory := range exporterFactories {
		exporter, err := exporterFactory.NewExporter(exporterParams)
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
	// Run feed monitor.
	feedMonitor := NewFeedMonitor(
		logger.With(lgr, "component", "feed-monitor"),
		pollers,
		exporters,
	)
	subs.Go(func() {
		feedMonitor.Run(ctx)
	})
}
