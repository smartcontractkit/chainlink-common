package datafeeds

import (
	"errors"
	"fmt"
	"strings"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	ErrInvalidConfig         = errors.New("invalid config")
	ErrInsufficientConsensus = errors.New("insufficient consensus")
	ErrEmptyObservation      = errors.New("empty observation")
	ErrNoEthReportFound      = errors.New("no eth report found")
)

// SecureMintAggregatorConfig is the config for the SecureMint aggregator.
// This aggregator is designed to pick out a specific report (hardcoded to "eth" for now),
// verify its signatures, and reattest it.
type SecureMintAggregatorConfig struct {
	// TargetFeedID is the feed ID to look for (hardcoded to "eth" for now)
	TargetFeedID string `mapstructure:"targetFeedId"`
	// AllowedSigners are the signers that are allowed to sign reports
	AllowedSigners [][]byte `mapstructure:"allowedSigners"`
	// MinRequiredSignatures is the minimum number of signatures required
	MinRequiredSignatures int `mapstructure:"minRequiredSignatures"`
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
		TargetFeedID: "eth", // hardcoded as requested
	}
	if err := m.UnwrapTo(&config); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap values.Map to SecureMintAggregatorConfig: %w", err)
	}

	return config, nil
}

var _ types.Aggregator = (*SecureMintAggregator)(nil)

type SecureMintAggregator struct {
	config      SecureMintAggregatorConfig
	reportCodec datastreams.ReportCodec
}

// NewSecureMintAggregator creates a new SecureMintAggregator instance based on the provided configuration.
// The config should be a [values.Map] that represents the [SecureMintAggregatorConfig]. See [SecureMintAggregatorConfig.ToMap]
func NewSecureMintAggregator(config values.Map, reportCodec datastreams.ReportCodec) (types.Aggregator, error) {
	parsedConfig, err := parseSecureMintConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &SecureMintAggregator{
		config:      parsedConfig,
		reportCodec: reportCodec,
	}, nil
}

// Aggregate implements the Aggregator interface
// This implementation:
// 1. Extracts reports from observations
// 2. Finds the target "eth" report
// 3. Verifies signatures on the report
// 4. Reattests the report by returning it
func (a *SecureMintAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "SecureMintAggregator")
	if len(observations) == 0 {
		return nil, ErrEmptyObservation
	}

	// Extract reports from all observations
	allReports, err := a.extractReports(lggr, observations)
	if err != nil {
		return nil, fmt.Errorf("failed to extract reports: %w", err)
	}

	// Find the target "eth" report
	targetReport, err := a.findTargetReport(lggr, allReports)
	if err != nil {
		return nil, fmt.Errorf("failed to find target report: %w", err)
	}

	// Verify signatures on the target report
	if err := a.verifySignatures(lggr, targetReport); err != nil {
		return nil, fmt.Errorf("failed to verify signatures: %w", err)
	}

	// Reattest the report by returning it
	reattestedReport, err := a.reattestReport(lggr, targetReport)
	if err != nil {
		return nil, fmt.Errorf("failed to reattest report: %w", err)
	}

	// Create the aggregation outcome
	outcome, err := a.createOutcome(lggr, reattestedReport)
	if err != nil {
		return nil, fmt.Errorf("failed to create outcome: %w", err)
	}

	lggr.Debugw("SecureMint Aggregate complete", "targetFeedID", a.config.TargetFeedID)
	return outcome, nil
}

// extractReports extracts all reports from the observations
func (a *SecureMintAggregator) extractReports(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value) ([]datastreams.FeedReport, error) {
	var allReports []datastreams.FeedReport

	for nodeID, nodeObservations := range observations {
		lggr = logger.With(lggr, "nodeID", nodeID)

		// Expect exactly one observation per node
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			lggr.Warn("empty observations")
			continue
		}
		if len(nodeObservations) > 1 {
			lggr.Warn("more than one observation")
			continue
		}

		// Extract reports from the observation
		reports, err := a.reportCodec.Unwrap(nodeObservations[0])
		if err != nil {
			lggr.Warnw("could not unwrap reports", "err", err)
			continue
		}

		allReports = append(allReports, reports...)
	}

	return allReports, nil
}

// findTargetReport finds the report with the target feed ID (hardcoded to "eth")
func (a *SecureMintAggregator) findTargetReport(lggr logger.Logger, reports []datastreams.FeedReport) (*datastreams.FeedReport, error) {
	for _, report := range reports {
		// Check if this report is for the target feed ID
		if strings.Contains(strings.ToLower(report.FeedID), strings.ToLower(a.config.TargetFeedID)) {
			lggr.Debugw("found target report", "feedID", report.FeedID, "targetFeedID", a.config.TargetFeedID)
			return &report, nil
		}
	}

	return nil, fmt.Errorf("%w: no report found for target feed ID %s", ErrNoEthReportFound, a.config.TargetFeedID)
}

// verifySignatures verifies the signatures on the target report
func (a *SecureMintAggregator) verifySignatures(lggr logger.Logger, report *datastreams.FeedReport) error {
	// Use the report codec to validate signatures
	if err := a.reportCodec.Validate(*report, a.config.AllowedSigners, a.config.MinRequiredSignatures); err != nil {
		return fmt.Errorf("signature validation failed: %w", err)
	}

	lggr.Debugw("signatures verified successfully", "feedID", report.FeedID, "nSignatures", len(report.Signatures))
	return nil
}

// reattestReport reattests the verified report
// For now, this simply returns the report as-is, but could be extended to add new signatures or modify the report
func (a *SecureMintAggregator) reattestReport(lggr logger.Logger, report *datastreams.FeedReport) (*datastreams.FeedReport, error) {
	// For now, we simply return the report as-is
	// In a real implementation, this might involve:
	// - Adding new signatures
	// - Modifying the report content
	// - Adding additional metadata

	lggr.Debugw("report reattested", "feedID", report.FeedID)
	return report, nil
}

// createOutcome creates the final aggregation outcome
func (a *SecureMintAggregator) createOutcome(lggr logger.Logger, report *datastreams.FeedReport) (*types.AggregationOutcome, error) {
	// Create the output in the same format as the feeds aggregator
	toWrap := []any{
		map[EVMEncoderKey]any{
			FeedIDOutputFieldName:     []byte(report.FeedID),
			RawReportOutputFieldName:  report.FullReport,
			PriceOutputFieldName:      report.BenchmarkPrice,
			TimestampOutputFieldName:  report.ObservationTimestamp,
			RemappedIDOutputFieldName: []byte(report.FeedID), // Use original feed ID as remapped ID
		},
	}

	wrappedReport, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wrap report: %w", err)
	}

	reportsProto := values.Proto(wrappedReport)

	// Create empty metadata since we don't need to maintain state between rounds
	metadata := []byte{}

	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         metadata,
		ShouldReport:     true, // Always report since we found and verified the target report
	}, nil
}

// parseSecureMintConfig parses the user-facing, type-less, SecureMint aggregator config into the internal typed config.
func parseSecureMintConfig(config values.Map) (SecureMintAggregatorConfig, error) {
	parsedConfig := SecureMintAggregatorConfig{
		TargetFeedID: "eth", // default value
	}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap config: %w", err)
	}

	// Validate configuration
	if parsedConfig.TargetFeedID == "" {
		return SecureMintAggregatorConfig{}, fmt.Errorf("%w: targetFeedId is required", ErrInvalidConfig)
	}

	if len(parsedConfig.AllowedSigners) == 0 {
		return SecureMintAggregatorConfig{}, fmt.Errorf("%w: allowedSigners is required", ErrInvalidConfig)
	}

	if parsedConfig.MinRequiredSignatures <= 0 {
		return SecureMintAggregatorConfig{}, fmt.Errorf("%w: minRequiredSignatures must be greater than 0", ErrInvalidConfig)
	}

	return parsedConfig, nil
}
