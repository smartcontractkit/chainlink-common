package datafeeds

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	// Aggregator outputs reports in the following format:
	//   []Reports{FeedID []byte, RawReport []byte, Price *big.Int, Timestamp int64}
	// Example of a compatible EVM encoder ABI config:
	//   (bytes32 FeedID, bytes RawReport, uint256 Price, uint64 Timestamp)[] Reports
	TopLevelListOutputFieldName = "Reports"
	FeedIDOutputFieldName       = "FeedID"
	RawReportOutputFieldName    = "RawReport"
	PriceOutputFieldName        = "Price"
	TimestampOutputFieldName    = "Timestamp"
	RemappedIDOutputFieldName   = "RemappedID"

	addrLen = 20
)

type aggregatorConfig struct {
	Feeds map[datastreams.FeedID]feedConfig
	// AllowedPartialStaleness is an optional optimization that tries to maximize batching.
	// Once any deviation or heartbeat threshold hits, we will include all other feeds that are
	// within the AllowedPartialStaleness range of their own heartbeat.
	// For example, setting 0.2 will include all feeds that are within 20% of their heartbeat.
	AllowedPartialStaleness    float64 `mapstructure:"-"`
	AllowedPartialStalenessStr string  `mapstructure:"allowedPartialStaleness"`
}

type feedConfig struct {
	Deviation       decimal.Decimal `mapstructure:"-"`
	Heartbeat       int
	DeviationString string `mapstructure:"deviation"`
	RemappedIDHex   string `mapstructure:"remappedId"`
	RemappedID      []byte `mapstructure:"-"`
}

type dataFeedsAggregator struct {
	config      aggregatorConfig
	reportCodec datastreams.ReportCodec
}

var _ types.Aggregator = (*dataFeedsAggregator)(nil)

// This Aggregator has two phases:
//  1. Agree on valid trigger signers by extracting them from event metadata and aggregating using MODE (at least F+1 copies needed).
//  2. For each FeedID, select latest valid report, using signers list obtained in phase 1.
//
// EncodableOutcome is a list of aggregated price points.
// Metadata is a map of feedID -> (timestamp, price) representing onchain state (see DataFeedsOutcomeMetadata proto)
func (a *dataFeedsAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	allowedSigners, minRequiredSignatures, events := a.extractSignersAndPayloads(lggr, observations, f)
	if len(events) > 0 && minRequiredSignatures == 0 {
		return nil, fmt.Errorf("cannot process non-empty observation payloads with minRequiredSignatures set to 0")
	}
	lggr.Debugw("extracted signers", "nAllowedSigners", len(allowedSigners), "minRequired", minRequiredSignatures, "nEvents", len(events))
	// find latest valid report for each feed ID
	latestReportPerFeed := make(map[datastreams.FeedID]datastreams.FeedReport)
	for nodeID, event := range events {
		mercuryReports, err := a.reportCodec.Unwrap(event)
		if err != nil {
			lggr.Errorf("node %d contributed with invalid reports: %v", nodeID, err)
			continue
		}
		for _, report := range mercuryReports {
			latest, ok := latestReportPerFeed[datastreams.FeedID(report.FeedID)]
			if !ok || report.ObservationTimestamp > latest.ObservationTimestamp {
				// lazy signature validation
				if err := a.reportCodec.Validate(report, allowedSigners, minRequiredSignatures); err != nil {
					lggr.Errorf("node %d contributed with an invalid report: %v", nodeID, err)
				} else {
					latestReportPerFeed[datastreams.FeedID(report.FeedID)] = report
				}
			}
		}
	}
	lggr.Debugw("collected latestReportPerFeed", "len", len(latestReportPerFeed))

	currentState, err := a.initializeCurrentState(lggr, previousOutcome)
	if err != nil {
		return nil, err
	}

	reportsNeedingUpdate := []datastreams.FeedReport{}
	allIDs := []string{}
	for feedID := range currentState.FeedInfo {
		allIDs = append(allIDs, feedID)
	}

	lggr.Debugw("determined feeds to check", "nFeedIds", len(allIDs))
	// ensure deterministic order of reportsNeedingUpdate
	sort.Slice(allIDs, func(i, j int) bool { return allIDs[i] < allIDs[j] })
	candidateIDs := []string{}
	for _, feedIDStr := range allIDs {
		previousReportInfo := currentState.FeedInfo[feedIDStr]
		feedID, err2 := datastreams.NewFeedID(feedIDStr)
		if err2 != nil {
			lggr.Errorw("could not convert %s to feedID", "feedID", feedID)
			continue
		}
		latestReport, ok := latestReportPerFeed[feedID]
		if !ok {
			lggr.Errorw("no new Mercury report for feed", "feedID", feedID)
			continue
		}
		config := a.config.Feeds[feedID]
		oldPrice := big.NewInt(0).SetBytes(previousReportInfo.BenchmarkPrice)
		newPrice := big.NewInt(0).SetBytes(latestReport.BenchmarkPrice)
		currDeviation := deviation(oldPrice, newPrice)
		currStaleness := latestReport.ObservationTimestamp - previousReportInfo.ObservationTimestamp
		lggr.Debugw("checking deviation and heartbeat",
			"feedID", feedID,
			"currentTs", latestReport.ObservationTimestamp,
			"oldTs", previousReportInfo.ObservationTimestamp,
			"currStaleness", currStaleness,
			"heartbeat", config.Heartbeat,
			"oldPrice", oldPrice,
			"newPrice", newPrice,
			"currDeviation", currDeviation,
			"deviation", config.Deviation.InexactFloat64(),
		)
		if currStaleness > int64(config.Heartbeat) ||
			currDeviation > config.Deviation.InexactFloat64() {
			previousReportInfo.ObservationTimestamp = latestReport.ObservationTimestamp
			previousReportInfo.BenchmarkPrice = latestReport.BenchmarkPrice
			reportsNeedingUpdate = append(reportsNeedingUpdate, latestReport)
		} else if float64(currStaleness) > float64(config.Heartbeat)*(1.0-a.config.AllowedPartialStaleness) {
			candidateIDs = append(candidateIDs, feedIDStr)
		}
	}

	// optimization that allows for more efficient batching
	if len(reportsNeedingUpdate) > 0 {
		for _, feedIDStr := range candidateIDs {
			previousReportInfo := currentState.FeedInfo[feedIDStr]
			latestReport := latestReportPerFeed[datastreams.FeedID(feedIDStr)]
			previousReportInfo.ObservationTimestamp = latestReport.ObservationTimestamp
			previousReportInfo.BenchmarkPrice = latestReport.BenchmarkPrice
			reportsNeedingUpdate = append(reportsNeedingUpdate, latestReport)
		}
	}

	marshalledState, err := proto.MarshalOptions{Deterministic: true}.Marshal(currentState)
	if err != nil {
		return nil, err
	}

	var toWrap []any
	for _, report := range reportsNeedingUpdate {
		feedID := datastreams.FeedID(report.FeedID).Bytes()
		remappedID := a.config.Feeds[datastreams.FeedID(report.FeedID)].RemappedID
		if len(remappedID) == 0 { // fall back to original ID
			remappedID = feedID[:]
		}
		toWrap = append(toWrap,
			map[string]any{
				FeedIDOutputFieldName:     feedID[:],
				RawReportOutputFieldName:  report.FullReport,
				PriceOutputFieldName:      big.NewInt(0).SetBytes(report.BenchmarkPrice),
				TimestampOutputFieldName:  report.ObservationTimestamp,
				RemappedIDOutputFieldName: remappedID,
			})
	}

	wrappedReportsNeedingUpdates, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, err
	}
	reportsProto := values.Proto(wrappedReportsNeedingUpdates)

	lggr.Debugw("Aggregate complete", "nReportsNeedingUpdate", len(reportsNeedingUpdate))
	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         marshalledState,
		ShouldReport:     len(reportsNeedingUpdate) > 0,
	}, nil
}

func (a *dataFeedsAggregator) initializeCurrentState(lggr logger.Logger, previousOutcome *types.AggregationOutcome) (*DataFeedsOutcomeMetadata, error) {
	currentState := &DataFeedsOutcomeMetadata{}
	if previousOutcome != nil {
		err := proto.Unmarshal(previousOutcome.Metadata, currentState)
		if err != nil {
			return nil, err
		}
	}
	// initialize empty state for missing feeds
	if currentState.FeedInfo == nil {
		currentState.FeedInfo = make(map[string]*DataFeedsMercuryReportInfo)
	}
	for feedID := range a.config.Feeds {
		if _, ok := currentState.FeedInfo[feedID.String()]; !ok {
			currentState.FeedInfo[feedID.String()] = &DataFeedsMercuryReportInfo{
				ObservationTimestamp: 0, // will always trigger an update
				BenchmarkPrice:       big.NewInt(0).Bytes(),
			}
			lggr.Debugw("initializing empty onchain state for feed", "feedID", feedID.String())
		}
	}
	// remove obsolete feeds from state
	for feedID := range currentState.FeedInfo {
		if _, ok := a.config.Feeds[datastreams.FeedID(feedID)]; !ok {
			delete(currentState.FeedInfo, feedID)
			lggr.Debugw("removed obsolete feedID from state", "feedID", feedID)
		}
	}
	lggr.Debugw("current state initialized", "state", currentState, "previousOutcome", previousOutcome)
	return currentState, nil
}

func (a *dataFeedsAggregator) extractSignersAndPayloads(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value, fConsensus int) ([][]byte, int, map[ocrcommon.OracleID]values.Value) {
	events := make(map[ocrcommon.OracleID]values.Value)
	signers := make(map[[addrLen]byte]int)
	mins := make(map[int]int)
	for nodeID, nodeObservations := range observations {
		// we only expect a single observation per node - a Streams trigger event
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			lggr.Warnf("node %d contributed with empty observations", nodeID)
			continue
		}
		if len(nodeObservations) > 1 {
			lggr.Warnf("node %d contributed with more than one observation", nodeID)
			continue
		}
		triggerEvent := &datastreams.StreamsTriggerEvent{}
		if err := nodeObservations[0].UnwrapTo(triggerEvent); err != nil {
			lggr.Warnf("could not parse observations from node %d: %v", nodeID, err)
			continue
		}
		meta := triggerEvent.Metadata
		currentNodeSigners, err := extractUniqueSigners(meta.Signers)
		if err != nil {
			lggr.Warnf("could not extract signers from node %d: %v", nodeID, err)
			continue
		}
		for signer := range currentNodeSigners {
			signers[signer]++
		}
		mins[meta.MinRequiredSignatures]++
		events[nodeID] = nodeObservations[0]
	}
	// Agree on signers list and min-required. It's technically possible to have F+1 valid values from one trigger DON and F+1 from another trigger DON.
	// In that case both values are legitimate and signers list will contain nodes from both DONs. However, min-required value will be the higher one (if different).
	allowedSigners := [][]byte{}
	for signer, count := range signers {
		signer := signer
		if count >= fConsensus+1 {
			allowedSigners = append(allowedSigners, signer[:])
		}
	}
	minRequired := 0
	for minCandidate, count := range mins {
		if count >= fConsensus+1 && minCandidate > minRequired {
			minRequired = minCandidate
		}
	}
	return allowedSigners, minRequired, events
}

func extractUniqueSigners(signers [][]byte) (map[[addrLen]byte]struct{}, error) {
	uniqueSigners := make(map[[addrLen]byte]struct{})
	for _, signer := range signers {
		if len(signer) != addrLen {
			return nil, fmt.Errorf("invalid signer length: %d", len(signer))
		}
		var signerBytes [addrLen]byte
		copy(signerBytes[:], signer)
		uniqueSigners[signerBytes] = struct{}{}
	}
	return uniqueSigners, nil
}

func deviation(oldPrice, newPrice *big.Int) float64 {
	diff := &big.Int{}
	diff.Sub(oldPrice, newPrice)
	diff.Abs(diff)
	if oldPrice.Cmp(big.NewInt(0)) == 0 {
		if diff.Cmp(big.NewInt(0)) == 0 {
			return 0.0
		}
		return math.MaxFloat64
	}
	diffFl, _ := diff.Float64()
	oldFl, _ := oldPrice.Float64()
	return diffFl / oldFl
}

func NewDataFeedsAggregator(config values.Map, reportCodec datastreams.ReportCodec) (types.Aggregator, error) {
	parsedConfig, err := ParseConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &dataFeedsAggregator{
		config:      parsedConfig,
		reportCodec: reportCodec,
	}, nil
}

func ParseConfig(config values.Map) (aggregatorConfig, error) {
	parsedConfig := aggregatorConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return aggregatorConfig{}, err
	}

	for feedID, feedCfg := range parsedConfig.Feeds {
		if feedCfg.DeviationString != "" {
			if _, err := datastreams.NewFeedID(feedID.String()); err != nil {
				return aggregatorConfig{}, fmt.Errorf("cannot parse feedID config for feed %s: %w", feedID, err)
			}
			dec, err := decimal.NewFromString(feedCfg.DeviationString)
			if err != nil {
				return aggregatorConfig{}, fmt.Errorf("cannot parse deviation config for feed %s: %w", feedID, err)
			}
			feedCfg.Deviation = dec
			parsedConfig.Feeds[feedID] = feedCfg
		}
		trimmed, nonEmpty := strings.CutPrefix(feedCfg.RemappedIDHex, "0x")
		if nonEmpty {
			rawRemappedID, err := hex.DecodeString(trimmed)
			if err != nil {
				return aggregatorConfig{}, fmt.Errorf("cannot parse remappedId config for feed %s: %w", feedID, err)
			}
			feedCfg.RemappedID = rawRemappedID
			parsedConfig.Feeds[feedID] = feedCfg
		}
	}
	// convert allowedPartialStaleness from string to float64
	if parsedConfig.AllowedPartialStalenessStr != "" {
		allowedPartialStaleness, err := decimal.NewFromString(parsedConfig.AllowedPartialStalenessStr)
		if err != nil {
			return aggregatorConfig{}, fmt.Errorf("cannot parse allowedPartialStaleness: %w", err)
		}
		parsedConfig.AllowedPartialStaleness = allowedPartialStaleness.InexactFloat64()
	}
	return parsedConfig, nil
}
