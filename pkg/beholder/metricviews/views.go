// Package metricviews defines Beholder's default OTel metric views for
// cardinality control: PerWorkflow histogram bucket reduction and
// attribute-filter deny/allow lists.
//
// Callers (e.g. chainlink core/cmd/shell.go via beholder.Config.MetricViews)
// may supply additional views—typically histogram bucket Aggregation overrides
// for specific instrument names. Beholder merges caller views before these
// defaults (see beholder.Config.metricViews); callers do not need to invoke
// Default themselves.
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

// Default returns views appended after caller-supplied MetricViews by
// beholder.Config.metricViews, in the registration order the precedence rule
// in the package doc depends on:
//
//  1. PerWorkflow histogram bucket views, so they win for CRE limit metrics
//     (chainlink supplies no caller views for those names). Each carries the
//     denyKeys attribute filter on its own Stream mask, because winning the
//     stream identity for a bucket override also excludes the "*" deny view.
//  2. Attribute-filter allow-list views, most-specific matcher first.
//  3. The global "*" deny-list catch-all, built from denyKeys. Skipped when
//     denyKeys is empty—no attributes are stripped globally until configured—
//     but the fixed views above still apply.
func Default(denyKeys []string) []sdkmetric.View {
	denyFilter := denyKeysFilter(denyKeys)

	views := perWorkflowHistogramViews(denyFilter)
	views = append(views,
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: stoppedResendingInstrument},
			sdkmetric.Stream{AttributeFilter: stoppedResendingAllow},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: baseTriggerInstrumentGlob},
			sdkmetric.Stream{AttributeFilter: baseTriggerAllow},
		),
	)
	if denyFilter == nil {
		return views
	}
	return append(views,
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: "*"},
			sdkmetric.Stream{AttributeFilter: denyFilter},
		),
	)
}

func denyKeysFilter(denyKeys []string) attribute.Filter {
	if len(denyKeys) == 0 {
		return nil
	}
	keys := make([]attribute.Key, len(denyKeys))
	for i, k := range denyKeys {
		keys[i] = attribute.Key(k)
	}
	return attribute.NewDenyKeysFilter(keys...)
}
