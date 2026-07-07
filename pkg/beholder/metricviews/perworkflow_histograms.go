package metricviews

import sdkmetric "go.opentelemetry.io/otel/sdk/metric"

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
	perWorkflowCountBoundaries = []float64{
		0, 1, 10, 100, 1e3, 1e4, 1e5,
	}
)

func perWorkflowHistogramViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
				Unit: "By",
			},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: perWorkflowBytesBoundaries,
			}},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
				Unit: "s",
			},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: perWorkflowSecondsBoundaries,
			}},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name: perWorkflowInstrumentGlob,
				Kind: sdkmetric.InstrumentKindHistogram,
			},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: perWorkflowCountBoundaries,
			}},
		),
	}
}
