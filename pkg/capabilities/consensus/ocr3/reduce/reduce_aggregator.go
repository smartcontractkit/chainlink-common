package reduce

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type AggregatorConfig struct {
	// Static data that is added into every report
	AdditionalData map[string]any `mapstructure:"additionalData"`
	// Configuration on how to aggregate one or more data points
	Fields []AggregationField `mapstructure:"fields"`
	// The top level field name that report data is put into
	OutputFieldName string `mapstructure:"outputFieldName" json:"outputFieldName" default:"Reports"` // TODO: is this actually needed?
	// The structure surrounding the report data that is put on to "OutputFieldName"
	ReportFormat string `mapstructure:"reportFormat" json:"reportFormat" default:"map" jsonschema:"enum=map,enum=array"`
}

type AggregationField struct {
	// The key to find a data point within the input data
	InputKey string `mapstructure:"inputKey" json:"inputKey" required:"true"`
	// The key that the aggregated data is put under
	OutputKey string `mapstructure:"outputKey" json:"outputKey" required:"true"`
	// How the data set should be aggregated
	Method string `mapstructure:"method" json:"method" jsonschema:"enum=median,enum=mode" required:"true"`
	// An optional check to only report when the difference from the previous report exceeds a certain threshold.
	// If not provided, there will always be a report once minimum observations are reached.
	DeviationString string          `mapstructure:"deviation"` // TODO: omit empty
	Deviation       decimal.Decimal `mapstructure:"-"`         // TODO: omit empty
	// The format of the deviation being provided
	// * percent - a percentage deviation
	// * absolute - a numeric difference
	DeviationType string `mapstructure:"deviationType" json:"deviationType" jsonschema:"enum=percent,enum=absolute"` // TODO: omit empty
}

type reduceAggregator struct {
	config AggregatorConfig
	lggr   logger.Logger
}

var _ types.Aggregator = (*reduceAggregator)(nil)

func (a *reduceAggregator) Aggregate(previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	currentState, err := a.initializeCurrentState(previousOutcome)
	if err != nil {
		return nil, err
	}

	report := map[string]any{}
	shouldReport := false
	if len(observations) > 2*f {
		for _, field := range a.config.Fields {
			values := a.extractValues(observations, field.InputKey)
			a.lggr.Debugw("extracted values from observations", "nVals", len(values), "f", f)

			// only report if every field has reached the minimum number of observations
			if len(values) < 2*f+1 {
				shouldReport = false
				break
			}

			singleValue, err := reduce(field.Method, values)
			if err != nil {
				a.lggr.Debugw("unable to reduce", "err", err.Error(), "method", field.Method)
			}

			oldValue := (*currentState)[field.InputKey]

			currDeviation, err := deviation(oldValue, singleValue)
			if oldValue != nil && err != nil {
				a.lggr.Debugw("unable to determine deviation", "err", err.Error())
			}

			(*currentState)[field.InputKey] = singleValue
			report[field.OutputKey] = singleValue

			if oldValue == nil || currDeviation > field.Deviation.InexactFloat64() {
				shouldReport = true
			}
		}
	}

	// Add static additonal data to the report
	for key, value := range a.config.AdditionalData {
		report[key] = value
	}

	// TODO: name below better ----
	vm, err := values.WrapMap(currentState)
	if err != nil {
		return nil, err
	}
	marshalledState := values.ProtoMap(vm)
	metadata, err := proto.Marshal(marshalledState)
	if err != nil {
		return nil, err
	}

	var toWrap any
	switch a.config.ReportFormat {
	case "array":
		toWrap = []map[string]any{report}
	case "map":
		toWrap = report
	default:
		// unsupported type
		toWrap = map[string]any{}
	}

	wrappedReportsNeedingUpdates, err := values.NewMap(map[string]any{
		a.config.OutputFieldName: toWrap,
	})
	if err != nil {
		return nil, err
	}
	reportsProto := values.Proto(wrappedReportsNeedingUpdates)

	// ----

	a.lggr.Debugw("Aggregate complete", "shouldReport", shouldReport)
	return &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         metadata,
		ShouldReport:     shouldReport,
	}, nil
}

func (a *reduceAggregator) initializeCurrentState(previousOutcome *types.AggregationOutcome) (*map[string]any, error) {
	currentState := map[string]any{}
	if previousOutcome != nil {
		pb := &pb.Map{}
		proto.Unmarshal(previousOutcome.Metadata, pb)
		mv, err := values.FromMapValueProto(pb)
		if err != nil {
			return nil, err
		}
		mv.UnwrapTo(currentState)
		if err != nil {
			return nil, err
		}
	}
	for _, field := range a.config.Fields {
		if _, ok := currentState[field.InputKey]; !ok {
			// TODO: set default type. user defined type?
			currentState[field.InputKey] = nil
			a.lggr.Debugw("initializing empty onchain state for feed", "fieldInputKey", field.InputKey)
		}
	}
	a.lggr.Debugw("current state initialized", "state", currentState, "previousOutcome", previousOutcome)
	return &currentState, nil
}

func (a *reduceAggregator) extractValues(observations map[ocrcommon.OracleID][]values.Value, aggregationKey string) []any {
	events := map[string][]any{}
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
		vmap, err := nodeObservations[0].Unwrap()
		if err != nil {
			a.lggr.Warnf("could not unwrap observation value from node %d: %v", nodeID, err)
			continue
		}

		value := vmap.(map[string]any)[aggregationKey]

		switch v := value.(type) {
		case string:
			events["string"] = append(events["string"], value)
		case decimal.Decimal:
			events["decimal"] = append(events["decimal"], value)
		case int64:
			events["int64"] = append(events["int64"], value)
		case []byte:
			events["bytes"] = append(events["bytes"], value)
		case big.Int:
			events["bigint"] = append(events["bigint"], value)
		case time.Time:
			events["time"] = append(events["time"], value)
		case float64:
			events["float64"] = append(events["float64"], value)
		default:
			fmt.Printf("I don't know about type to extract values %T!\n", v)
			// unsupported type
		}
	}

	// it's possible that values are not all of the same type
	// they could either be coerced to the same type, or only use observations from f+1 of the same type
	// for now, skip if mixed types TODO: handle this better
	if len(events) > 1 {
		fmt.Println("multiple types seen")
		a.lggr.Warnf("multiple %s", "remove")
		return []any{}
	}

	for _, typedEvents := range events {
		return typedEvents
	}

	return []any{}
}

func reduce(method string, items []any) (any, error) {
	switch method {
	case "median":
		return median(items)
	case "mode":
		return mode(items)
	default:
		// unsupported method
		return nil, errors.New("unsupported aggregation method")
	}
}

func median(items []any) (any, error) {
	if len(items) == 0 {
		return nil, errors.New("items cannot be empty")
	}
	sortAny(items)
	return items[(len(items)-1)/2], nil
}

func sortAny(items []any) error {
	switch v := items[0].(type) {
	case string:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(string) < items[j].(string)
		})
		return nil
	case decimal.Decimal:
		// TODO
		return nil
	case int64:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(int64) < items[j].(int64)
		})
		return nil
	case []byte:
		// TODO
		return nil
	case big.Int:
		// TODO
		return nil
	case time.Time:
		// TODO
		return nil
	case float64:
		sort.Slice(items, func(i, j int) bool {
			return items[i].(float64) < items[j].(float64)
		})
		return nil
	default:
		return fmt.Errorf("I don't know about type %T!\n", v)
		// unsupported type
	}
}

func mode(items []any) (any, error) {
	if len(items) == 0 {
		return nil, errors.New("items cannot be empty")
	}

	counts := make(map[any]int)
	for _, item := range items {
		counts[item]++
	}

	var maxCount int
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}

	var modes []any
	for item, count := range counts {
		if count == maxCount {
			modes = append(modes, item)
		}
	}

	// If more than one mode found, choose first

	return modes[0], nil
}

func deviation(oldValue, newValue any) (float64, error) {
	prevType := reflect.TypeOf(oldValue)
	newType := reflect.TypeOf(newValue)

	if prevType != newType {
		return 0, fmt.Errorf("deviation type mismatch: old value %s, new value %s", prevType, newType)
	}

	// TODO: use deviation method, e.g. percent vs. absolute

	switch v := oldValue.(type) {
	case string:
		// TODO
		return 0, nil
	case decimal.Decimal:
		// TODO
		return 0, nil
	case int64:
		bigOld := big.NewInt(0).SetInt64(oldValue.(int64))
		bigNew := big.NewInt(0).SetInt64(newValue.(int64))
		diff := &big.Int{}
		diff.Sub(bigOld, bigNew)
		diff.Abs(diff)
		if bigOld.Cmp(big.NewInt(0)) == 0 {
			if diff.Cmp(big.NewInt(0)) == 0 {
				return 0.0, nil
			}
			return math.MaxFloat64, nil
		}
		diffFl, _ := diff.Float64()
		oldFl, _ := bigOld.Float64()
		return diffFl / oldFl, nil
	case []byte:
		// TODO
		return 0, nil
	case big.Int:
		bigOld := oldValue.(big.Int)
		bigNew := newValue.(big.Int)
		diff := &big.Int{}
		diff.Sub(&bigOld, &bigNew)
		diff.Abs(diff)
		if bigOld.Cmp(big.NewInt(0)) == 0 {
			if diff.Cmp(big.NewInt(0)) == 0 {
				return 0.0, nil
			}
			return math.MaxFloat64, nil
		}
		diffFl, _ := diff.Float64()
		oldFl, _ := bigOld.Float64()
		return diffFl / oldFl, nil
	case time.Time:
		// TODO
		return 0, nil
	case float64:
		// TODO
		return 0, nil
	default:
		// unsupported type
		return 0, fmt.Errorf("deviation doesn't know about type %T", v)
	}
}

func NewReduceAggregator(config values.Map, lggr logger.Logger) (types.Aggregator, error) {
	parsedConfig, err := ParseConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &reduceAggregator{
		config: parsedConfig,
		lggr:   logger.Named(lggr, "MedianAggregator"),
	}, nil
}

func ParseConfig(config values.Map) (AggregatorConfig, error) {
	parsedConfig := AggregatorConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return AggregatorConfig{}, err
	}

	// TODO: validations

	return parsedConfig, nil
}
