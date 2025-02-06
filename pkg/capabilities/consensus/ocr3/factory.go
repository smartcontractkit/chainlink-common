package ocr3

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type factory struct {
	store                   *requests.Store
	capability              *capability
	batchSize               int
	outcomePruningThreshold uint64
	lggr                    logger.Logger

	services.StateMachine
}

func newFactory(s *requests.Store, c *capability, lggr logger.Logger) (*factory, error) {
	return &factory{
		store:      s,
		capability: c,
		lggr:       logger.Named(lggr, "OCR3ReportingPluginFactory"),
	}, nil
}

func (o *factory) NewReportingPlugin(_ context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	var configProto types.ReportingPluginConfig
	err := proto.Unmarshal(config.OffchainConfig, &configProto)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	rp, err := newReportingPlugin(o.store, o.capability, int(configProto.MaxBatchSize), config, configProto.OutcomePruningThreshold, o.lggr)
	rpInfo := ocr3types.ReportingPluginInfo{
		Name: "OCR3 Capability Plugin",
		Limits: ocr3types.ReportingPluginLimits{
			MaxQueryLength:       int(configProto.MaxQueryLengthBytes),
			MaxObservationLength: int(configProto.MaxObservationLengthBytes),
			MaxOutcomeLength:     int(configProto.MaxOutcomeLengthBytes),
			MaxReportLength:      int(configProto.MaxReportLengthBytes),
			MaxReportCount:       int(configProto.MaxReportCount),
		},
	}
	return rp, rpInfo, err
}

func (o *factory) Start(ctx context.Context) error {
	return o.StartOnce("OCR3ReportingPlugin", func() error {
		return nil
	})
}

func (o *factory) Close() error {
	return o.StopOnce("OCR3ReportingPlugin", func() error {
		return nil
	})
}

func (o *factory) Name() string { return o.lggr.Name() }

func (o *factory) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}
