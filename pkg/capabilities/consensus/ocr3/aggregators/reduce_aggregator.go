package aggregators

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

const (
	AGGREGATION_METHOD_MEDIAN = "median"
	AGGREGATION_METHOD_MODE   = "mode"
	// DEVIATION_TYPE_NONE is no deviation check
	DEVIATION_TYPE_NONE = "none"
	// DEVIATION_TYPE_ANY is any difference from the previous value to the next value
	DEVIATION_TYPE_ANY = "any"
	// DEVIATION_TYPE_PERCENT is a numeric percentage difference
	DEVIATION_TYPE_PERCENT = "percent"
	// DEVIATION_TYPE_ABSOLUTE is a numeric unsigned difference
	DEVIATION_TYPE_ABSOLUTE = "absolute"
	REPORT_FORMAT_MAP       = "map"
	REPORT_FORMAT_ARRAY     = "array"
	REPORT_FORMAT_VALUE     = "value"
	MODE_QUORUM_OCR         = "ocr"
	MODE_QUORUM_ALL         = "all"
	MODE_QUORUM_ANY         = "any"

	DEFAULT_REPORT_FORMAT     = REPORT_FORMAT_MAP
	DEFAULT_OUTPUT_FIELD_NAME = "Reports"
	DEFAULT_MODE_QUORUM       = MODE_QUORUM_OCR
)

type ReduceAggConfig struct {
	// Configuration on how to aggregate one or more data points
	Fields []AggregationField `mapstructure:"fields"  required:"true"`
	// The top level field name that report data is put into
	OutputFieldName string `mapstructure:"outputFieldName" json:"outputFieldName" default:"Reports"`
	// The structure surrounding the report data that is put on to "OutputFieldName"
	ReportFormat string `mapstructure:"reportFormat" json:"reportFormat" default:"map" jsonschema:"enum=map,enum=array,enum=value"`
	// Optional key name, that when given will contain a nested map with designated Fields moved into it
	// If given, one or more fields must be given SubMapField: true
	SubMapKey string `mapstructure:"subMapKey" json:"subMapKey" default:""`
}

type AggregationField struct {
	// An optional check to only report when the difference from the previous report exceeds a certain threshold.
	// Can only be used when the field is of a numeric type: string, decimal, uint64, int64, big.Int, time.Time, float64
	// If no deviation is provided on any field, there will always be a report once minimum observations are reached.
	Deviation       decimal.Decimal `mapstructure:"-"  json:"-"`
	DeviationString string          `mapstructure:"deviation"  json:"deviation,omitempty"`
	// The format of the deviation being provided
	// * percent - a percentage deviation
	// * absolute - an unsigned numeric difference
	// * none - no deviation check
	// * any - any difference from the previous value to the next value
	DeviationType string `mapstructure:"deviationType" json:"deviationType,omitempty" jsonschema:"enum=percent,enum=absolute,enum=none,enum=any"`
	// The key to find a data point within the input data
	// If omitted, the entire input will be used
	InputKey string `mapstructure:"inputKey" json:"inputKey"`
	// How the data set should be aggregated to a single value
	// * median - take the centermost value of the sorted data set of observations. can only be used on numeric types. not a true median, because no average if two middle values.
	// * mode - take the most frequent value. if tied, use the "first". use "ModeQuorom" to configure the minimum number of seen values.
	Method string `mapstructure:"method" json:"method" jsonschema:"enum=median,enum=mode" required:"true"`
	// When using Method=mode, this will configure the minimum number of values that must be seen
	// * ocr - (default) enforces that the number of matching values must be at least f+1, otherwise consensus fails
	// * any - do not enforce any limit on the minimum viable count. this may result in unexpected answers if every observation is unique.
	ModeQuorum string `mapstructure:"modeQuorum" json:"modeQuorum,omitempty" jsonschema:"enum=ocr,enum=any" default:"ocr"`
	// The key that the aggregated data is put under
	// If omitted, the InputKey will be used
	OutputKey string `mapstructure:"outputKey" json:"outputKey"`
	// If enabled, this field will be moved from the top level map
	// into a nested map on the key defined by "SubMapKey"
	SubMapField bool `mapstructure:"subMapField"  json:"subMapField,omitempty"`
}

type reduceAggregator struct {
	config ReduceAggConfig
}

var _ types.Aggregator = (*reduceAggregator)(nil)

// Condenses multiple observations into a single encodable outcome
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

		singleValue, err := reduce(field.Method, vals, f, field.ModeQuorum)
		if err != nil {
			return nil, fmt.Errorf("unable to reduce on method %s, err: %s", field.Method, err.Error())
		}

		shouldReportField, err := a.shouldReport(lggr, field, singleValue, currentState)
		if err != nil {
			return nil, fmt.Errorf("unable to determine if should report, err: %s", err.Error())
		}

		if shouldReportField || field.DeviationType == DEVIATION_TYPE_NONE {
			(*currentState)[field.OutputKey] = singleValue
		}

		if shouldReportField {
			shouldReport = true
		}

		if len(field.OutputKey) > 0 {
			report[field.OutputKey] = singleValue
		} else {
			report[field.InputKey] = singleValue
		}
	}

	// if SubMapKey is provided, move fields in a nested map
	if len(a.config.SubMapKey) > 0 {
		subMap := map[string]any{}
		for _, field := range a.config.Fields {
			if field.SubMapField {
				if len(field.OutputKey) > 0 {
					subMap[field.OutputKey] = report[field.OutputKey]
					delete(report, field.OutputKey)
				} else {
					subMap[field.InputKey] = report[field.InputKey]
					delete(report, field.InputKey)
				}
			}
		}
		report[a.config.SubMapKey] = subMap
	}

	// if none of the AggregationFields define deviation, always report
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
		return nil, fmt.Errorf("aggregate state wrapmap error: %s", err.Error())
	}
	stateBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.ProtoMap(stateValuesMap))
	if err != nil {
		return nil, fmt.Errorf("aggregate state proto marshal error: %s", err.Error())
	}

	toWrap, err := formatReport(report, a.config.ReportFormat)
	if err != nil {
		return nil, fmt.Errorf("aggregate formatReport error: %s", err.Error())
	}
	reportValuesMap, err := values.NewMap(map[string]any{
		a.config.OutputFieldName: toWrap,
	})
	if err != nil {
		return nil, fmt.Errorf("aggregate new map error: %s", err.Error())
	}
	reportProtoMap := values.Proto(reportValuesMap).GetMapValue()

	lggr.Debugw("Aggregation complete", "shouldReport", shouldReport)

	return &types.AggregationOutcome{
		EncodableOutcome: reportProtoMap,
		Metadata:         stateBytes,
		ShouldReport:     shouldReport,
	}, nil
}

func (a *reduceAggregator) shouldReport(lggr logger.Logger, field AggregationField, singleValue values.Value, currentState *map[string]values.Value) (bool, error) {
	if field.DeviationType == DEVIATION_TYPE_NONE {
		return false, nil
	}

	oldValue := (*currentState)[field.OutputKey]
	// this means its the first round and the field has not been initialised
	if oldValue == nil {
		return true, nil
	}

	if field.DeviationType == DEVIATION_TYPE_ANY {
		unwrappedOldValue, err := oldValue.Unwrap()
		if err != nil {
			return false, err
		}

		unwrappedSingleValue, err := singleValue.Unwrap()
		if err != nil {
			return false, err
		}

		// we will only report in case of a change in value
		switch v := unwrappedOldValue.(type) {
		case []byte:
			if !bytes.Equal(v, unwrappedSingleValue.([]byte)) {
				return true, nil
			}
		case map[string]any, []any:
			marshalledOldValue, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(oldValue))
			if err != nil {
				return false, err
			}

			marshalledSingleValue, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(singleValue))
			if err != nil {
				return false, err
			}
			if !bytes.Equal(marshalledOldValue, marshalledSingleValue) {
				return true, nil
			}
		default:
			if unwrappedOldValue != unwrappedSingleValue {
				return true, nil
			}
		}

		return false, nil
	}

	currDeviation, err := deviation(field.DeviationType, oldValue, singleValue)
	if err != nil {
		return false, fmt.Errorf("unable to determine deviation %s", err.Error())
	}

	if currDeviation.GreaterThan(field.Deviation) {
		lggr.Debugw("checked deviation", "key", field.OutputKey, "deviationType", field.DeviationType, "currentDeviation", currDeviation.String(), "targetDeviation", field.Deviation.String(), "shouldReport", true)
		return true, nil
	}

	return false, nil
}

func (a *reduceAggregator) initializeCurrentState(lggr logger.Logger, previousOutcome *types.AggregationOutcome) (*map[string]values.Value, error) {
	currentState := map[string]values.Value{}

	if previousOutcome != nil {
		pb := &pb.Map{}
		err := proto.Unmarshal(previousOutcome.Metadata, pb)
		if err != nil {
			return nil, fmt.Errorf("initializeCurrentState Unmarshal error: %w", err)
		}
		mv, err := values.FromMapValueProto(pb)
		if err != nil {
			return nil, fmt.Errorf("initializeCurrentState FromMapValueProto error: %w", err)
		}
		currentState = mv.Underlying
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
		case map[string]any:
			_, ok := val[aggregationKey]
			if !ok {
				continue
			}
			if val[aggregationKey] == nil {
				lggr.Warnf("node %d contributed with a nil value under key %s", nodeID, aggregationKey)
				continue
			}

			rewrapped, err := values.Wrap(val[aggregationKey])
			if err != nil {
				lggr.Warnf("unable to wrap value %s", val[aggregationKey])
				continue
			}
			vals = append(vals, rewrapped)
		case []any:
			i, err := strconv.Atoi(aggregationKey)
			if err != nil {
				lggr.Warnf("aggregation key %s could not be used to index a list type", aggregationKey)
				continue
			}
			if i >= len(val) {
				lggr.Warnf("node %d contributed with an array shorter than index %s", nodeID, aggregationKey)
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

func reduce(method string, items []values.Value, f int, modeQuorum string) (values.Value, error) {
	switch method {
	case AGGREGATION_METHOD_MEDIAN:
		return median(items)
	case AGGREGATION_METHOD_MODE:
		value, count, err := mode(items)
		if err != nil {
			return value, err
		}
		err = modeHasQuorum(modeQuorum, count, f)
		if err != nil {
			return value, err
		}
		return value, err
	default:
		// invariant, config should be validated
		return nil, fmt.Errorf("unsupported aggregation method %s", method)
	}
}

func median(items []values.Value) (values.Value, error) {
	if len(items) == 0 {
		// invariant, as long as f > 0 there should be items
		return nil, errors.New("items cannot be empty")
	}
	err := sortAsDecimal(items)
	if err != nil {
		return nil, err
	}
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
		deci, err := decimal.NewFromString(v)
		if err != nil {
			return decimal.NewFromInt(0), err
		}
		return deci, nil
	case decimal.Decimal:
		return v, nil
	case int64:
		return decimal.NewFromInt(v), nil
	case uint64:
		return decimal.NewFromUint64(v), nil
	case *big.Int:
		big := unwrapped.(*big.Int)
		return decimal.NewFromBigInt(big, 10), nil
	case time.Time:
		return decimal.NewFromInt(v.Unix()), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	default:
		// unsupported type
		return decimal.NewFromInt(0), fmt.Errorf("unable to convert type %T to decimal", v)
	}
}

func mode(items []values.Value) (values.Value, int, error) {
	if len(items) == 0 {
		// invariant, as long as f > 0 there should be items
		return nil, 0, errors.New("items cannot be empty")
	}

	counts := make(map[[32]byte]*counter)
	for _, item := range items {
		marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(item))
		if err != nil {
			// invariant: values should always be able to be proto marshalled
			return nil, 0, err
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

	// Collect all SHA keys that have the max count, then sort them to ensure
	// deterministic tie-breaking regardless of map iteration order.
	var modeKeys [][32]byte
	for sha, ctr := range counts {
		if ctr.count == maxCount {
			modeKeys = append(modeKeys, sha)
		}
	}
	sort.Slice(modeKeys, func(i, j int) bool {
		return bytes.Compare(modeKeys[i][:], modeKeys[j][:]) < 0
	})

	return counts[modeKeys[0]].fullObservation, maxCount, nil
}

func modeHasQuorum(quorumType string, count int, f int) error {
	switch quorumType {
	case MODE_QUORUM_ANY:
		return nil
	case MODE_QUORUM_OCR:
		if count < f+1 {
			return fmt.Errorf("mode quorum not reached. have: %d, want: %d", count, f+1)
		}
		return nil
	case MODE_QUORUM_ALL:
		if count < 2*f+1 {
			return fmt.Errorf("mode quorum not reached. have: %d, want: %d", count, 2*f+1)
		}
		return nil
	default:
		// invariant, config should be validated
		return fmt.Errorf("unsupported mode quorum %s", quorumType)
	}
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
	case REPORT_FORMAT_ARRAY:
		return []map[string]any{report}, nil
	case REPORT_FORMAT_MAP:
		return report, nil
	case REPORT_FORMAT_VALUE:
		for _, value := range report {
			return value, nil
		}
		// invariant: validation enforces only one output value
		return nil, errors.New("value format must contain at least one output")
	default:
		return nil, errors.New("unsupported report format")
	}
}

func isOneOf(toCheck string, options []string) bool {
	return slices.Contains(options, toCheck)
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
	if len(parsedConfig.OutputFieldName) == 0 {
		parsedConfig.OutputFieldName = DEFAULT_OUTPUT_FIELD_NAME
	}
	if len(parsedConfig.ReportFormat) == 0 {
		parsedConfig.ReportFormat = DEFAULT_REPORT_FORMAT
	}
	if len(parsedConfig.Fields) > 1 && parsedConfig.ReportFormat == REPORT_FORMAT_VALUE {
		return ReduceAggConfig{}, errors.New("report type of value can only have one field")
	}
	hasSubMapField := false
	outputKeyCount := make(map[any]bool)
	for i, field := range parsedConfig.Fields {
		if (parsedConfig.ReportFormat == REPORT_FORMAT_ARRAY || parsedConfig.ReportFormat == REPORT_FORMAT_MAP) && len(field.OutputKey) == 0 {
			return ReduceAggConfig{}, fmt.Errorf("report type %s or %s must have an OutputKey to put the result under", REPORT_FORMAT_ARRAY, REPORT_FORMAT_MAP)
		}
		if len(field.DeviationType) == 0 {
			field.DeviationType = DEVIATION_TYPE_NONE
			parsedConfig.Fields[i].DeviationType = DEVIATION_TYPE_NONE
		}
		if !isOneOf(field.DeviationType, []string{DEVIATION_TYPE_ABSOLUTE, DEVIATION_TYPE_PERCENT, DEVIATION_TYPE_NONE, DEVIATION_TYPE_ANY}) {
			return ReduceAggConfig{}, fmt.Errorf("invalid config DeviationType. received: %s. options: [%s, %s, %s]", field.DeviationType, DEVIATION_TYPE_ABSOLUTE, DEVIATION_TYPE_PERCENT, DEVIATION_TYPE_NONE)
		}
		if !isOneOf(field.DeviationType, []string{DEVIATION_TYPE_NONE, DEVIATION_TYPE_ANY}) && len(field.DeviationString) == 0 {
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
		if field.Method == AGGREGATION_METHOD_MODE && len(field.ModeQuorum) == 0 {
			field.ModeQuorum = DEFAULT_MODE_QUORUM
			parsedConfig.Fields[i].ModeQuorum = DEFAULT_MODE_QUORUM
		}
		if field.Method == AGGREGATION_METHOD_MODE && !isOneOf(field.ModeQuorum, []string{MODE_QUORUM_ANY, MODE_QUORUM_OCR, MODE_QUORUM_ALL}) {
			return ReduceAggConfig{}, fmt.Errorf("mode quorum must be one of options: [%s, %s, %s]", MODE_QUORUM_ANY, MODE_QUORUM_OCR, MODE_QUORUM_ALL)
		}
		if len(field.DeviationString) > 0 && isOneOf(field.DeviationType, []string{DEVIATION_TYPE_NONE, DEVIATION_TYPE_ANY}) {
			return ReduceAggConfig{}, fmt.Errorf("aggregation field cannot have deviation with a deviation type of %s", field.DeviationType)
		}
		if field.SubMapField {
			hasSubMapField = true
		}
		if outputKeyCount[field.OutputKey] {
			return ReduceAggConfig{}, errors.New("multiple fields have the same outputkey, which will overwrite each other")
		}
		outputKeyCount[field.OutputKey] = true
	}
	if len(parsedConfig.SubMapKey) > 0 && !hasSubMapField {
		return ReduceAggConfig{}, fmt.Errorf("sub Map key %s given, but no fields are marked as sub map fields", parsedConfig.SubMapKey)
	}
	if hasSubMapField && len(parsedConfig.SubMapKey) == 0 {
		return ReduceAggConfig{}, errors.New("fields are marked as sub Map fields, but no sub map key given")
	}
	if !isOneOf(parsedConfig.ReportFormat, []string{REPORT_FORMAT_ARRAY, REPORT_FORMAT_MAP, REPORT_FORMAT_VALUE}) {
		return ReduceAggConfig{}, fmt.Errorf("invalid config ReportFormat. received: %s. options: %s, %s, %s", parsedConfig.ReportFormat, REPORT_FORMAT_ARRAY, REPORT_FORMAT_MAP, REPORT_FORMAT_VALUE)
	}

	return parsedConfig, nil
}
