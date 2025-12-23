package ring

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

const (
	defaultMaxPhaseOutputBytes = 1000000 // 1 MB
	defaultMaxReportCount      = 1
	defaultBatchSize           = 100
)

var _ core.OCR3ReportingPluginFactory = &Factory{}

type Factory struct {
	store  *Store
	config *ConsensusConfig
	lggr   logger.Logger

	services.StateMachine
}

// NewFactory creates a factory for the shard orchestrator consensus plugin
func NewFactory(s *Store, lggr logger.Logger, cfg *ConsensusConfig) (*Factory, error) {
	if cfg == nil {
		cfg = &ConsensusConfig{
			MinShardCount: 1,
			MaxShardCount: 10,
			BatchSize:     defaultBatchSize,
		}
	}
	return &Factory{
		store:  s,
		config: cfg,
		lggr:   logger.Named(lggr, "ShardOrchestratorFactory"),
	}, nil
}

func (o *Factory) NewReportingPlugin(_ context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	plugin, err := NewPlugin(o.store, config, o.lggr, o.config)
	pluginInfo := ocr3types.ReportingPluginInfo{
		Name: "Shard Orchestrator Consensus Plugin",
		Limits: ocr3types.ReportingPluginLimits{
			MaxQueryLength:       defaultMaxPhaseOutputBytes,
			MaxObservationLength: defaultMaxPhaseOutputBytes,
			MaxOutcomeLength:     defaultMaxPhaseOutputBytes,
			MaxReportLength:      defaultMaxPhaseOutputBytes,
			MaxReportCount:       defaultMaxReportCount,
		},
	}
	return plugin, pluginInfo, err
}

func (o *Factory) Start(ctx context.Context) error {
	return o.StartOnce("ShardOrchestratorPlugin", func() error {
		return nil
	})
}

func (o *Factory) Close() error {
	return o.StopOnce("ShardOrchestratorPlugin", func() error {
		return nil
	})
}

func (o *Factory) Name() string { return o.lggr.Name() }

func (o *Factory) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}
