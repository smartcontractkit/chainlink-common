package aggregators

import (
	"crypto/sha256"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
)

type identicalAggregator struct {
	config aggregatorConfig
	lggr   logger.Logger
}

type aggregatorConfig struct {
	// Length of the list of observations that each node is expected to provide.
	// Aggregator's output (i.e. EncodableOutcome) will be a values.Map with the same
	// number of elements and keyed by indices 0,1,2,... (unless KeyOverrides are provided).
	ExpectedObservationsLen int
	// If non-empty, the keys in the outcome map will be replaced with these values.
	// Must be of length ExpectedObservationsLen.
	KeyOverrides []string
}

type counter struct {
	fullObservation values.Value
	count           int
}

var _ types.Aggregator = (*identicalAggregator)(nil)

func (a *identicalAggregator) Aggregate(_ *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	counters := []map[[32]byte]*counter{}
	for i := 0; i < a.config.ExpectedObservationsLen; i++ {
		counters = append(counters, map[[32]byte]*counter{})
	}
	for nodeID, nodeObservations := range observations {
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			a.lggr.Warnf("node %d contributed with empty observations", nodeID)
			continue
		}
		if len(nodeObservations) != a.config.ExpectedObservationsLen {
			a.lggr.Warnf("node %d contributed with an incorrect number of observations %d - ignoring them", nodeID, len(nodeObservations))
			continue
		}
		for idx, observation := range nodeObservations {
			marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(observation))
			if err != nil {
				return nil, err
			}
			sha := sha256.Sum256(marshalled)
			elem, ok := counters[idx][sha]
			if !ok {
				counters[idx][sha] = &counter{
					fullObservation: observation,
					count:           1,
				}
			} else {
				elem.count++
			}
		}
	}
	return a.collectHighestCounts(counters, f)
}

func (a *identicalAggregator) collectHighestCounts(counters []map[[32]byte]*counter, f int) (*types.AggregationOutcome, error) {
	useOverrides := len(a.config.KeyOverrides) == len(counters)
	outcome := make(map[string]any)
	for idx, shaToCounter := range counters {
		highestCount := 0
		var highestObservation values.Value
		for _, counter := range shaToCounter {
			if counter.count > highestCount {
				highestCount = counter.count
				highestObservation = counter.fullObservation
			}
		}
		if highestCount < 2*f+1 {
			return nil, fmt.Errorf("can't reach consensus on observations with index %d", idx)
		}
		if useOverrides {
			outcome[a.config.KeyOverrides[idx]] = highestObservation
		} else {
			outcome[fmt.Sprintf("%d", idx)] = highestObservation
		}
	}
	valMap, err := values.NewMap(outcome)
	if err != nil {
		return nil, err
	}
	return &types.AggregationOutcome{
		EncodableOutcome: values.ProtoMap(valMap),
		Metadata:         nil,
		ShouldReport:     true,
	}, nil
}

func NewIdenticalAggregator(config values.Map, lggr logger.Logger) (*identicalAggregator, error) {
	parsedConfig, err := ParseConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &identicalAggregator{
		config: parsedConfig,
		lggr:   lggr,
	}, nil
}

func ParseConfig(config values.Map) (aggregatorConfig, error) {
	parsedConfig := aggregatorConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return aggregatorConfig{}, err
	}
	return parsedConfig, nil
}
