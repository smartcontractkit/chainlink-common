// Package metricviews defines Beholder's default OTel metric views for
// cardinality control: PerWorkflow histogram bucket reduction and
// attribute-filter deny/allow lists.
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

// DefaultViews returns views appended after caller-supplied MetricViews by
// beholder.mergeMetricViews. PerWorkflow histogram bucket views are registered
// first so they apply to CRE limit metrics (chainlink does not supply caller
// views for those names). Attribute-filter views follow; within that group,
// more specific instrument matchers precede the global "*" catch-all.
func DefaultViews() []sdkmetric.View {
	views := perWorkflowHistogramViews()
	views = append(views,
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
	)
	return views
}
