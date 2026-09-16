// Package aggregators holds the workflow-facing configuration for the
// consensus aggregation methods.
package aggregators

import (
	"github.com/shopspring/decimal"
)

// IdenticalAggConfig configures aggregation by the most frequent observation
// for each index of a data set.
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

// ReduceAggConfig configures condensing multiple observations into a single
// encodable outcome.
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
