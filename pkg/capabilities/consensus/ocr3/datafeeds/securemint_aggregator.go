package datafeeds

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr3types "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	ErrNoMatchingChainSelector = errors.New("no matching chain selector found")
	ErrSequenceNumberTooLow    = errors.New("sequence number too low")
)

// secureMintReport represents the inner report structure, mimics the Report type in the SM plugin repo
type secureMintReport struct {
	ConfigDigest ocr2types.ConfigDigest `json:"configDigest"`
	SeqNr        uint64                 `json:"seqNr"`
	Block        uint64                 `json:"block"`
	Mintable     *big.Int               `json:"mintable"`
}

// chainSelector represents the chain selector type, mimics the ChainSelector type in the SM plugin repo
type chainSelector int64

// SecureMintAggregatorConfig is the config for the SecureMint aggregator.
// This aggregator is designed to pick out reports for a specific chain selector.
type SecureMintAggregatorConfig struct {
	// TargetChainSelector is the chain selector to look for
	TargetChainSelector chainSelector `mapstructure:"targetChainSelector"`
}

// ToMap converts the SecureMintAggregatorConfig to a values.Map, which is suitable for the
// [NewAggregator] function in the OCR3 Aggregator interface.
func (c SecureMintAggregatorConfig) ToMap() (*values.Map, error) {
	v, err := values.WrapMap(c)
	if err != nil {
		// this should never happen since we are wrapping a struct
		return &values.Map{}, fmt.Errorf("failed to wrap SecureMintAggregatorConfig: %w", err)
	}
	return v, nil
}

func NewSecureMintConfig(m values.Map) (SecureMintAggregatorConfig, error) {
	var config SecureMintAggregatorConfig
	if err := m.UnwrapTo(&config); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap values.Map %+v to SecureMintAggregatorConfig: %w", m, err)
	}

	return config, nil
}

var _ types.Aggregator = (*SecureMintAggregator)(nil)

type SecureMintAggregator struct {
	config SecureMintAggregatorConfig
}

// NewSecureMintAggregator creates a new SecureMintAggregator instance based on the provided configuration.
// The config should be a [values.Map] that represents the [SecureMintAggregatorConfig]. See [SecureMintAggregatorConfig.ToMap]
func NewSecureMintAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := parseSecureMintConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &SecureMintAggregator{
		config: parsedConfig,
	}, nil
}

// Aggregate implements the Aggregator interface
// This implementation:
// 1. Extracts OCRTriggerEvent from observations
// 2. Deserializes the inner ReportWithInfo to get chain selector and report
// 3. Validates chain selector matches target and sequence number is higher than previous
// 4. Returns the report in the same format as feeds aggregator
func (a *SecureMintAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "SecureMintAggregator")

	lggr.Debugw("Aggregate called", "config", a.config, "observations", len(observations), "f", f, "previousOutcome", previousOutcome)

	if len(observations) == 0 {
		return nil, errors.New("no observations")
	}

	// Extract and validate reports from all observations
	validReports, err := a.extractAndValidateReports(lggr, observations, previousOutcome)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and validate reports: %w", err)
	}

	// TODO(gg): heartbeat check?

	if len(validReports) == 0 {
		lggr.Infow("no reports selected", "targetChainSelector", a.config.TargetChainSelector)
		return &types.AggregationOutcome{
			ShouldReport: false,
		}, nil
	}

	// Take the first valid report
	targetReport := validReports[0]

	// Create the aggregation outcome
	outcome, err := a.createOutcome(lggr, targetReport)
	if err != nil {
		return nil, fmt.Errorf("failed to create outcome: %w", err)
	}

	lggr.Debugw("SecureMint Aggregate complete", "targetChainSelector", a.config.TargetChainSelector)
	return outcome, nil
}

// extractAndValidateReports extracts OCRTriggerEvent from observations and validates them
func (a *SecureMintAggregator) extractAndValidateReports(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value, previousOutcome *types.AggregationOutcome) ([]*secureMintReport, error) {
	var validReports []*secureMintReport
	var sequenceNumberTooLow bool
	var foundMatchingChainSelector bool
	previousSeqNr := uint64(0)
	if previousOutcome != nil {
		previousSeqNr = previousOutcome.LastSeenAt
	}

	for nodeID, nodeObservations := range observations {
		lggr = logger.With(lggr, "nodeID", nodeID)

		for _, observation := range nodeObservations {
			lggr.Debugw("processing observation", "observation", observation)
			lggr.Debugf("processing observation %+v", observation)

			// Extract OCRTriggerEvent from the observation
			triggerEvent := &capabilities.OCRTriggerEvent{}
			if err := observation.UnwrapTo(triggerEvent); err != nil {
				lggr.Warnw("could not unwrap OCRTriggerEvent", "err", err, "observation", observation)
				continue
			}

			lggr.Debugw("triggerEvent", "triggerEvent", triggerEvent)

			// Deserialize the ReportWithInfo
			var reportWithInfo ocr3types.ReportWithInfo[chainSelector]
			if err := json.Unmarshal(triggerEvent.Report, &reportWithInfo); err != nil {
				lggr.Errorw("failed to unmarshal ReportWithInfo", "err", err)
				continue
			}

			// Check if chain selector matches target
			if reportWithInfo.Info != a.config.TargetChainSelector {
				lggr.Debugw("chain selector mismatch", "got", reportWithInfo.Info, "expected", a.config.TargetChainSelector)
				continue
			}

			// We found a matching chain selector
			foundMatchingChainSelector = true

			// Validate sequence number
			if triggerEvent.SeqNr <= previousSeqNr {
				lggr.Warnw("sequence number too low", "seqNr", triggerEvent.SeqNr, "previousSeqNr", previousSeqNr)
				sequenceNumberTooLow = true
				continue
			}

			// Deserialize the inner secureMintReport
			var innerReport secureMintReport
			if err := json.Unmarshal(reportWithInfo.Report, &innerReport); err != nil {
				lggr.Errorw("failed to unmarshal secureMintReport", "err", err)
				continue
			}

			validReports = append(validReports, &innerReport)
		}
	}

	// Return appropriate error based on what we found
	if !foundMatchingChainSelector {
		lggr.Infow("no reports found for target chain selector, ignoring", "targetChainSelector", a.config.TargetChainSelector)
		return nil, nil
	}

	if sequenceNumberTooLow {
		return nil, fmt.Errorf("%w: all reports had sequence numbers <= %d", ErrSequenceNumberTooLow, previousSeqNr)
	}

	return validReports, nil
}

// createOutcome creates the final aggregation outcome which can be sent to the KeystoneForwarder
func (a *SecureMintAggregator) createOutcome(lggr logger.Logger, report *secureMintReport) (*types.AggregationOutcome, error) {
	// Convert chain selector to bytes for feed ID TODO(gg): check if this works for us
	var chainSelectorAsFeedId [32]byte
	binary.BigEndian.PutUint64(chainSelectorAsFeedId[24:], uint64(a.config.TargetChainSelector)) // right-aligned

	smReportAsPrice, err := packSecureMintReportForIntoUint224ForEVM(report.Mintable, report.Block)
	if err != nil {
		return nil, fmt.Errorf("failed to pack secure mint report for into uint224: %w", err)
	}

	// Create the output in the same format as the feeds aggregator
	// abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports"
	toWrap := []any{
		map[EVMEncoderKey]any{
			FeedIDOutputFieldName: chainSelectorAsFeedId,
			// RawReportOutputFieldName:  packedReport, // TODO(gg): check if we need this
			PriceOutputFieldName:     smReportAsPrice,
			TimestampOutputFieldName: int64(report.Block), // TODO(gg): not sure if we want this
		},
	}

	wrappedReport, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wrap report: %w", err)
	}

	reportsProto := values.Proto(wrappedReport)

	// Store the sequence number in metadata for next round
	metadata := []byte{byte(report.SeqNr)} // Simple metadata for now

	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         metadata,
		LastSeenAt:       report.SeqNr,
		ShouldReport:     true, // Always report since we found and verified the target report
	}, nil
}

// parseSecureMintConfig parses the user-facing, type-less, SecureMint aggregator config into the internal typed config.
func parseSecureMintConfig(config values.Map) (SecureMintAggregatorConfig, error) {
	parsedConfig := SecureMintAggregatorConfig{
		TargetChainSelector: 1, // default value
	}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap config: %w", err)
	}

	// Validate configuration
	if parsedConfig.TargetChainSelector <= 0 {
		return SecureMintAggregatorConfig{}, fmt.Errorf("targetChainSelector is required")
	}

	return parsedConfig, nil
}

var maxMintable = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)) // 2^128 - 1

// packSecureMintReportForIntoUint224ForEVM packs the mintable and block number into a single uint224 so that it can be used as a price in the DF Cache contract
// (top 32 - not used / middle 64 - block number / lower 128 - mintable amount)
func packSecureMintReportForIntoUint224ForEVM(mintable *big.Int, blockNumber uint64) (*big.Int, error) {
	// Handle nil mintable
	if mintable == nil {
		return nil, fmt.Errorf("mintable cannot be nil")
	}

	// Validate that mintable fits in 128 bits
	if mintable.Cmp(maxMintable) > 0 {
		return nil, fmt.Errorf("mintable amount %v exceeds maximum 128-bit value %v", mintable, maxMintable)
	}

	packed := big.NewInt(0)
	// Put mintable in lower 128 bits
	packed.Or(packed, mintable)

	// Put block number in middle 64 bits (bits 128-191)
	blockNumberAsBigInt := new(big.Int).SetUint64(blockNumber)
	packed.Or(packed, new(big.Int).Lsh(blockNumberAsBigInt, 128))

	return packed, nil
}
