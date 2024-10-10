package median

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
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
)

type aggregatorConfig struct {
	ValueKey        string
	Deviation       decimal.Decimal    `mapstructure:"-"`
	DeviationString string             `mapstructure:"deviation"`
	FeedID          datastreams.FeedID `mapstructure:"feedId"`
	Heartbeat       int
	RemappedID      []byte `mapstructure:"-"`
	RemappedIDHex   string `mapstructure:"remappedId"`
}

type medianAggregator struct {
	clock       clockwork.Clock
	config      aggregatorConfig
	reportCodec datastreams.ReportCodec
	lggr        logger.Logger
}

var _ types.Aggregator = (*medianAggregator)(nil)

// EncodableOutcome is a list of aggregated price points.
// Metadata is a map of feedID -> (timestamp, price) representing onchain state (see MedianOutcomeMetadata proto)
func (a *medianAggregator) Aggregate(previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	events := a.extractPayloads(observations, a.config.ValueKey)
	a.lggr.Debugw("extracted values from observations", "nEvents", len(events), "f", f)
	hasMinObs := len(events) > f

	currentState, err := a.initializeCurrentState(previousOutcome)
	if err != nil {
		return nil, err
	}

	shouldReport := false
	toWrap := []any{}

	if hasMinObs {
		latestValue := median(events)
		feedID := string(a.config.FeedID)
		previousReportInfo := currentState.FeedInfo[feedID]
		oldPrice := big.NewInt(0).SetBytes(previousReportInfo.BenchmarkPrice)
		newPrice := big.NewInt(0).SetBytes(latestValue)
		currDeviation := deviation(oldPrice, newPrice)
		currTimestamp := time.Now().Unix()
		currStaleness := currTimestamp - previousReportInfo.ObservationTimestamp
		a.lggr.Debugw("checking deviation and heartbeat", "feedID", feedID, "currentTs", currTimestamp, "oldTs", previousReportInfo.ObservationTimestamp, "oldPrice", oldPrice, "newPrice", newPrice, "deviation", currDeviation)
		if currStaleness > int64(a.config.Heartbeat) ||
			currDeviation > a.config.Deviation.InexactFloat64() {
			previousReportInfo.ObservationTimestamp = currTimestamp
			previousReportInfo.BenchmarkPrice = newPrice.Bytes()

			feedIDBytes := a.config.FeedID.Bytes()
			remappedID := a.config.RemappedID
			if len(remappedID) == 0 { // if not provided, fall back to original ID
				remappedID = feedIDBytes[:]
			}
			toWrap = append(toWrap,
				map[string]any{
					FeedIDOutputFieldName:     feedIDBytes[:],
					RawReportOutputFieldName:  nil,
					PriceOutputFieldName:      newPrice,
					TimestampOutputFieldName:  currTimestamp,
					RemappedIDOutputFieldName: remappedID,
				})

			shouldReport = true
		}
	}

	marshalledState, err := proto.MarshalOptions{Deterministic: true}.Marshal(currentState)
	if err != nil {
		return nil, err
	}

	wrappedReportsNeedingUpdates, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, err
	}
	reportsProto := values.Proto(wrappedReportsNeedingUpdates)

	a.lggr.Debugw("Aggregate complete", "shouldReport", shouldReport)
	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         marshalledState,
		ShouldReport:     shouldReport,
	}, nil
}

func (a *medianAggregator) initializeCurrentState(previousOutcome *types.AggregationOutcome) (*MedianOutcomeMetadata, error) {
	currentState := &MedianOutcomeMetadata{}
	if previousOutcome != nil {
		err := proto.Unmarshal(previousOutcome.Metadata, currentState)
		if err != nil {
			return nil, err
		}
	}
	// initialize empty state for missing feeds
	if currentState.FeedInfo == nil {
		currentState.FeedInfo = make(map[string]*MedianReportInfo)
	}
	feedID := a.config.FeedID
	if _, ok := currentState.FeedInfo[feedID.String()]; !ok {
		currentState.FeedInfo[feedID.String()] = &MedianReportInfo{
			ObservationTimestamp: 0, // will always trigger an update
			BenchmarkPrice:       big.NewInt(0).Bytes(),
		}
		a.lggr.Debugw("initializing empty onchain state for feed", "feedID", feedID.String())
	}
	a.lggr.Debugw("current state initialized", "state", currentState, "previousOutcome", previousOutcome)
	return currentState, nil
}

func (a *medianAggregator) extractPayloads(observations map[ocrcommon.OracleID][]values.Value, aggregationKey string) [][]byte {
	var events [][]byte
	for nodeID, nodeObservations := range observations {
		// TODO: check the following comment:
		// we only expect a single observation per node
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			a.lggr.Warnf("node %d contributed with empty observations", nodeID)
			continue
		}
		if len(nodeObservations) > 1 {
			a.lggr.Warnf("node %d contributed with more than one observation", nodeID)
			continue
		}
		payload := &capabilities.CapabilityResponse{}
		if err := nodeObservations[0].UnwrapTo(payload); err != nil {
			a.lggr.Warnf("could not parse observations as capability response from node %d: %v", nodeID, err)
			continue
		}
		value, exists := payload.Value.Underlying[aggregationKey]
		if !exists {
			a.lggr.Warnf("no key %s found on observation payload from node %d", aggregationKey, nodeID)
			continue
		}
		var valueBytes []byte
		if err := value.UnwrapTo(&valueBytes); err != nil {
			a.lggr.Warnf("could not parse capability response as bytes from node %d: %v", nodeID, err)
			continue
		}
		events = append(events, valueBytes)
	}
	return events
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

func median(items [][]byte) []byte {
	sort.Slice(items, func(i, j int) bool {
		if len(items[i]) != len(items[j]) {
			// NOTE: this doesn't account for extra leading zeros
			return len(items[i]) < len(items[j])
		}
		return bytes.Compare(items[i], items[j]) < 0
	})
	return items[(len(items)-1)/2]
}

func NewMedianAggregator(config values.Map, reportCodec datastreams.ReportCodec, lggr logger.Logger, clock clockwork.Clock) (types.Aggregator, error) {
	parsedConfig, err := ParseConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &medianAggregator{
		clock:       clock,
		config:      parsedConfig,
		reportCodec: reportCodec,
		lggr:        logger.Named(lggr, "DataFeedsAggregator"),
	}, nil
}

func ParseConfig(config values.Map) (aggregatorConfig, error) {
	parsedConfig := aggregatorConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return aggregatorConfig{}, err
	}

	feedID := parsedConfig.FeedID

	if parsedConfig.ValueKey == "" {
		return aggregatorConfig{}, fmt.Errorf("must provide valueKey config for feed %s", feedID)
	}

	if parsedConfig.DeviationString != "" {
		if _, err := datastreams.NewFeedID(feedID.String()); err != nil {
			return aggregatorConfig{}, fmt.Errorf("cannot parse feedID config for feed %s: %w", feedID, err)
		}
		dec, err := decimal.NewFromString(parsedConfig.DeviationString)
		if err != nil {
			return aggregatorConfig{}, fmt.Errorf("cannot parse deviation config for feed %s: %w", feedID, err)
		}
		parsedConfig.Deviation = dec
	}

	trimmed, nonEmpty := strings.CutPrefix(parsedConfig.RemappedIDHex, "0x")
	if nonEmpty {
		rawRemappedID, err := hex.DecodeString(trimmed)
		if err != nil {
			return aggregatorConfig{}, fmt.Errorf("cannot parse remappedId config for feed %s: %w", feedID, err)
		}
		parsedConfig.RemappedID = rawRemappedID
	}

	return parsedConfig, nil
}
