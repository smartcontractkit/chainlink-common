// Package metricviews defines Beholder's default OTel metric views for
// cardinality-limiting attribute filters.
//
// Callers (e.g. chainlink core/cmd/shell.go via beholder.Config.MetricViews)
// may supply additional views—typically histogram bucket Aggregation overrides
// for specific instrument names. Beholder merges caller views before these
// defaults (see beholder.Config.metricViews); callers do not need to invoke
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
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// DefaultViews returns deny-only attribute-filter views appended after
// caller-supplied MetricViews by beholder.Config.metricViews. The denylist is
// empty by default so no attributes are stripped until keys are configured.
func DefaultViews() []sdkmetric.View {
	return nil
}
