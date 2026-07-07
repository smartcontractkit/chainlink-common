// Package metricviews defines Beholder's default OTel metric views for
// cardinality-limiting attribute filters.
//
// Callers (e.g. chainlink core/cmd/shell.go via beholder.Config.MetricViews)
// may supply additional views—typically histogram bucket Aggregation overrides
// for specific instrument names. Beholder merges caller views before these
// defaults (see beholder.mergeMetricViews); callers do not need to invoke
// DefaultViews themselves.
//
// An instrument may match multiple views. When several matching views resolve
// to the same output stream (same name/description/unit/kind), the SDK keeps
// only the first in registration order and logs a duplicate-stream warning for
// the rest; attribute filters and aggregations do not compose across them.
// Because Beholder registers caller views ahead of these defaults, a matching
// caller view wins and the default attribute filter for that stream is dropped.
package metricviews

import (
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	baseTriggerInstrumentGlob  = "capabilities_base_trigger_*"
	stoppedResendingInstrument = "capabilities_base_trigger_stopped_resending_timestamp"
)

var (
	globalHighCardinalityDeny = attribute.NewDenyKeysFilter(
		attribute.Key("event_id"),
		attribute.Key("trigger_id"),
		attribute.Key("workflow_execution_id"),
	)

	baseTriggerAllow = attribute.NewAllowKeysFilter(
		attribute.Key("capability_id"),
		attribute.Key("reason"),
		attribute.Key("outcome"),
	)

	stoppedResendingAllow = attribute.NewAllowKeysFilter(
		attribute.Key("capability_id"),
		attribute.Key("trigger_id"),
	)
)

// DefaultViews returns attribute-filter views appended after caller-supplied
// MetricViews by beholder.mergeMetricViews. Within this slice, more specific
// instrument matchers precede the global "*" catch-all: an instrument like
// capabilities_base_trigger_stopped_resending_timestamp matches all three, and
// since these views resolve to the same stream identity the SDK keeps the first
// in registration order. Ordering most-specific-first gives it precedence:
//
//  1. capabilities_base_trigger_stopped_resending_timestamp — allow capability_id, trigger_id
//  2. capabilities_base_trigger_* — allow capability_id, reason, outcome
//  3. * — deny event_id, trigger_id, workflow_execution_id
func DefaultViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: stoppedResendingInstrument},
			sdkmetric.Stream{AttributeFilter: stoppedResendingAllow},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: baseTriggerInstrumentGlob},
			sdkmetric.Stream{AttributeFilter: baseTriggerAllow},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: "*"},
			sdkmetric.Stream{AttributeFilter: globalHighCardinalityDeny},
		),
	}
}
