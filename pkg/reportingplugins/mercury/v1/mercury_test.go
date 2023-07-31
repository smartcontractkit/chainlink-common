package mercury_v1

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
	"github.com/stretchr/testify/require"
)

type testDataSource struct{}

func (ds testDataSource) Observe(ctx context.Context, repts ocrtypes.ReportTimestamp, fetchMaxFinalizedTimestamp bool) (Observation, error) {
	return Observation{}, nil
}

type testReportCodec struct {
	observationTimestamp uint32
	validFromTimestamp   uint32
	builtReport          ocrtypes.Report
}

func (rc *testReportCodec) BuildReport(paos []ParsedAttributedObservation, f int, validFromTimestamp uint32, expiresAt uint32) (ocrtypes.Report, error) {
	rc.validFromTimestamp = validFromTimestamp
	return rc.builtReport, nil
}

func (rc testReportCodec) MaxReportLength(n int) (int, error) {
	return 100, nil
}

func (rc testReportCodec) ObservationTimestampFromReport(ocrtypes.Report) (uint32, error) {
	return rc.observationTimestamp, nil
}

func newAttributedObservation(t *testing.T, p *MercuryObservationProto) ocrtypes.AttributedObservation {
	marshalledObs, err := proto.Marshal(p)
	require.NoError(t, err)
	return ocrtypes.AttributedObservation{
		Observation: ocrtypes.Observation(marshalledObs),
		Observer:    commontypes.OracleID(42),
	}
}

func Test_Report(t *testing.T) {
	dataSource := testDataSource{}
	codec := &testReportCodec{}
	offchainConfig := mercury.OffchainConfig{
		ExpirationWindow: 1,
		BaseUSDFeeCents:  100,
	}
	onchainConfig := mercury.OnchainConfig{
		Min: big.NewInt(1),
		Max: big.NewInt(1000),
	}
	rp := reportingPlugin{
		offchainConfig:           offchainConfig,
		onchainConfig:            onchainConfig,
		dataSource:               dataSource,
		logger:                   logger.Test(t),
		reportCodec:              codec,
		configDigest:             ocrtypes.ConfigDigest{},
		f:                        1,
		latestAcceptedEpochRound: mercury.EpochRound{},
		latestAcceptedMedian:     big.NewInt(0),
		maxReportLength:          100,
	}

	t.Run("when previous report is not nil", func(t *testing.T) {
		previousReport := ocrtypes.Report{}

		t.Run("reports if more than f+1 observations are valid", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 45,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(234)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(230)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(240)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 47,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 39,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}
			codec.observationTimestamp = 11
			codec.builtReport = ocrtypes.Report{1, 2, 3}

			should, report, err := rp.Report(ocrtypes.ReportTimestamp{}, previousReport, aos)
			assert.True(t, should)
			assert.NoError(t, err)
			assert.Equal(t, codec.builtReport, report)
			assert.Equal(t, codec.validFromTimestamp, codec.observationTimestamp)
		})

		t.Run("reports if no f+1 maxFinalizedTimestamp observations available", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 45,

					PricesValid: false,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 47,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 39,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}
			codec.observationTimestamp = 22
			codec.builtReport = ocrtypes.Report{2, 3, 4}

			should, report, err := rp.Report(ocrtypes.ReportTimestamp{}, previousReport, aos)
			assert.True(t, should)
			assert.NoError(t, err)
			assert.Equal(t, codec.builtReport, report)
			assert.Equal(t, codec.validFromTimestamp, codec.observationTimestamp)
		})

		t.Run("errors when less than f+1 valid observations available", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
			}

			_, _, err := rp.Report(ocrtypes.ReportTimestamp{}, previousReport, aos)
			assert.EqualError(t, err, "only received 1 valid attributed observations, but need at least f+1 (2)")
		})
	})

	t.Run("when previous report is nil", func(t *testing.T) {

		t.Run("reports if more than f+1 observations are valid", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 45,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(234)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(230)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(240)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 47,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 39,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      120,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}
			codec.builtReport = ocrtypes.Report{1, 2, 3}

			should, report, err := rp.Report(ocrtypes.ReportTimestamp{}, nil, aos)
			assert.True(t, should)
			assert.NoError(t, err)
			assert.Equal(t, codec.builtReport, report)
			assert.Equal(t, int(codec.validFromTimestamp), 40)
		})

		t.Run("errors when less than f+1 maxFinalizedTimestamp observations available", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 45,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(234)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(230)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(240)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 47,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 39,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: false,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}

			should, _, err := rp.Report(ocrtypes.ReportTimestamp{}, nil, aos)
			assert.False(t, should)
			assert.EqualError(t, err, "fewer than f+1 observations have a valid maxFinalizedTimestamp (got: 1/4)")
		})

		t.Run("errors when cannot come to consensus on MaxFinalizedTimestamp", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 42,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      40,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 45,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(234)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(230)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(240)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      41,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 47,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      42,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 39,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      43,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}

			should, _, err := rp.Report(ocrtypes.ReportTimestamp{}, nil, aos)
			assert.False(t, should)
			assert.EqualError(t, err, "no valid maxFinalizedTimestamp with at least f+1 votes (got counts: map[40:1 41:1 42:1 43:1])")
		})

		t.Run("maxFinalizedTimestamp equals to observationTimestamp when consensus on MaxFinalizedTimestamp = 0", func(t *testing.T) {
			aos := []ocrtypes.AttributedObservation{
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 55,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(123)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(120)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(130)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      0,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.1e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.1e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 55,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(234)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(230)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(240)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      0,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.2e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.2e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 55,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(345)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(340)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(350)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      0,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.3e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.3e18)),
					NativeFeeValid: true,
				}),
				newAttributedObservation(t, &MercuryObservationProto{
					Timestamp: 55,

					BenchmarkPrice: mercury.MustEncodeValueInt192(big.NewInt(456)),
					Bid:            mercury.MustEncodeValueInt192(big.NewInt(450)),
					Ask:            mercury.MustEncodeValueInt192(big.NewInt(460)),
					PricesValid:    true,

					MaxFinalizedTimestamp:      43,
					MaxFinalizedTimestampValid: true,

					LinkFee:        mercury.MustEncodeValueInt192(big.NewInt(1.4e18)),
					LinkFeeValid:   true,
					NativeFee:      mercury.MustEncodeValueInt192(big.NewInt(2.4e18)),
					NativeFeeValid: true,
				}),
			}
			codec.builtReport = ocrtypes.Report{7, 8, 9}

			should, report, err := rp.Report(ocrtypes.ReportTimestamp{}, nil, aos)
			assert.True(t, should)
			assert.NoError(t, err)
			assert.Equal(t, codec.builtReport, report)
			assert.Equal(t, int(codec.validFromTimestamp), 55)
		})
	})
}
