package ocr3

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3_1types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// OCR3_1 observation/report bytes defaults. Held below the libocr hard caps:
//   MaxMaxObservationBytes = 512 KiB   (halved vs OCR3)
//   MaxMaxQueryBytes       = 512 KiB
//   MaxMaxReportBytes      = 5   MiB
// Any DON offchain config that exceeds these will fail ReportingPluginInfo
// validation at factory time — preflight rotation (plan §3.7) is mandatory.
const (
	defaultMaxObservationBytesOCR3_1          = 400 * 1024 // 400 KiB (~80% of the 512 KiB cap)
	defaultMaxQueryBytesOCR3_1                = 400 * 1024
	defaultMaxReportsPlusPrecursorBytesOCR3_1 = 1 * 1024 * 1024 // 1 MiB — small, precursor only
	defaultMaxReportBytesOCR3_1               = 1 * 1024 * 1024
	defaultMaxReportCountOCR3_1               = 20

	// KV write budget. Bounded by batch size × AggregationOutcome size.
	// Well below the libocr caps (10_000 keys / 10 MiB).
	defaultMaxKeyValueModifiedKeysOCR3_1                = 1024
	defaultMaxKeyValueModifiedKeysPlusValuesBytesOCR3_1 = 4 * 1024 * 1024

	// Blob limits. v1 uses blobs for observation payloads only.
	defaultMaxBlobPayloadBytesOCR3_1                          = 1 * 1024 * 1024 // 1 MiB per blob
	defaultMaxPerOracleUnexpiredBlobCountOCR3_1               = 500
	defaultMaxPerOracleUnexpiredBlobCumulativePayloadBytesOCR3_1 = 500 * 1024 * 1024
)

type factoryOCR3_1 struct {
	store      *requests.Store[*ReportRequest]
	capability *capability
	lggr       logger.Logger

	services.StateMachine
}

func newFactoryOCR3_1(
	s *requests.Store[*ReportRequest],
	c *capability,
	lggr logger.Logger,
) (*factoryOCR3_1, error) {
	return &factoryOCR3_1{
		store:      s,
		capability: c,
		lggr:       logger.Named(lggr, "OCR3_1ReportingPluginFactory"),
	}, nil
}

// NewReportingPlugin implements ocr3_1types.ReportingPluginFactory[[]byte].
// The BlobBroadcastFetcher must not be captured long-term; libocr only
// guarantees it within method scopes (see ocr3_1types/plugin.go doc). We
// deliberately do not stash it on the factory — each method on the plugin
// receives it fresh.
func (o *factoryOCR3_1) NewReportingPlugin(
	_ context.Context,
	config ocr3types.ReportingPluginConfig,
	_ ocr3_1types.BlobBroadcastFetcher,
) (ocr3_1types.ReportingPlugin[[]byte], ocr3_1types.ReportingPluginInfo, error) {
	var configProto types.ReportingPluginConfig
	if err := proto.Unmarshal(config.OffchainConfig, &configProto); err != nil {
		return nil, ocr3_1types.ReportingPluginInfo1{}, err
	}

	// Defaults: OCR3_1 caps are tighter than OCR3, so we cannot inherit the
	// OCR3 1 MiB defaults. Any value the operator supplied is kept; zero
	// values are filled with OCR3_1-safe defaults.
	if configProto.MaxQueryLengthBytes <= 0 {
		configProto.MaxQueryLengthBytes = defaultMaxQueryBytesOCR3_1
	}
	if configProto.MaxObservationLengthBytes <= 0 {
		configProto.MaxObservationLengthBytes = defaultMaxObservationBytesOCR3_1
	}
	if configProto.MaxOutcomeLengthBytes <= 0 {
		configProto.MaxOutcomeLengthBytes = defaultMaxReportsPlusPrecursorBytesOCR3_1
	}
	if configProto.MaxReportLengthBytes <= 0 {
		configProto.MaxReportLengthBytes = defaultMaxReportBytesOCR3_1
	}
	if configProto.MaxReportCount <= 0 {
		configProto.MaxReportCount = defaultMaxReportCountOCR3_1
	}
	if configProto.OutcomePruningThreshold <= 0 {
		configProto.OutcomePruningThreshold = defaultOutcomePruningThreshold
	}
	if configProto.RequestTimeout == nil {
		configProto.RequestTimeout = durationpb.New(defaultRequestExpiry)
	}
	// OCR3_1-only fields: honor operator-supplied values; fall back to
	// defaults when unset. Keeps OCR3 offchain configs forward-compatible.
	if configProto.MaxReportsPlusPrecursorBytes == 0 {
		configProto.MaxReportsPlusPrecursorBytes = defaultMaxReportsPlusPrecursorBytesOCR3_1
	}
	if configProto.MaxKeyValueModifiedKeysPlusValuesBytes == 0 {
		configProto.MaxKeyValueModifiedKeysPlusValuesBytes = defaultMaxKeyValueModifiedKeysPlusValuesBytesOCR3_1
	}
	if configProto.MaxBlobPayloadBytes == 0 {
		configProto.MaxBlobPayloadBytes = defaultMaxBlobPayloadBytesOCR3_1
	}
	if configProto.BlobExpirationK == 0 {
		configProto.BlobExpirationK = defaultBlobExpirationK
	}
	if configProto.MaxKeyValueModifiedKeys == 0 {
		configProto.MaxKeyValueModifiedKeys = defaultMaxKeyValueModifiedKeysOCR3_1
	}
	if configProto.MaxPerOracleUnexpiredBlobCount == 0 {
		configProto.MaxPerOracleUnexpiredBlobCount = defaultMaxPerOracleUnexpiredBlobCountOCR3_1
	}
	if configProto.MaxPerOracleUnexpiredBlobCumulativePayloadBytes == 0 {
		configProto.MaxPerOracleUnexpiredBlobCumulativePayloadBytes = defaultMaxPerOracleUnexpiredBlobCumulativePayloadBytesOCR3_1
	}
	o.capability.setRequestTimeout(configProto.RequestTimeout.AsDuration())

	rp, err := newReportingPluginOCR3_1(o.store, o.capability, config, &configProto, o.lggr)
	if err != nil {
		return nil, ocr3_1types.ReportingPluginInfo1{}, err
	}

	info := ocr3_1types.ReportingPluginInfo1{
		Name: "OCR3_1 CRE Consensus Plugin",
		Limits: ocr3_1types.ReportingPluginLimits{
			MaxQueryBytes:                int(configProto.MaxQueryLengthBytes),
			MaxObservationBytes:          int(configProto.MaxObservationLengthBytes),
			MaxReportsPlusPrecursorBytes: int(configProto.MaxReportsPlusPrecursorBytes),
			MaxReportBytes:               int(configProto.MaxReportLengthBytes),
			MaxReportCount:               int(configProto.MaxReportCount),

			MaxKeyValueModifiedKeys:                int(configProto.MaxKeyValueModifiedKeys),
			MaxKeyValueModifiedKeysPlusValuesBytes: int(configProto.MaxKeyValueModifiedKeysPlusValuesBytes),

			MaxBlobPayloadBytes:                             int(configProto.MaxBlobPayloadBytes),
			MaxPerOracleUnexpiredBlobCount:                  int(configProto.MaxPerOracleUnexpiredBlobCount),
			MaxPerOracleUnexpiredBlobCumulativePayloadBytes: int(configProto.MaxPerOracleUnexpiredBlobCumulativePayloadBytes),
		},
	}
	return rp, info, nil
}

func (o *factoryOCR3_1) Start(ctx context.Context) error {
	return o.StartOnce("OCR3_1ReportingPlugin", func() error { return nil })
}

func (o *factoryOCR3_1) Close() error {
	return o.StopOnce("OCR3_1ReportingPlugin", func() error { return nil })
}

func (o *factoryOCR3_1) Name() string                   { return o.lggr.Name() }
func (o *factoryOCR3_1) HealthReport() map[string]error { return map[string]error{o.Name(): o.Healthy()} }

// Ensure factoryOCR3_1 satisfies the libocr interface.
var _ ocr3_1types.ReportingPluginFactory[[]byte] = (*factoryOCR3_1)(nil)

// ensure time import is retained regardless of future edits
var _ = time.Second
