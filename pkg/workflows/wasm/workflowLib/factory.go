package workflowLib

import (
	"context"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/workflowLib/pb"
)

const (
	// TODO: What should these defaults be?
	defaultMaxPhaseOutputBytes     = 1000000 // 1 MB
	defaultMaxReportCount          = 20
	defaultBatchSize               = 20
	defaultOutcomePruningThreshold = 3600
	defaultRequestExpiry           = 10 * time.Minute // CRE workflow time limit
	defaultMinTimeIncrease         = time.Millisecond
)

type factory struct {
	store                   *DonTimeStore
	batchSize               int
	outcomePruningThreshold uint64
	lggr                    logger.Logger

	services.StateMachine
}

func newFactory(s *DonTimeStore, lggr logger.Logger) (*factory, error) {
	return &factory{
		store: s,
		lggr:  logger.Named(lggr, "OCR3WorkflowLibFactory"),
	}, nil
}

func (o *factory) NewReportingPlugin(_ context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[struct{}], ocr3types.ReportingPluginInfo, error) {
	var configProto pb.WorkflowLibConfig
	err := proto.Unmarshal(config.OffchainConfig, &configProto)
	if err != nil {
		// an empty byte array will be unmarshalled into zero values without error
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	if configProto.MaxQueryLengthBytes <= 0 {
		configProto.MaxQueryLengthBytes = defaultMaxPhaseOutputBytes
	}
	if configProto.MaxObservationLengthBytes <= 0 {
		configProto.MaxObservationLengthBytes = defaultMaxPhaseOutputBytes
	}
	if configProto.MaxOutcomeLengthBytes <= 0 {
		configProto.MaxOutcomeLengthBytes = defaultMaxPhaseOutputBytes
	}
	if configProto.MaxReportLengthBytes <= 0 {
		configProto.MaxReportLengthBytes = defaultMaxPhaseOutputBytes
	}
	if configProto.MaxBatchSize <= 0 {
		configProto.MaxBatchSize = defaultBatchSize
	}
	if configProto.MinTimeIncrease <= 0 {
		configProto.MinTimeIncrease = int64(defaultMinTimeIncrease)
	}

	plugin, err := NewWorkflowLibPlugin(o.store, config, o.lggr)
	pluginInfo := ocr3types.ReportingPluginInfo{
		Name: "OCR3 Capability Plugin",
		Limits: ocr3types.ReportingPluginLimits{
			// TODO: Do we need to change any of these default values? Sync with Bolek on this.
			MaxQueryLength:       int(configProto.MaxQueryLengthBytes),
			MaxObservationLength: int(configProto.MaxObservationLengthBytes),
			MaxOutcomeLength:     int(configProto.MaxOutcomeLengthBytes),
			MaxReportLength:      int(configProto.MaxReportLengthBytes),
			MaxReportCount:       int(configProto.MaxReportCount),
		},
	}
	return plugin, pluginInfo, err
}

func (o *factory) Start(ctx context.Context) error {
	return o.StartOnce("OCR3WorkflowLibPlugin", func() error {
		return nil
	})
}

func (o *factory) Close() error {
	return o.StopOnce("OCR3WorkflowLibPlugin", func() error {
		return nil
	})
}

func (o *factory) Name() string { return o.lggr.Name() }

func (o *factory) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}
