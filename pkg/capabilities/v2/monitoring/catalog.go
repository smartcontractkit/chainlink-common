package monitoring

import (
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	MetricPrefix = "capabilities_v2"

	OutcomeSuccess = "success"
	OutcomeError   = "error"

	MetricSuffixCount             = "_count"
	MetricSuffixCapTimestampStart = "_cap_timestamp_start"
	MetricSuffixCapTimestampEmit  = "_cap_timestamp_emit"
	MetricSuffixCapDuration       = "_cap_duration"

	// DurationInstrumentPattern matches all v2 action duration histograms for MetricViews.
	DurationInstrumentPattern = MetricPrefix + "_*" + MetricSuffixCapDuration

	LabelMethod          = "method"
	LabelChainFamilyName = "chain_family_name"
	LabelChainID         = "chain_id"
	LabelNetworkName     = "network_name"
	LabelNetworkNameFull = "network_name_full"
	LabelWorkflowDonID   = "workflow_don_id"
	LabelCapabilityType  = "capability_type"
	LabelCapabilityID    = "capability_id"
)

// ActionLatencyBucketBoundariesMs matches read-action latency buckets.
var ActionLatencyBucketBoundariesMs = []float64{
	0, 5, 10, 25, 50, 75, 100,
	250, 500, 750, 1000,
	2500, 5000, 7500, 10000,
	15000, 30000,
}

// ActionMetricPrefix returns the metric name prefix for a method/outcome pair.
// Example: capabilities_v2_writereport_success
func ActionMetricPrefix(method, outcome string) string {
	return fmt.Sprintf("%s_%s_%s", MetricPrefix, SanitizeMetricToken(method), outcome)
}

// MetricName returns a full instrument name for dashboards and observability codegen.
// Example: MetricName("WriteReport", OutcomeSuccess, MetricSuffixCapDuration)
func MetricName(method, outcome, suffix string) string {
	return ActionMetricPrefix(method, outcome) + suffix
}

// ActionMetricEventRef returns a stable event reference string for metric descriptions.
func ActionMetricEventRef(method, outcome string) string {
	return fmt.Sprintf("capabilities.v2.%s.%s", SanitizeMetricToken(method), outcome)
}

// ActionMetricInfo returns instrument metadata for a v2 action method outcome.
func ActionMetricInfo(method, outcome string) metricsInfoCapBasic {
	prefix := ActionMetricPrefix(method, outcome)
	eventRef := ActionMetricEventRef(method, outcome)
	return metricsInfoCapBasic{
		count: beholder.MetricInfo{
			Name:        prefix + MetricSuffixCount,
			Description: fmt.Sprintf("The count of message: '%s' emitted", eventRef),
		},
		capTimestampStart: beholder.MetricInfo{
			Name:        prefix + MetricSuffixCapTimestampStart,
			Unit:        "ms",
			Description: fmt.Sprintf("The timestamp (local) at capability exec start that resulted in message: '%s' emit", eventRef),
		},
		capTimestampEmit: beholder.MetricInfo{
			Name:        prefix + MetricSuffixCapTimestampEmit,
			Unit:        "ms",
			Description: fmt.Sprintf("The timestamp (local) at message: '%s' emit", eventRef),
		},
		capDuration: beholder.MetricInfo{
			Name:        prefix + MetricSuffixCapDuration,
			Unit:        "ms",
			Description: fmt.Sprintf("The duration (local) since capability exec start to message: '%s' emit", eventRef),
		},
	}
}

// SanitizeMetricToken normalizes RPC method names for metric instrument names.
func SanitizeMetricToken(s string) string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return replacer.Replace(s)
}
