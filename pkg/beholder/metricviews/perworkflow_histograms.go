package metricviews

import (
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const perWorkflowInstrumentGlob = "*.PerWorkflow.*"

// OTel SDK default histogram boundaries (15 values → 16 Prometheus buckets).
// PerWorkflow limit metrics from pkg/settings/limits use the default because
// they do not pass metric.WithExplicitBucketBoundaries at creation time.
var (
	perWorkflowBytesBoundaries = []float64{
		0, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8,
	}
	perWorkflowSecondsBoundaries = []float64{
		0, 1, 10, 60, 300, 900, 3600,
	}
	// perWorkflowGasBoundaries covers pkg/settings/cresettings ChainWrite gas
	// limit defaults (Solana 300_000, Aptos 2_000_000, EVM 5_000_000, up to
	// 50_000_000 for per-chain-selector overrides) without collapsing them
	// into the +Inf overflow bucket.
	perWorkflowGasBoundaries = []float64{
		0, 1e5, 5e5, 1e6, 5e6, 1e7, 5e7,
	}
	// perWorkflowCountBoundaries is the fallback for PerWorkflow histograms
	// whose unit is neither "By", "s", nor "{gas}" (e.g. dimensionless counts).
	perWorkflowCountBoundaries = []float64{
		0, 1, 10, 100, 1e3, 1e4, 1e5,
	}
)

// perWorkflowHistogramViews returns bucket-boundary overrides for
// *.PerWorkflow.* histograms, keyed by unit. Each view carries denyFilter so
// the deny-list travels with the bucket override (see the view-precedence
// rule in the package doc); denyFilter is nil when no deny keys are
// configured, which is a no-op attribute filter. The unit-less count view is
// registered last and acts as a fallback: the more specific By/s/{gas} views
// win for their units, and the count view claims every other unit.
func perWorkflowHistogramViews(denyFilter attribute.Filter) []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
				Unit: "By",
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: perWorkflowBytesBoundaries,
				},
				AttributeFilter: denyFilter,
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
				Unit: "s",
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: perWorkflowSecondsBoundaries,
				},
				AttributeFilter: denyFilter,
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
				Unit: "{gas}",
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: perWorkflowGasBoundaries,
				},
				AttributeFilter: denyFilter,
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: perWorkflowCountBoundaries,
				},
				AttributeFilter: denyFilter,
			},
		),
	}
}
