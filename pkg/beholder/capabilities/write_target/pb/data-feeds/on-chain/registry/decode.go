package registry

import (
	"fmt"
	"math"
	"math/big"

	wt_msg "github.com/smartcontractkit/chainlink-common/pkg/beholder/capabilities/write_target/pb/platform/write-target"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/report/data_feeds"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/report/platform"

	mercury_vX "github.com/smartcontractkit/chainlink-common/pkg/beholder/report/mercury/common"
	mercury_v3 "github.com/smartcontractkit/chainlink-common/pkg/beholder/report/mercury/v3"
	mercury_v4 "github.com/smartcontractkit/chainlink-common/pkg/beholder/report/mercury/v4"
)

// DecodeAsFeedUpdated decodes a 'platform.write-target.WriteConfirmed' message
// as a 'data-feeds.registry.ReportProcessed' message
func DecodeAsFeedUpdated(m *wt_msg.WriteConfirmed) ([]*FeedUpdated, error) {
	// Decode the confirmed report (WT -> DF contract event)
	r, err := platform.Decode(m.Report)
	if err != nil {
		return nil, fmt.Errorf("failed to decode report: %w", err)
	}

	// Decode the underlying Data Feeds reports
	reports, err := data_feeds.Decode(r.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Data Feeds report: %w", err)
	}

	// Allocate space for the messages (event per updated feed)
	msgs := make([]*FeedUpdated, 0, len(*reports))

	// Iterate over the underlying Mercury reports
	for _, rf := range *reports {
		// Notice: we assume that Mercury will be the only source of reports used for Data Feeds,
		// at least for the foreseeable future. If this assumption changes, we should check the
		// the report type here (potentially encoded in the feed ID) and decode accordingly.

		// Decode the common Mercury report
		rm, err := mercury_vX.Decode(rf.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Mercury report: %w", err)
		}

		// Parse the report type
		t := mercury_vX.GetReportType(rm.FeedId)

		// Notice: we publish the DataFeed FeedID, not the unrelying DataStream FeedID
		feedID := data_feeds.FeedID(rf.FeedId)

		switch t {
		case uint16(3):
			rm, err := mercury_v3.Decode(rf.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode Mercury v%d report: %w", t, err)
			}

			msgs = append(msgs, &FeedUpdated{
				// Event data
				FeedId:                feedID.String(),
				ObservationsTimestamp: rm.ObservationsTimestamp,
				Benchmark:             rm.BenchmarkPrice.Bytes(), // Map big.Int as []byte
				Report:                rf.Data,

				// Notice: i192 will not fit if scaled number bigger than f64
				BenchmarkVal: toBenchmarkVal(feedID, rm.BenchmarkPrice),

				// Head data - when was the event produced on-chain
				BlockHash:      m.BlockHash,
				BlockHeight:    m.BlockHeight,
				BlockTimestamp: m.BlockTimestamp,

				// Transaction data - info about the tx that mained the event (optional)
				// Notice: we skip SOME head/tx data here (unknown), as we map from 'platform.write-target.WriteConfirmed'
				// and not from tx/event data (e.g., 'platform.write-target.WriteTxConfirmed')
				TxSender:   m.Transmitter,
				TxReceiver: m.Forwarder,

				// Execution Context - Source
				MetaSourceId: m.MetaSourceId,

				// Execution Context - Chain
				MetaChainFamilyName: m.MetaChainFamilyName,
				MetaChainId:         m.MetaChainId,
				MetaNetworkName:     m.MetaNetworkName,
				MetaNetworkNameFull: m.MetaNetworkNameFull,

				// Execution Context - Workflow (capabilities.RequestMetadata)
				MetaWorkflowId:               m.MetaWorkflowId,
				MetaWorkflowOwner:            m.MetaWorkflowOwner,
				MetaWorkflowExecutionId:      m.MetaWorkflowExecutionId,
				MetaWorkflowName:             m.MetaWorkflowName,
				MetaWorkflowDonId:            m.MetaWorkflowDonId,
				MetaWorkflowDonConfigVersion: m.MetaWorkflowDonConfigVersion,
				MetaReferenceId:              m.MetaReferenceId,

				// Execution Context - Capability
				MetaCapabilityType:           m.MetaCapabilityType,
				MetaCapabilityId:             m.MetaCapabilityId,
				MetaCapabilityTimestampStart: m.MetaCapabilityTimestampStart,
				MetaCapabilityTimestampEmit:  m.MetaCapabilityTimestampEmit,
			})
		case uint16(4):
			rm, err := mercury_v4.Decode(rf.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode Mercury v%d report: %w", t, err)
			}

			msgs = append(msgs, &FeedUpdated{
				// Event data
				FeedId:                feedID.String(),
				ObservationsTimestamp: rm.ObservationsTimestamp,
				Benchmark:             rm.BenchmarkPrice.Bytes(), // Map big.Int as []byte
				Report:                rf.Data,

				// Notice: i192 will not fit if scaled number bigger than f64
				BenchmarkVal: toBenchmarkVal(feedID, rm.BenchmarkPrice),

				// Notice: we skip head/tx data here (unknown), as we map from 'platform.write-target.WriteConfirmed'
				// and not from tx/event data (e.g., 'platform.write-target.WriteTxConfirmed')

				BlockHash:      m.BlockHash,
				BlockHeight:    m.BlockHeight,
				BlockTimestamp: m.BlockTimestamp,

				// Execution Context - Source
				MetaSourceId: m.MetaSourceId,

				// Execution Context - Chain
				MetaChainFamilyName: m.MetaChainFamilyName,
				MetaChainId:         m.MetaChainId,
				MetaNetworkName:     m.MetaNetworkName,
				MetaNetworkNameFull: m.MetaNetworkNameFull,

				// Execution Context - Workflow (capabilities.RequestMetadata)
				MetaWorkflowId:               m.MetaWorkflowId,
				MetaWorkflowOwner:            m.MetaWorkflowOwner,
				MetaWorkflowExecutionId:      m.MetaWorkflowExecutionId,
				MetaWorkflowName:             m.MetaWorkflowName,
				MetaWorkflowDonId:            m.MetaWorkflowDonId,
				MetaWorkflowDonConfigVersion: m.MetaWorkflowDonConfigVersion,
				MetaReferenceId:              m.MetaReferenceId,

				// Execution Context - Capability
				MetaCapabilityType:           m.MetaCapabilityType,
				MetaCapabilityId:             m.MetaCapabilityId,
				MetaCapabilityTimestampStart: m.MetaCapabilityTimestampStart,
				MetaCapabilityTimestampEmit:  m.MetaCapabilityTimestampEmit,
			})
		default:
			return nil, fmt.Errorf("unsupported Mercury report type: %d", t)
		}
	}

	return msgs, nil
}

// toBenchmarkVal returns the benchmark i192 on-chain value decoded as an double (float64), scaled by number of decimals (e.g., 1e-18)
// Where the number of decimals is extracted from the feed ID.
//
// This is the largest type Prometheus supports, and this conversion can overflow but so far was sufficient
// for most use-cases. For big numbers, benchmark bytes should be used instead.
//
// Returns `math.NaN()` if report data type not a number, or `+/-Inf` if number doesn't fit in double.
func toBenchmarkVal(feedID data_feeds.FeedID, val *big.Int) float64 {
	// Return NaN if the value is nil
	if val == nil {
		return math.NaN()
	}

	// Get the number of decimals from the feed ID
	t := feedID.GetDataType()
	decimals, isNumber := data_feeds.GetDecimals(t)

	// Return NaN if the value is not a number
	if !isNumber {
		return math.NaN()
	}

	// Convert the i192 to a big Float, scaled by the number of decimals
	valF := new(big.Float).SetInt(val)

	if decimals > 0 {
		denominator := big.NewFloat(math.Pow10(int(decimals)))
		valF = new(big.Float).Quo(valF, denominator)
	}

	// Notice: this can overflow, but so far was sufficient for most use-cases
	// On overflow, returns +/-Inf (valid Prometheus value)
	valRes, _ := valF.Float64()
	return valRes
}
