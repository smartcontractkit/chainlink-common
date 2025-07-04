package datafeeds

import (
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

// secureMintReport represents the inner report structure
type secureMintReport struct {
	ConfigDigest ocr2types.ConfigDigest `json:"configDigest"`
	SeqNr        uint64                 `json:"seqNr"`
	Block        uint64                 `json:"block"`
	Mintable     *big.Int               `json:"mintable"`
}

// chainSelector represents the chain selector type
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
	// Create a default SecureMintAggregatorConfig
	config := SecureMintAggregatorConfig{
		TargetChainSelector: 1, // default to Ethereum mainnet
	}
	if err := m.UnwrapTo(&config); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap values.Map to SecureMintAggregatorConfig: %w", err)
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
			// Extract OCRTriggerEvent from the observation
			triggerEvent := &capabilities.OCRTriggerEvent{}
			if err := observation.UnwrapTo(triggerEvent); err != nil {
				lggr.Warnw("could not unwrap OCRTriggerEvent", "err", err, "observation", observation)
				continue
			}

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

// TODO(gg): update this piece to comply with KeystoneForwarder/DF Cache
// createOutcome creates the final aggregation outcome in the same format as feeds aggregator
func (a *SecureMintAggregator) createOutcome(lggr logger.Logger, report *secureMintReport) (*types.AggregationOutcome, error) {
	// Convert chain selector to bytes for feed ID
	chainSelectorBytes := big.NewInt(int64(a.config.TargetChainSelector)).Bytes()

	// Create the output in the same format as the feeds aggregator
	toWrap := []any{
		map[EVMEncoderKey]any{
			FeedIDOutputFieldName:     chainSelectorBytes,
			RawReportOutputFieldName:  report.Mintable.Bytes(), // Use Mintable as the raw report
			PriceOutputFieldName:      report.Mintable.Bytes(), // Use Mintable as the price
			TimestampOutputFieldName:  int64(report.Block),     // Use Block as timestamp // TODO(gg): fix
			RemappedIDOutputFieldName: chainSelectorBytes,      // Use chain selector as remapped ID
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
