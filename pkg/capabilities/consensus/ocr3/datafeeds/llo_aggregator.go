package datafeeds

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/shopspring/decimal"
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
)

// LLOAggregatorConfig is the config for the LLO aggregator.
// Example config:
// remappedID but a hex string
// streams:
//
//		"1":
//		  deviation: "0.1"
//		  heartbeat: 10
//	   remappedID: "0x680084f7347baFfb5C323c2982dfC90e04F9F918"
//		"2":
//		  deviation: "0.2"
//		  heartbeat: 20
//
// allowedPartialStaleness: "0.2"
// The streams are the stream IDs that the aggregator will aggregate.
type LLOAggregatorConfig struct {
	// workaround for the fact that mapstructure doesn't support uint32 keys
	//streams    map[uint32]feedConfig `mapstructure:"-"`
	Streams map[string]FeedConfig `mapstructure:"streams"`
	// allowedPartialStaleness is an optional optimization that tries to maximize batching.
	// Once any deviation or heartbeat threshold hits, we will include all other feeds that are
	// within the allowedPartialStaleness range of their own heartbeat.
	// For example, setting 0.2 will include all feeds that are within 20% of their heartbeat.
	//allowedPartialStaleness float64 `mapstructure:"-"`
	// workaround for the fact that mapstructure doesn't support float64 keys
	AllowedPartialStaleness string `mapstructure:"allowedPartialStaleness"`
}

// ToMap converts the LLOAggregatorConfig to a values.Map, which is suitable for the
// [NewAggegator] function in the OCR3 Aggregator interface.
func (c LLOAggregatorConfig) ToMap() (*values.Map, error) {
	v, err := values.WrapMap(c)
	if err != nil {
		// this should never happen since we are wrapping a struct
		return &values.Map{}, fmt.Errorf("failed to wrap LLOAggregatorConfig: %w", err)
	}
	return v, nil
}

func NewLLOConfig(m values.Map) (LLOAggregatorConfig, error) {
	// Create a default LLOAggregatorConfig
	config := LLOAggregatorConfig{
		Streams: make(map[string]FeedConfig),
	}
	if err := m.UnwrapTo(&config); err != nil {
		return LLOAggregatorConfig{}, fmt.Errorf("failed to unwrap values.Map to LLOAggregatorConfig: %w", err)
	}

	return config, nil
}

func (c LLOAggregatorConfig) convertToInternal() (parsedLLOAggregatorConfig, error) {
	parsedConfig := parsedLLOAggregatorConfig{
		streams: make(map[uint32]FeedConfig),
	}
	cfgErr := func(err error) error {
		cfgErr := fmt.Errorf("llo aggregator config: %w", ErrInvalidConfig)
		return errors.Join(cfgErr, err)
	}
	for s, cfg := range c.Streams {
		id, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			// this should never happen since we are using a mapstructure-compatible config
			return parsedConfig, cfgErr(fmt.Errorf("cannot parse stream ID %s: %w", s, err))
		}
		id32 := uint32(id) //nolint:gosec // G115
		parsedConfig.streams[id32] = cfg
	}
	// TODO some copy-pasta from feeds_aggregator.go - maybe reuse the same code?
	for streamID, cfg := range parsedConfig.streams {
		if cfg.RemappedIDHex == "" {
			return parsedConfig, cfgErr(fmt.Errorf("remappedID is required for stream %d", streamID))
		}
		if cfg.Deviation != "" {
			dec, err := decimal.NewFromString(cfg.Deviation)
			if err != nil {
				return parsedConfig, cfgErr(fmt.Errorf("cannot parse deviation config for feed %d: %w", streamID, err))
			}
			cfg.parsedDeviation = dec
			parsedConfig.streams[streamID] = cfg
		}
		trimmed := strings.TrimPrefix(cfg.RemappedIDHex, "0x")
		rawRemappedID, err := hex.DecodeString(trimmed)
		if err != nil {
			return parsedConfig, cfgErr(fmt.Errorf("cannot parse remappedId config for feed %d: %w", streamID, err))
		}
		cfg.remappedID = rawRemappedID
		parsedConfig.streams[streamID] = cfg
	}
	// convert allowedPartialStaleness from string to float64
	if c.AllowedPartialStaleness != "" {
		allowedPartialStaleness, err := decimal.NewFromString(c.AllowedPartialStaleness)
		if err != nil {
			return parsedConfig, cfgErr(fmt.Errorf("cannot parse allowedPartialStaleness: %w", err))
		}
		parsedConfig.allowedPartialStaleness = allowedPartialStaleness.InexactFloat64()
	}
	return parsedConfig, nil
}

// parsedLLOAggregatorConfig is the internal representation of the LLO aggregator config.
// the separation is because mapstructure only supports string keys.
// the are exposed in LLOAggregatorConfig for the config which is then processed into this internal representation.
type parsedLLOAggregatorConfig struct {
	streams                 map[uint32]FeedConfig
	allowedPartialStaleness float64
}

var _ types.Aggregator = (*LLOAggregator)(nil)

type LLOAggregator struct {
	config parsedLLOAggregatorConfig
}

// NewLLOAggregator creates a new LLOAggregator instance based on the provided configuration.
// The config should be a [values.Map] that has represents from the [LLOAggregatorConfig]. See [LLOAggreagatorConfig.ToMap]
func NewLLOAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := parseLLOConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &LLOAggregator{
		config: parsedConfig,
	}, nil
}

// Aggregate implements the Aggregator interface
// For this implementation, we expect the LLO events to be the same across all nodes.
// And we expect the every observation only contains a single LLO event, ie len(observations.["some-oracle-id"]) == 1.
func (a *LLOAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "LLOAggregator")
	if len(observations) == 0 {
		return nil, ErrEmptyObservation
	}
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
	slices.Sort(allStreamIDs)
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
		config := a.config.streams[streamID]
		oldPrice := new(decimal.Decimal)
		if uerr := oldPrice.UnmarshalBinary(previousStreamInfo.Price); uerr != nil {
			lggr.Errorw("failed to unmarshal previous price", "streamID", streamID, "err", uerr)
			continue
		}
		// If we don't have a price for this stream, we will use the previous price.
		// This is to prevent instantly zeroing out onchain values due to a bad LLO Stream Job/observation.
		newPrice, ok := prices[streamID]
		if !ok {
			newPrice = *oldPrice
			// Ensure we have a price for this stream in the prices map in case a heartbeat is triggered
			prices[streamID] = newPrice
			lggr.Debugw("using previous price for stream", "streamID", streamID, "newPrice", newPrice.String())
		}
		priceDeviation := deviationDecimal(*oldPrice, newPrice)
		timeDiffNs := observationTimestamp.UnixNano() - previousStreamInfo.Timestamp
		lggr.Debugw("checking deviation and heartbeat",
			"streamID", streamID,
			"observationNs", observationTimestamp,
			"perviousNs", previousStreamInfo.Timestamp,
			"currStalenessNs", timeDiffNs,
			"heartbeatSec", config.Heartbeat,
			"oldPrice", oldPrice,
			"newPrice", newPrice,
			"currDeviation", priceDeviation,
			"deviation", config.DeviationAsDecimal().InexactFloat64(),
		)
		if timeDiffNs > config.HeartbeatNanos() ||
			priceDeviation > config.DeviationAsDecimal().InexactFloat64() {
			// this stream needs an update
			previousStreamInfo.Timestamp = observationTimestamp.UnixNano()
			var err2 error
			previousStreamInfo.Price, err2 = prices[streamID].MarshalBinary()
			if err2 != nil {
				lggr.Errorw("failed to marshal price", "streamID", streamID, "err", err2)
				continue
			}
			mustUpdateIDs = append(mustUpdateIDs, streamID)
		} else if float64(timeDiffNs) > float64(config.HeartbeatNanos())*(1.0-a.config.allowedPartialStaleness) {
			maybeUpdateIDs = append(maybeUpdateIDs, streamID)
		}
	}

	// optimization that allows for more efficient batching
	// if there is at least one stream that actually hit its deviation or heartbeat threshold,
	// append all others that were within AllowedPartialStaleness percentage of their heartbeat
	if len(mustUpdateIDs) > 0 {
		mustUpdateIDs = append(mustUpdateIDs, maybeUpdateIDs...)
		// deterministic order
		slices.Sort(mustUpdateIDs)
	}

	marshalledState, err := proto.MarshalOptions{Deterministic: true}.Marshal(currentState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current state: %w", err)
	}

	toWrap := make([]*EVMEncodableStreamUpdate, 0, len(mustUpdateIDs))
	for _, streamID := range mustUpdateIDs {
		remappedID := a.config.streams[streamID].RemappedID()
		newPrice := prices[streamID]
		w := &EVMEncodableStreamUpdate{
			StreamID:   streamID,
			Price:      decimalToBigInt(newPrice),
			Timestamp:  uint32(observationTimestamp.Unix()), //nolint:gosec // G115
			RemappedID: remappedID,
		}
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

// EVMEncodableStreamUpdate is the EVM encodable representation of a stream update.
// The field name must match the field name in the EVM encoder, and must be a valid EVMEncoderKey.
type EVMEncodableStreamUpdate struct {
	StreamID   uint32
	Price      *big.Int
	Timestamp  uint32 // unix timestamp in seconds
	RemappedID []byte
}

func decimalToBigInt(d decimal.Decimal) *big.Int {
	return d.BigInt()
}

// extractLLOEvents decodes the untyped wire format into LLOStreamsTriggerEvent.
// every observation ios expected to be len 1, a single wrapped LLOStreamsTriggerEvent.
func (a *LLOAggregator) extractLLOEvents(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value) map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent {
	events := make(map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent)
	for nodeID, nodeObservations := range observations {
		lggr = logger.With(lggr, "nodeID", nodeID)
		// do not error on unexected number of observations
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
			lggr.Warnw("could not parse observations", "err", err)
			continue
		}
		events[nodeID] = triggerEvent
	}
	return events
}

// AggregationOutcome.Metadata is used to store extra data that is passed between OCR rounds as part of previous outcome.
// For LLO aggregator, that data is a serialized LLOOutcomeMetadata proto.
// This helper initializes current state by adjusting previous state with current config (adding missing streams, removing obsolete ones).
func (a *LLOAggregator) initializeLLOState(lggr logger.Logger, previousOutcome *types.AggregationOutcome) (*LLOOutcomeMetadata, error) {
	currentState := &LLOOutcomeMetadata{
		StreamInfo: make(map[uint32]*LLOStreamInfo),
	}
	if previousOutcome != nil && len(previousOutcome.Metadata) != 0 {
		err := proto.Unmarshal(previousOutcome.Metadata, currentState)
		if err != nil {
			return nil, err
		}
	}

	zero, err := decimal.Zero.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zero: %w", err)
	}
	for streamID := range a.config.streams {
		if _, ok := currentState.StreamInfo[streamID]; !ok {
			currentState.StreamInfo[streamID] = &LLOStreamInfo{
				Timestamp: 0, // trigger an update for every realistic heartbeat value
				Price:     zero,
			}
			lggr.Debugw("initializing empty stream info", "streamID", streamID)
		}
	}
	// remove obsolete streams from state
	for streamID := range currentState.StreamInfo {
		if _, ok := a.config.streams[streamID]; !ok {
			delete(currentState.StreamInfo, streamID)
			lggr.Debugw("removed obsolete stream", "streamID", streamID)
		}
	}
	return currentState, nil
}

// getObservationTimestamp returns the observation timestamp that appears at least f+1 times in the LLO events.
// it is optimistic and takes the first one that appears at least f+1 times. this is valid be we know that LLO events are coming from an OCR consensus output.
// ErrInsufficientConsensus is returned if no timestamp appears at least f+1 times.
func getObservationTimestamp(lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (time.Time, error) {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust the timestamp that appears at least f+1 times.
	counts := make(map[uint64]int)
	for _, event := range lloEvents {
		counts[event.ObservationTimestampNanoseconds]++
		if counts[event.ObservationTimestampNanoseconds] >= f+1 {
			return time.Unix(0, int64(event.ObservationTimestampNanoseconds)), nil //nolint:gosec // G115
		}
	}
	return time.Time{}, fmt.Errorf("%w: no timestamp appeared at least %d times", ErrInsufficientConsensus, f+1)
}

// lloStreamPrices returns the prices for the streams at the consensus observation timestamp.
// it ignores any events that are not from the consensus observation timestamp.
func lloStreamPrices(lggr logger.Logger, wantStreamIDs []uint32, lloEvents map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent, f int) (observationTimestamp time.Time, out map[uint32]decimal.Decimal, err error) {
	// All honest nodes are expected to include the same streams trigger event in their observation.
	// We can trust any price that appears at least f+1 times.

	out = make(map[uint32]decimal.Decimal)

	// Create a filter for the stream IDs we are interested in and initialize the candidate prices from which the output will be selected
	idFilter := make(map[uint32]struct{})
	candidatePrices := make(map[uint32]map[string]int) // streamID -> price -> count; string for price to avoid using decimal.Decimal as a map key
	for _, streamID := range wantStreamIDs {
		idFilter[streamID] = struct{}{}
		candidatePrices[streamID] = make(map[string]int)
	}

	// count all the prices across all events for the stream IDs we are interested in
	observationTimestamp, err = getObservationTimestamp(lloEvents, f)
	if err != nil {
		return time.Time{}, nil, err
	}
	observationTimestampNS := uint64(observationTimestamp.UnixNano()) //nolint:gosec // G115
	for _, event := range lloEvents {
		if event.ObservationTimestampNanoseconds != observationTimestampNS {
			// Ignore events with different timestamps. This shouldn't happen unless there are malicious nodes
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
				lggr.Errorw("failed to unmarshal decimal", "streamID", p.StreamID, "err", err)
				continue
			}
			// string key b/c decimal.Decimal is not comparable
			candidatePrices[p.StreamID][price.String()]++
		}
	}

	// find the price that appears at least f+1 times for each stream ID in the candidate prices
	for streamID, priceCount := range candidatePrices {
		found := false
		for priceStr, count := range priceCount {
			if count >= f+1 {
				// Convert the string back to decimal.Decimal
				price, err := decimal.NewFromString(priceStr)
				if err != nil {
					// this shouldn't happen since we just created the string from a decimal.Decimal
					lggr.Errorw("failed to parse price", "streamID", streamID, "priceStr", priceStr, "err", err)
				}
				out[streamID] = price
				found = true
				break
			}
		}
		if !found {
			lggr.Warnw("no price found in candidates with quorum", "streamID", streamID, "f", f, "candidates", priceCount, "err", ErrInsufficientConsensus)
		}
	}
	if len(out) != len(wantStreamIDs) {
		lggr.Warnw("not all streams have prices", "wantStreamIDs", len(wantStreamIDs), "out", len(out))
	}

	return observationTimestamp, out, nil
}

// parseLLOConfig parses the user-facing, type-less, LLO aggregator in the internal typed config.
func parseLLOConfig(config values.Map) (parsedLLOAggregatorConfig, error) {
	converter := LLOAggregatorConfig{
		Streams: make(map[string]FeedConfig),
	}
	if err := config.UnwrapTo(&converter); err != nil {
		return parsedLLOAggregatorConfig{}, err
	}
	x, err := converter.convertToInternal()
	if err != nil {
		return parsedLLOAggregatorConfig{}, err
	}
	return x, nil
}
