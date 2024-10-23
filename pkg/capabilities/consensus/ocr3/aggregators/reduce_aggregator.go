package aggregators

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

const (
	AGGREGATION_METHOD_MEDIAN = "median"
	AGGREGATION_METHOD_MODE   = "mode"
	DEVIATION_TYPE_NONE       = "none"
	DEVIATION_TYPE_PERCENT    = "percent"
	DEVIATION_TYPE_ABSOLUTE   = "absolute"
	REPORT_FORMAT_MAP         = "map"
	REPORT_FORMAT_ARRAY       = "array"

	DEFAULT_REPORT_FORMAT     = REPORT_FORMAT_MAP
	DEFAULT_OUTPUT_FIELD_NAME = "Reports"
)

type ReduceAggConfig struct {
	// Configuration on how to aggregate one or more data points
	Fields []AggregationField `mapstructure:"fields"  required:"true"`
	// The top level field name that report data is put into
	OutputFieldName string `mapstructure:"outputFieldName" json:"outputFieldName" default:"Reports"`
	// The structure surrounding the report data that is put on to "OutputFieldName"
	ReportFormat string `mapstructure:"reportFormat" json:"reportFormat" default:"map" jsonschema:"enum=map,enum=array"`
}

type AggregationField struct {
	// The key to find a data point within the input data
	// If omitted, the entire input will be used
	InputKey string `mapstructure:"inputKey" json:"inputKey"`
	// The key that the aggregated data is put under
	// If omitted, the InputKey will be used
	OutputKey string `mapstructure:"outputKey" json:"outputKey"`
	// How the data set should be aggregated
	Method string `mapstructure:"method" json:"method" jsonschema:"enum=median,enum=mode" required:"true"`
	// An optional check to only report when the difference from the previous report exceeds a certain threshold.
	// Can only be used when the field is of a numeric type: string, decimal, int64, big.Int, time.Time, float64
	// If no deviation is provided on any field, there will always be a report once minimum observations are reached.
	DeviationString string          `mapstructure:"deviation"  json:"deviation,omitempty"`
	Deviation       decimal.Decimal `mapstructure:"-"  json:",omitempty"`
	// The format of the deviation being provided
	// * percent - a percentage deviation
	// * absolute - an unsigned numeric difference
	DeviationType string `mapstructure:"deviationType" json:"deviationType,omitempty" jsonschema:"enum=percent,enum=absolute,enum=none"`
}

type reduceAggregator struct {
	config ReduceAggConfig
}

var _ types.Aggregator = (*reduceAggregator)(nil)

func (a *reduceAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	if len(observations) < 2*f+1 {
		return nil, fmt.Errorf("not enough observations, have %d want %d", len(observations), 2*f+1)
	}

	currentState, err := a.initializeCurrentState(lggr, previousOutcome)
	if err != nil {
		return nil, err
	}

	report := map[string]any{}
	shouldReport := false

	for _, field := range a.config.Fields {
		vals := a.extractValues(lggr, observations, field.InputKey)

		// only proceed if every field has reached the minimum number of observations
		if len(vals) < 2*f+1 {
			return nil, fmt.Errorf("not enough observations provided %s, have %d want %d", field.InputKey, len(vals), 2*f+1)
		}

		singleValue, err := reduce(field.Method, vals)
		if err != nil {
			return nil, fmt.Errorf("unable to reduce on method %s, err: %s", field.Method, err.Error())
		}

		if field.DeviationType != DEVIATION_TYPE_NONE {
			oldValue := (*currentState)[field.InputKey]
			currDeviation, err := deviation(field.DeviationType, oldValue, singleValue)
			if oldValue != nil && err != nil {
				return nil, fmt.Errorf("unable to determine deviation %s", err.Error())
			}
			if oldValue == nil || currDeviation.GreaterThan(field.Deviation) {
				shouldReport = true
			}
			lggr.Debugw("checked deviation", "key", field.InputKey, "deviationType", field.DeviationType, "currentDeviation", currDeviation.String(), "targetDeviation", field.Deviation.String(), "shouldReport", shouldReport)
		}

		(*currentState)[field.InputKey] = singleValue
		if len(field.OutputKey) > 0 {
			report[field.OutputKey] = singleValue
		} else {
			report[field.InputKey] = singleValue
		}
	}

	// If none of the AggregationFields define deviation, always report
	hasNoDeviation := true
	for _, field := range a.config.Fields {
		if field.DeviationType != DEVIATION_TYPE_NONE {
			hasNoDeviation = false
			break
		}
	}
	if hasNoDeviation {
		lggr.Debugw("no deviation defined, reporting")
		shouldReport = true
	}

	stateValuesMap, err := values.WrapMap(currentState)
	if err != nil {
		return nil, err
	}
	stateBytes, err := proto.Marshal(values.ProtoMap(stateValuesMap))
	if err != nil {
		return nil, err
	}

	toWrap, err := formatReport(report, a.config.ReportFormat)
	if err != nil {
		return nil, err
	}
	reportValuesMap, err := values.NewMap(map[string]any{
		a.config.OutputFieldName: toWrap,
	})
	if err != nil {
		return nil, err
	}
	reportProtoMap := values.Proto(reportValuesMap).GetMapValue()

	lggr.Debugw("Aggregation complete", "shouldReport", shouldReport)

	return &types.AggregationOutcome{
		EncodableOutcome: reportProtoMap,
		Metadata:         stateBytes,
		ShouldReport:     shouldReport,
	}, nil
}

func (a *reduceAggregator) initializeCurrentState(lggr logger.Logger, previousOutcome *types.AggregationOutcome) (*map[string]values.Value, error) {
	currentState := map[string]values.Value{}

	if previousOutcome != nil {
		pb := &pb.Map{}
		proto.Unmarshal(previousOutcome.Metadata, pb)
		mv, err := values.FromMapValueProto(pb)
		if err != nil {
			return nil, err
		}
		err = mv.UnwrapTo(currentState)
		if err != nil {
			return nil, err
		}
	}

	zeroValue := values.NewDecimal(decimal.Zero)
	for _, field := range a.config.Fields {
		if _, ok := currentState[field.InputKey]; !ok {
			currentState[field.InputKey] = zeroValue
			lggr.Debugw("initializing empty onchain state for feed", "fieldInputKey", field.InputKey)
		}
	}

	lggr.Debugw("current state initialized", "state", currentState, "previousOutcome", previousOutcome)
	return &currentState, nil
}

func (a *reduceAggregator) extractValues(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value, aggregationKey string) (vals []values.Value) {
	for nodeID, nodeObservations := range observations {
		// we only expect a single observation per node
		if len(nodeObservations) == 0 || nodeObservations[0] == nil {
			lggr.Warnf("node %d contributed with empty observations", nodeID)
			continue
		}
		if len(nodeObservations) > 1 {
			lggr.Warnf("node %d contributed with more than one observation", nodeID)
			continue
		}

		val, err := nodeObservations[0].Unwrap()
		if err != nil {
			lggr.Warnf("node %d contributed a Value that could not be unwrapped", nodeID)
			continue
		}

		// if the observation data is a complex type, extract the value using the inputKey
		// values are then re-wrapped here to handle aggregating against Value types
		// which is used for mode aggregation
		switch val := val.(type) {
		case map[string]interface{}:
			_, ok := val[aggregationKey]
			if !ok {
				continue
			}

			rewrapped, err := values.Wrap(val[aggregationKey])
			if err != nil {
				lggr.Warnf("unable to wrap value %s", val[aggregationKey])
				continue
			}
			vals = append(vals, rewrapped)
		case []interface{}:
			i, err := strconv.Atoi(aggregationKey)
			if err != nil {
				lggr.Warnf("aggregation key %s could not be used to index a list type", aggregationKey)
				continue
			}
			rewrapped, err := values.Wrap(val[i])
			if err != nil {
				lggr.Warnf("unable to wrap value %s", val[i])
				continue
			}
			vals = append(vals, rewrapped)
		default:
			// not a complex type, use raw value
			if len(aggregationKey) == 0 {
				vals = append(vals, nodeObservations[0])
			} else {
				lggr.Warnf("aggregation key %s provided, but value is not an indexable type", aggregationKey)
			}
		}
	}

	return vals
}

func reduce(method string, items []values.Value) (values.Value, error) {
	switch method {
	case AGGREGATION_METHOD_MEDIAN:
		return median(items)
	case AGGREGATION_METHOD_MODE:
		return mode(items)
	default:
		return nil, fmt.Errorf("unsupported aggregation method %s", method)
	}
}

func median(items []values.Value) (values.Value, error) {
	if len(items) == 0 {
		return nil, errors.New("items cannot be empty")
	}
	sortAsDecimal(items)
	return items[(len(items)-1)/2], nil
}

func sortAsDecimal(items []values.Value) error {
	var err error
	sort.Slice(items, func(i, j int) bool {
		decimalI, errI := toDecimal(items[i])
		if errI != nil {
			err = errI
		}
		decimalJ, errJ := toDecimal(items[j])
		if errJ != nil {
			err = errJ
		}
		return decimalI.GreaterThan(decimalJ)
	})
	if err != nil {
		return err
	}
	return nil
}

func toDecimal(item values.Value) (decimal.Decimal, error) {
	unwrapped, err := item.Unwrap()
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	switch v := unwrapped.(type) {
	case string:
		deci, err := decimal.NewFromString(unwrapped.(string))
		if err != nil {
			return decimal.NewFromInt(0), err
		}
		return deci, nil
	case decimal.Decimal:
		return unwrapped.(decimal.Decimal), nil
	case int64:
		return decimal.NewFromInt(unwrapped.(int64)), nil
	case *big.Int:
		big := unwrapped.(*big.Int)
		return decimal.NewFromBigInt(big, 10), nil
	case time.Time:
		return decimal.NewFromInt(unwrapped.(time.Time).Unix()), nil
	case float64:
		return decimal.NewFromFloat(unwrapped.(float64)), nil
	default:
		// unsupported type
		return decimal.NewFromInt(0), fmt.Errorf("unable to convert type %T to decimal", v)
	}
}

func mode(items []values.Value) (values.Value, error) {
	if len(items) == 0 {
		return nil, errors.New("items cannot be empty")
	}

	counts := make(map[[32]byte]*counter)
	for _, item := range items {
		marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(item))
		if err != nil {
			return nil, err
		}
		sha := sha256.Sum256(marshalled)
		elem, ok := counts[sha]
		if !ok {
			counts[sha] = &counter{
				fullObservation: item,
				count:           1,
			}
		} else {
			elem.count++
		}
	}

	var maxCount int
	for _, ctr := range counts {
		if ctr.count > maxCount {
			maxCount = ctr.count
		}
	}

	var modes []values.Value
	for _, ctr := range counts {
		if ctr.count == maxCount {
			modes = append(modes, ctr.fullObservation)
		}
	}

	// If more than one mode found, choose first

	return modes[0], nil
}

func deviation(method string, previousValue values.Value, nextValue values.Value) (decimal.Decimal, error) {
	prevDeci, err := toDecimal(previousValue)
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	nextDeci, err := toDecimal(nextValue)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	diff := prevDeci.Sub(nextDeci).Abs()

	switch method {
	case DEVIATION_TYPE_ABSOLUTE:
		return diff, nil
	case DEVIATION_TYPE_PERCENT:
		if prevDeci.Cmp(decimal.NewFromInt(0)) == 0 {
			if diff.Cmp(decimal.NewFromInt(0)) == 0 {
				return decimal.NewFromInt(0), nil
			}
			return decimal.NewFromInt(math.MaxInt), nil
		}
		return diff.Div(prevDeci), nil
	default:
		return decimal.NewFromInt(0), fmt.Errorf("unsupported deviation method %s", method)
	}
}

func formatReport(report map[string]any, format string) (any, error) {
	switch format {
	case "array":
		return []map[string]any{report}, nil
	case "map":
		return report, nil
	default:
		return nil, errors.New("unsupported report format")
	}
}

func isOneOf(toCheck string, options []string) bool {
	for _, option := range options {
		if toCheck == option {
			return true
		}
	}
	return false
}

func NewReduceAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := ParseConfigReduceAggregator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	return &reduceAggregator{
		config: parsedConfig,
	}, nil
}

func ParseConfigReduceAggregator(config values.Map) (ReduceAggConfig, error) {
	parsedConfig := ReduceAggConfig{}
	if err := config.UnwrapTo(&parsedConfig); err != nil {
		return ReduceAggConfig{}, err
	}

	// validations & fill defaults
	if len(parsedConfig.Fields) == 0 {
		return ReduceAggConfig{}, errors.New("reduce aggregator must contain config for Fields to aggregate")
	}
	for i, field := range parsedConfig.Fields {
		if len(field.DeviationType) == 0 {
			field.DeviationType = DEVIATION_TYPE_NONE
			parsedConfig.Fields[i].DeviationType = DEVIATION_TYPE_NONE
		}
		if !isOneOf(field.DeviationType, []string{DEVIATION_TYPE_ABSOLUTE, DEVIATION_TYPE_PERCENT, DEVIATION_TYPE_NONE}) {
			return ReduceAggConfig{}, fmt.Errorf("invalid config DeviationType. received: %s. options: [%s, %s, %s]", field.DeviationType, DEVIATION_TYPE_ABSOLUTE, DEVIATION_TYPE_PERCENT, DEVIATION_TYPE_NONE)
		}
		if field.DeviationType != DEVIATION_TYPE_NONE && len(field.DeviationString) == 0 {
			return ReduceAggConfig{}, errors.New("aggregation field deviation must contain DeviationString amount")
		}
		if field.DeviationType != DEVIATION_TYPE_NONE && len(field.DeviationString) > 0 {
			deci, err := decimal.NewFromString(field.DeviationString)
			if err != nil {
				return ReduceAggConfig{}, fmt.Errorf("reduce aggregator could not parse deviation decimal from string %s", field.DeviationString)
			}
			parsedConfig.Fields[i].Deviation = deci
		}
		if len(field.Method) == 0 || !isOneOf(field.Method, []string{AGGREGATION_METHOD_MEDIAN, AGGREGATION_METHOD_MODE}) {
			return ReduceAggConfig{}, fmt.Errorf("aggregation field must contain a method. options: [%s, %s]", AGGREGATION_METHOD_MEDIAN, AGGREGATION_METHOD_MODE)
		}
		if len(field.DeviationString) > 0 && field.DeviationType == DEVIATION_TYPE_NONE {
			return ReduceAggConfig{}, fmt.Errorf("aggregation field cannot have deviation with a deviation type of %s", DEVIATION_TYPE_NONE)
		}
	}
	if len(parsedConfig.OutputFieldName) == 0 {
		parsedConfig.OutputFieldName = DEFAULT_OUTPUT_FIELD_NAME
	}
	if len(parsedConfig.ReportFormat) == 0 {
		parsedConfig.ReportFormat = DEFAULT_REPORT_FORMAT
	}
	if !isOneOf(parsedConfig.ReportFormat, []string{REPORT_FORMAT_ARRAY, REPORT_FORMAT_MAP}) {
		return ReduceAggConfig{}, fmt.Errorf("invalid config ReportFormat. received: %s. options: map, array", parsedConfig.ReportFormat)
	}

	return parsedConfig, nil
}
