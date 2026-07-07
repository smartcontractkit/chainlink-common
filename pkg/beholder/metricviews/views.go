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
		attribute.Key("event_id"), // kept: low-cardinality vs retry; required by stopped-resending alert
	)
)

// DefaultViews returns cardinality-limiting views prepended before caller-supplied
// MetricViews. More specific instrument matchers are registered first so they take
// precedence over the global catch-all denylist.
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
