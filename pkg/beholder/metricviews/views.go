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
// View precedence (the rule the rest of this package relies on): a single
// instrument may match several views, but the SDK identifies a stream by
// name/description/unit/kind only—not by aggregation or attribute filter—so
// every view matching one instrument resolves to the same stream identity. The
// SDK applies only the first matching view in registration order and drops the
// rest. No duplicate-stream warning is logged, because the identities are
// identical; that warning fires only on genuinely conflicting definitions.
//
// Two consequences follow: aggregations and attribute filters do not compose
// across matching views, and the first-registered matching view wins outright.
// Beholder registers caller views ahead of these defaults, so a caller view
// wins for its instrument and the default attribute filter no longer applies to
// that stream—which is why each bucket view below carries the deny filter
// itself instead of relying on the global "*" view.
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
// beholder.mergeMetricViews, in the registration order the precedence rule in
// the package doc depends on:
//
//  1. PerWorkflow histogram bucket views, so they win for CRE limit metrics
//     (chainlink supplies no caller views for those names). Each carries
//     globalHighCardinalityDeny on its own Stream mask, because winning the
//     stream identity for a bucket override also excludes the "*" deny view.
//  2. Attribute-filter views, most-specific matcher first, ending with the
//     global "*" catch-all.
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
