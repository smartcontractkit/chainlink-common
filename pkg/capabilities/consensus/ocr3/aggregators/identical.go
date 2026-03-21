package aggregators

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// Aggregates by the most frequent observation for each index of a data set
type identicalAggregator struct {
	config IdenticalAggConfig
}

type IdenticalAggConfig struct {
	// Length of the list of observations that each node is expected to provide.
	// Aggregator's output (i.e. EncodableOutcome) will be a values.Map with the same
	// number of elements and keyed by indices 0,1,2,... (unless KeyOverrides are provided).
	// Defaults to 1.
	ExpectedObservationsLen int
	// If non-empty, the keys in the outcome map will be replaced with these values.
	// If non-empty, must be of length ExpectedObservationsLen.
	KeyOverrides []string
}

type counter struct {
	fullObservation values.Value
	count           int
}

var _ types.Aggregator = (*identicalAggregator)(nil)

func (a *identicalAggregator) Aggregate(lggr logger.Logger, _ *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	counters := []map[[32]byte]*counter{}
	for i := 0; i < a.config.ExpectedObservationsLen; i++ {
		counters = append(counters, map[[32]byte]*counter{})
	}
	for nodeID, nodeObservations := range observations {
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			lggr.Warnf("node %d contributed with empty observations", nodeID)
			continue
		}
		if len(nodeObservations) != a.config.ExpectedObservationsLen {
			lggr.Warnf("node %d contributed with an incorrect number of observations %d - ignoring them", nodeID, len(nodeObservations))
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
		var highestSHA [32]byte
		var highestObservation values.Value
		for sha, counter := range shaToCounter {
			if counter.count > highestCount ||
				(counter.count == highestCount && bytes.Compare(sha[:], highestSHA[:]) < 0) {
				highestCount = counter.count
				highestSHA = sha
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

func NewIdenticalAggregator(config values.Map) (*identicalAggregator, error) {
	parsedConfig, err := ParseConfigIdenticalAggregator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &identicalAggregator{
		config: parsedConfig,
	}, nil
}

func ParseConfigIdenticalAggregator(config values.Map) (IdenticalAggConfig, error) {
	parsedConfig := IdenticalAggConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return IdenticalAggConfig{}, err
	}
	if parsedConfig.ExpectedObservationsLen == 0 {
		parsedConfig.ExpectedObservationsLen = 1
	}
	return parsedConfig, nil
}
