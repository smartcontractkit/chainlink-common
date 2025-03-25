package datafeeds

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/shopspring/decimal"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type lloAggregatorConfig struct {
	Streams map[uint32]feedConfig
	// AllowedPartialStaleness is an optional optimization that tries to maximize batching.
	// Once any deviation or heartbeat threshold hits, we will include all other feeds that are
	// within the AllowedPartialStaleness range of their own heartbeat.
	// For example, setting 0.2 will include all feeds that are within 20% of their heartbeat.
	AllowedPartialStaleness    float64 `mapstructure:"-"`
	AllowedPartialStalenessStr string  `mapstructure:"allowedPartialStaleness"`
}

// NOTE: in the future this could be defined per stream
const multiplier = 1e18

type lloAggregator struct {
	config lloAggregatorConfig
}

var _ types.Aggregator = (*lloAggregator)(nil)

func (a *lloAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lloEvents := a.extractLLOEvents(lggr, observations)

	currentState, err := a.initializeLLOState(lggr, previousOutcome)
	if err != nil {
		return nil, err
	}

	allStreamIDs := []uint32{}
	for streamID := range currentState.StreamInfo {
		allStreamIDs = append(allStreamIDs, streamID)
	}
	// ensure deterministic order
	sort.Slice(allStreamIDs, func(i, j int) bool { return allStreamIDs[i] < allStreamIDs[j] })
	lggr.Debugw("determined streams to aggregate", "nStreamIds", len(allStreamIDs))

	// in LLO, all streams belong to the same channel so they share the timestamp
	observationTimestamp, err := getObservationTimestamp(lloEvents, f)
	if err != nil {
		return nil, err
	}
	latestPrices := getLatestPrices(allStreamIDs, lloEvents, f)

	streamsNeedingUpdate := []uint32{}
	candidateIDs := []uint32{}
	for _, streamID := range allStreamIDs {
		previousStreamInfo := currentState.StreamInfo[streamID]
		config := a.config.Streams[streamID]
		oldPrice := big.NewInt(0).SetBytes(previousStreamInfo.Price)
		newPrice := latestPrices[streamID].Mul(decimal.NewFromInt(multiplier)).BigInt()
		currDeviation := deviation(oldPrice, newPrice)
		currStaleness := observationTimestamp - uint64(previousStreamInfo.Timestamp)
		lggr.Debugw("checking deviation and heartbeat",
			"streamID", streamID,
			"currentTs", observationTimestamp,
			"oldTs", previousStreamInfo.Timestamp,
			"currStaleness", currStaleness,
			"heartbeat", config.Heartbeat,
			"oldPrice", oldPrice,
			"newPrice", newPrice,
			"currDeviation", currDeviation,
			"deviation", config.Deviation.InexactFloat64(),
		)
		if currStaleness > uint64(config.Heartbeat) ||
			currDeviation > config.Deviation.InexactFloat64() {
			// this stream needs an update
			previousStreamInfo.Timestamp = int64(observationTimestamp)
			previousStreamInfo.Price, _ = latestPrices[streamID].MarshalBinary()
			streamsNeedingUpdate = append(streamsNeedingUpdate, streamID)
		} else if float64(currStaleness) > float64(config.Heartbeat)*(1.0-a.config.AllowedPartialStaleness) {
			candidateIDs = append(candidateIDs, streamID)
		}
	}

	// optimization that allows for more efficient batching
	// if there is at least one stream that actually hit its deviation or heartbeat threshold,
	// append all others that were withing AllowedPartialStaleness percentage of their heartbeat
	if len(streamsNeedingUpdate) > 0 {
		streamsNeedingUpdate = append(streamsNeedingUpdate, candidateIDs...)
	}

	marshalledState, err := proto.MarshalOptions{Deterministic: true}.Marshal(currentState)
	if err != nil {
		return nil, err
	}

	var toWrap []any
	for _, streamID := range streamsNeedingUpdate {
		// TODO what if remapped ID is not defined? How do we reconcile binary vs int? Should remapped IDs also be integers now?
		remappedID := a.config.Streams[streamID].RemappedID
		newPrice := latestPrices[streamID].Mul(decimal.NewFromInt(multiplier)).BigInt()
		toWrap = append(toWrap,
			map[string]any{
				StreamIDOutputFieldName:   streamID,
				PriceOutputFieldName:      newPrice,
				TimestampOutputFieldName:  observationTimestamp, // TODO: nanoseconds OK here? What does the new onchain contract want? Should we make it configurable?
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

	lggr.Debugw("Aggregate complete", "nStreamsNeedingUpdate", len(streamsNeedingUpdate))
	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         marshalledState,
		ShouldReport:     len(streamsNeedingUpdate) > 0,
	}, nil
}

// observations are expected to be wrapped LLOStreamsTriggerEvent structs
func (a *lloAggregator) extractLLOEvents(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value) map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent {
	events := make(map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent)
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
		triggerEvent := &datastreams.LLOStreamsTriggerEvent{}
		if err := nodeObservations[0].UnwrapTo(triggerEvent); err != nil {
			lggr.Warnf("could not parse observations from node %d: %v", nodeID, err)
			continue
		}
		events[nodeID] = triggerEvent
	}
	return events
}

// AggregationOutcome.Metadata is used to store extra data that is passed between OCR rounds as part of previous outcome.
// For LLO aggregator, that data is a serialized LLOOutcomeMetadata proto.
// This helper initializes current state by adjusting previous state with current config (adding missing streams, removing obsolete ones).
func (a *lloAggregator) initializeLLOState(lggr logger.Logger, previousOutcome *types.AggregationOutcome) (*LLOOutcomeMetadata, error) {
	currentState := &LLOOutcomeMetadata{}
	if previousOutcome != nil {
		err := proto.Unmarshal(previousOutcome.Metadata, currentState)
		if err != nil {
			return nil, err
		}
	}
	// initialize empty state for missing streams
	if currentState.StreamInfo == nil {
		currentState.StreamInfo = make(map[uint32]*LLOStreamInfo)
	}
	zero, _ := decimal.Zero.MarshalBinary()
	for streamID := range a.config.Streams {
		if _, ok := currentState.StreamInfo[streamID]; !ok {
			currentState.StreamInfo[streamID] = &LLOStreamInfo{
				Timestamp: 0, // will always trigger an update
				Price:     zero,
			}
			lggr.Debugw("initializing empty onchain state for stream", "streamID", streamID)
		}
	}
	// remove obsolete streams from state
	for streamID := range currentState.StreamInfo {
		if _, ok := a.config.Streams[streamID]; !ok {
			delete(currentState.StreamInfo, streamID)
			lggr.Debugw("removed obsolete stream from state", "streamID", streamID)
		}
	}
	lggr.Debugw("current state initialized", "nStreams", len(currentState.StreamInfo))
	return currentState, nil
}

func getObservationTimestamp(lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (uint64, error) {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust the timestamp that apprears at least f+1 times.
	counts := make(map[uint64]int)
	for _, event := range lloEvents {
		counts[event.ObservationTimestampNanoseconds]++
		if counts[event.ObservationTimestampNanoseconds] >= f+1 {
			return event.ObservationTimestampNanoseconds, nil
		}
	}
	return 0, fmt.Errorf("no timestamp appeared at least %d times", f+1)
}

func getLatestPrices(allStreamIDs []uint32, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) map[uint32]*decimal.Decimal {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust any price that apprears at least f+1 times.
	// Observations can contain streamIDs that we don't need - filter them out.

	// TODO implement similar logic to getObservationTimestamp but per streamID (only for those that are in allStreamIDs)
	return map[uint32]*decimal.Decimal{}
}

// TODO: add it to core/capabilities/aggregator_factory.go
func NewLLOAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := parseLLOConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &lloAggregator{
		config: parsedConfig,
	}, nil
}

func parseLLOConfig(config values.Map) (lloAggregatorConfig, error) {
	parsedConfig := lloAggregatorConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return lloAggregatorConfig{}, err
	}

	// TODO some copy-pasta from feeds_aggregator.go - maybe reuse the same code?
	for streamID, cfg := range parsedConfig.Streams {
		if cfg.DeviationString != "" {
			dec, err := decimal.NewFromString(cfg.DeviationString)
			if err != nil {
				return lloAggregatorConfig{}, fmt.Errorf("cannot parse deviation config for feed %s: %w", streamID, err)
			}
			cfg.Deviation = dec
			parsedConfig.Streams[streamID] = cfg
		}
		trimmed, nonEmpty := strings.CutPrefix(cfg.RemappedIDHex, "0x")
		if nonEmpty {
			rawRemappedID, err := hex.DecodeString(trimmed)
			if err != nil {
				return lloAggregatorConfig{}, fmt.Errorf("cannot parse remappedId config for feed %s: %w", streamID, err)
			}
			cfg.RemappedID = rawRemappedID
			parsedConfig.Streams[streamID] = cfg
		}
	}
	// convert allowedPartialStaleness from string to float64
	if parsedConfig.AllowedPartialStalenessStr != "" {
		allowedPartialStaleness, err := decimal.NewFromString(parsedConfig.AllowedPartialStalenessStr)
		if err != nil {
			return lloAggregatorConfig{}, fmt.Errorf("cannot parse allowedPartialStaleness: %w", err)
		}
		parsedConfig.AllowedPartialStaleness = allowedPartialStaleness.InexactFloat64()
	}
	return parsedConfig, nil
}
