package dontime

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
)

const (
	defaultMaxPhaseOutputBytes  = 1000000 // 1 MB
	defaultMaxReportCount       = 1
	defaultBatchSize            = 10000
	defaultExecutionRemovalTime = 20 * time.Minute // 2x CRE workflow time limit
	defaultMinTimeIncrease      = time.Millisecond
)

var _ core.OCR3ReportingPluginFactory = &Factory{}

type Factory struct {
	store                   *Store
	batchSize               int
	outcomePruningThreshold uint64
	lggr                    logger.Logger

	services.StateMachine
}

func NewFactory(s *Store, lggr logger.Logger) (*Factory, error) {
	return &Factory{
		store: s,
		lggr:  logger.Named(lggr, "OCR3DonTimeFactory"),
	}, nil
}

func (o *Factory) NewReportingPlugin(_ context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	var configProto pb.Config
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
	if configProto.MaxReportCount <= 0 {
		configProto.MaxReportCount = defaultMaxReportCount
	}
	if configProto.ExecutionRemovalTime == nil {
		configProto.ExecutionRemovalTime = durationpb.New(defaultExecutionRemovalTime)
	}
	if configProto.MinTimeIncrease <= 0 {
		configProto.MinTimeIncrease = int64(defaultMinTimeIncrease)
	}

	plugin, err := NewPlugin(o.store, config, &configProto, o.lggr)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	pluginInfo := ocr3types.ReportingPluginInfo{
		Name: "DON Time Plugin",
		Limits: ocr3types.ReportingPluginLimits{
			MaxQueryLength:       int(configProto.MaxQueryLengthBytes),
			MaxObservationLength: int(configProto.MaxObservationLengthBytes),
			MaxOutcomeLength:     int(configProto.MaxOutcomeLengthBytes),
			MaxReportLength:      int(configProto.MaxReportLengthBytes),
			MaxReportCount:       int(configProto.MaxReportCount),
		},
	}
	o.lggr.Infow("DON Time Plugin created with config",
		"maxQueryLengthBytes", configProto.MaxQueryLengthBytes,
		"maxObservationLengthBytes", configProto.MaxObservationLengthBytes,
		"maxOutcomeLengthBytes", configProto.MaxOutcomeLengthBytes,
		"maxReportLengthBytes", configProto.MaxReportLengthBytes,
		"maxReportCount", configProto.MaxReportCount,
		"maxBatchSize", configProto.MaxBatchSize,
		"executionRemovalTime", configProto.ExecutionRemovalTime.AsDuration(),
		"minTimeIncrease", configProto.MinTimeIncrease,
	)
	return plugin, pluginInfo, err
}

func (o *Factory) Start(ctx context.Context) error {
	return o.StartOnce("DonTimePlugin", func() error {
		return nil
	})
}

func (o *Factory) Close() error {
	return o.StopOnce("DonTimePlugin", func() error {
		return nil
	})
}

func (o *Factory) Name() string { return o.lggr.Name() }

func (o *Factory) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}
