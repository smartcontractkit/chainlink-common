package datafeeds

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/shopspring/decimal"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// NOTE: in the future this could be defined per stream
// TODO where does this magic number come from? Presumably it's shared with some other code...
const multiplier = 1e18

var (
	ErrInsufficientConsensus = fmt.Errorf("insufficient consensus")
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

var _ types.Aggregator = (*lloAggregator)(nil)

type lloAggregator struct {
	config lloAggregatorConfig
}

func NewLLOAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := parseLLOConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &lloAggregator{
		config: parsedConfig,
	}, nil
}

// Aggregate implements the Aggregator interface,
func (a *lloAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "LLOAggregator")
	lloEvents := a.extractLLOEvents(lggr, observations)
	if len(lloEvents) != len(observations) {
		lggr.Warnw("missing LLO events", "nNodes", len(observations), "nEvents", len(lloEvents))
	}
	currentState, err := a.initializeLLOState(lggr, previousOutcome)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize llo state: %w", err)
	}

	allStreamIDs := []uint32{}
	for streamID := range currentState.StreamInfo {
		allStreamIDs = append(allStreamIDs, streamID)
	}
	lggr.Debugw("determined streams to aggregate", "nStreamIds", len(allStreamIDs))

	observationTimestamp, prices, err := lloStreamPrices(lggr, allStreamIDs, lloEvents, f)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest prices: %w", err)
	}
	lggr = logger.With(lggr, "observationTimestamp", observationTimestamp)

	mustUpdateIDs := []uint32{}  // streamIDs that need to be updated per deviation or heartbeat
	maybeUpdateIDs := []uint32{} // streamIDs that are within AllowedPartialStaleness percentage of their heartbeat
	for _, streamID := range allStreamIDs {
		previousStreamInfo := currentState.StreamInfo[streamID]
		config := a.config.Streams[streamID]
		oldPrice := big.NewInt(0).SetBytes(previousStreamInfo.Price)
		newPrice := prices[streamID].Mul(decimal.NewFromInt(multiplier)).BigInt()
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
			var err2 error
			previousStreamInfo.Price, err2 = prices[streamID].MarshalBinary()
			if err2 != nil {
				lggr.Errorw("failed to marshal price", "streamID", streamID, "err", err2)
				continue
			}
			mustUpdateIDs = append(mustUpdateIDs, streamID)
		} else if float64(currStaleness) > float64(config.Heartbeat)*(1.0-a.config.AllowedPartialStaleness) {
			maybeUpdateIDs = append(maybeUpdateIDs, streamID)
		}
	}

	// optimization that allows for more efficient batching
	// if there is at least one stream that actually hit its deviation or heartbeat threshold,
	// append all others that were withing AllowedPartialStaleness percentage of their heartbeat
	if len(mustUpdateIDs) > 0 {
		mustUpdateIDs = append(mustUpdateIDs, maybeUpdateIDs...)
		// deterministic order
		slices.Sort(mustUpdateIDs)
	}

	marshalledState, err := proto.MarshalOptions{Deterministic: true}.Marshal(currentState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current state: %w", err)
	}

	var toWrap []any
	for _, streamID := range mustUpdateIDs {
		// TODO what if remapped ID is not defined? How do we reconcile binary vs int? Should remapped IDs also be integers now?
		remappedID := a.config.Streams[streamID].RemappedID
		newPrice := prices[streamID].Mul(decimal.NewFromInt(multiplier)).BigInt()
		w := &wrappableUpdate{
			StreamID:   streamID,
			Price:      newPrice,
			Timestamp:  uint64(observationTimestamp),
			RemappedID: remappedID,
		} /*
			toWrap = append(toWrap,
				map[string]any{
					StreamIDOutputFieldName:   streamID,
					PriceOutputFieldName:      newPrice,
					TimestampOutputFieldName:  observationTimestamp, // TODO: nanoseconds OK here? What does the new onchain contract want? Should we make it configurable?
					RemappedIDOutputFieldName: remappedID,
				})
		*/
		toWrap = append(toWrap, w)
	}

	wrappedReportsNeedingUpdates, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, err
	}
	reportsProto := values.Proto(wrappedReportsNeedingUpdates)

	lggr.Debugw("Aggregate complete", "nStreamsNeedingUpdate", len(mustUpdateIDs))
	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         marshalledState,
		ShouldReport:     len(mustUpdateIDs) > 0,
	}, nil
}

type wrappableUpdate struct {
	StreamID   uint32
	Price      *big.Int
	Timestamp  uint64
	RemappedID []byte
}

func (w *wrappableUpdate) wrap() (map[string]any, error) {
	return map[string]any{
		StreamIDOutputFieldName:   w.StreamID,
		PriceOutputFieldName:      w.Price,
		TimestampOutputFieldName:  w.Timestamp,
		RemappedIDOutputFieldName: w.RemappedID,
	}, nil
}
func newwrappableUpdate(m map[string]any) *wrappableUpdate {
	return &wrappableUpdate{
		StreamID:   m[StreamIDOutputFieldName].(uint32),
		Price:      m[PriceOutputFieldName].(*big.Int),
		Timestamp:  m[TimestampOutputFieldName].(uint64),
		RemappedID: m[RemappedIDOutputFieldName].([]byte),
	}
}

// observations are expected to be wrapped LLOStreamsTriggerEvent structs
func (a *lloAggregator) extractLLOEvents(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value) map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent {
	events := make(map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent)
	for nodeID, nodeObservations := range observations {
		lggr = logger.With(lggr, "nodeID", nodeID)
		// we only expect a single observation per node - a Streams trigger event
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			lggr.Warn("empty observations")
			continue
		}
		if len(nodeObservations) > 1 {
			lggr.Warn("more than one observation")
			continue
		}
		triggerEvent := &datastreams.LLOStreamsTriggerEvent{}
		if err := nodeObservations[0].UnwrapTo(triggerEvent); err != nil {
			lggr.Warnw("could not parse observations", err)
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
	currentState := &LLOOutcomeMetadata{
		StreamInfo: make(map[uint32]*LLOStreamInfo),
	}
	if previousOutcome != nil {
		err := proto.Unmarshal(previousOutcome.Metadata, currentState)
		if err != nil {
			return nil, err
		}
	}

	zero, err := decimal.Zero.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zero: %w", err)
	}
	for streamID := range a.config.Streams {
		if _, ok := currentState.StreamInfo[streamID]; !ok {
			currentState.StreamInfo[streamID] = &LLOStreamInfo{
				Timestamp: 0, // will always trigger an update
				Price:     zero,
			}
			lggr.Debugw("initializing empty stream info", "streamID", streamID)
		}
	}
	// remove obsolete streams from state
	for streamID := range currentState.StreamInfo {
		if _, ok := a.config.Streams[streamID]; !ok {
			delete(currentState.StreamInfo, streamID)
			lggr.Debugw("removed obsolete stream", "streamID", streamID)
		}
	}
	return currentState, nil
}

func getObservationTimestamp(lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (uint64, error) {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust the timestamp that appears at least f+1 times.
	counts := make(map[uint64]int)
	for _, event := range lloEvents {
		counts[event.ObservationTimestampNanoseconds]++
		if counts[event.ObservationTimestampNanoseconds] >= f+1 {
			return event.ObservationTimestampNanoseconds, nil
		}
	}
	return 0, fmt.Errorf("%w: no timestamp appeared at least %d times", ErrInsufficientConsensus, f+1)
}

/*
	type lloObservation struct {
		lggr            logger.Logger
		ts              uint64
		streamPrice     map[uint32]decimal.Decimal
		state           *LLOOutcomeMetadata
		sortedStreamIDs []uint32
		events          map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent
		f               int
	}

	func newLLOObservation(lggr logger.Logger, allStreamIDs []uint32, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (*lloObservation, error) {
		// count all the prices across all events for the stream IDs we are interested in
		obTs, err := getObservationTimestamp(lloEvents, f)
		if err != nil {
			return nil, err
		}

		prices, err := priceAt(obTs, allStreamIDs, lloEvents, f)
		if err != nil {
			return nil, err
		}
		lggr = logger.With(lggr, "observationTimestamp", obTs)
		return &lloObservation{
			ts:          obTs,
			streamPrice: prices,
			lggr:        lggr,
		}, nil
	}

	func newLLOObservationx(lggr logger.Logger, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, state *LLOOutcomeMetadata, f int) (*lloObservation, error) {
		// count all the prices across all events for the stream IDs we are interested in
		obTs, err := getObservationTimestamp(lloEvents, f)
		if err != nil {
			return nil, err
		}
		ids := []uint32{}
		for streamID := range state.StreamInfo {
			ids = append(ids, streamID)
		}
		// ensure deterministic order
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		lggr.Debugw("determined streams to aggregate", "nStreamIds", len(ids))
		prices, err := priceAt(obTs, ids, lloEvents, f)
		if err != nil {
			return nil, err
		}
		lggr = logger.With(lggr, "observationTimestamp", obTs)
		return &lloObservation{
			ts:              obTs,
			streamPrice:     prices,
			state:           state,
			lggr:            lggr,
			sortedStreamIDs: ids,
			f:               f,
		}, nil
	}

	func (a *lloObservation) prices() (map[uint32]decimal.Decimal, error) {
		return priceAt(a.ts, a.sortedStreamIDs, a.events, a.f)
	}

	func priceAt(ts uint64, allStreamIDs []uint32, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (map[uint32]decimal.Decimal, error) {
		// All honest nodes are expected to include the same streams trigger event in their observation.
		// We can trust any price that appears at least f+1 times.
		// Observations can contain streamIDs that we don't need - filter them out.

		result := make(map[uint32]decimal.Decimal)

		// Create a filter for the stream IDs we are interested in and initialize the candidate prices
		idFilter := make(map[uint32]struct{})
		candidatePrices := make(map[uint32]map[string]int) // streamID -> price -> count; string for price to avoid using decimal.Decimal as a map key
		for _, streamID := range allStreamIDs {
			idFilter[streamID] = struct{}{}
			candidatePrices[streamID] = make(map[string]int)
		}

		// count all the prices across all events for the stream IDs we are interested in
		for _, event := range lloEvents {
			if event.ObservationTimestampNanoseconds != ts {
				continue
			}
			// Check if the event contains the stream ID we are interested in
			for _, p := range event.Payload {
				if _, ok := idFilter[p.StreamID]; !ok {
					continue
				}

				price := new(decimal.Decimal)
				if err := price.UnmarshalBinary(p.Decimal); err != nil {
					return nil, fmt.Errorf("failed to unmarshal binary: %w", err)
				}
				candidatePrices[p.StreamID][price.String()]++
			}
		}

		for streamID, priceCount := range candidatePrices {
			for priceStr, count := range priceCount {
				if count >= f+1 {
					price, err := decimal.NewFromString(priceStr)
					if err != nil {
						return nil, fmt.Errorf("failed to parse price %s for streamID %d: %w", priceStr, streamID, err)
					}
					result[streamID] = price
					break
				}
			}
		}

		return result, nil
	}
*/
func lloStreamPrices(lggr logger.Logger, allStreamIDs []uint32, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (observationTimestamp uint64, streamPrices map[uint32]decimal.Decimal, err error) {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust any price that appears at least f+1 times.
	// Observations can contain streamIDs that we don't need - filter them out.

	streamPrices = make(map[uint32]decimal.Decimal)

	// Create a filter for the stream IDs we are interested in and initialize the candidate prices
	idFilter := make(map[uint32]struct{})
	candidatePrices := make(map[uint32]map[string]int) // streamID -> price -> count; string for price to avoid using decimal.Decimal as a map key
	for _, streamID := range allStreamIDs {
		idFilter[streamID] = struct{}{}
		candidatePrices[streamID] = make(map[string]int)
	}

	// count all the prices across all events for the stream IDs we are interested in
	observationTimestamp, err = getObservationTimestamp(lloEvents, f)
	if err != nil {
		return 0, nil, err
	}
	for _, event := range lloEvents {
		if event.ObservationTimestampNanoseconds != observationTimestamp {
			// Ignore events with different timestamps
			// this really shouldn't happen unless there are malicious nodes
			// todo log warning
			lggr.Warnw("observation timestamp mismatch", "expected", observationTimestamp, "actual", event.ObservationTimestampNanoseconds)
			continue
		}
		// Check if the event contains the stream ID we are interested in
		for _, p := range event.Payload {
			if _, ok := idFilter[p.StreamID]; !ok {
				continue
			}

			// Convert the binary representation to decimal.Decimal
			price := new(decimal.Decimal)
			if err := price.UnmarshalBinary(p.Decimal); err != nil {
				// todo log error
				lggr.Errorw("failed to unmarshal decimal", "streamID", p.StreamID, "err", err)
				continue
			}
			candidatePrices[p.StreamID][price.String()]++
		}
	}

	// find the price that appears at least f+1 times for each stream ID
	for streamID, priceCount := range candidatePrices {
		// Check if any price appears at least f+1 times
		found := false
		for priceStr, count := range priceCount {
			if count >= f+1 {
				// Convert the string back to decimal.Decimal
				price, err := decimal.NewFromString(priceStr)
				if err != nil {
					// this shouldn't happen since we just created the string from a decimal.Decimal
					lggr.Errorw("failed to parse price", "streamID", streamID, "priceStr", priceStr, "err", err)
				}
				streamPrices[streamID] = price
				found = true
				break
			}
		}
		if !found {
			// todo log warning
			lggr.Warnw("no price found in candidates with quorum", "streamID", streamID, "f", f, "candidates", priceCount, "err", ErrInsufficientConsensus)
		}
	}

	return observationTimestamp, streamPrices, nil
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
				return lloAggregatorConfig{}, fmt.Errorf("cannot parse deviation config for feed %d: %w", streamID, err)
			}
			cfg.Deviation = dec
			parsedConfig.Streams[streamID] = cfg
		}
		trimmed, nonEmpty := strings.CutPrefix(cfg.RemappedIDHex, "0x")
		if nonEmpty {
			rawRemappedID, err := hex.DecodeString(trimmed)
			if err != nil {
				return lloAggregatorConfig{}, fmt.Errorf("cannot parse remappedId config for feed %d: %w", streamID, err)
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
