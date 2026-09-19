package sqlutil

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// sqlQueryTimeMetric records SQL query time as a percentage of timeout, duplicating [PromSQLQueryTime]
// as an OTel histogram via the Beholder meter.
type sqlQueryTimeMetric interface {
	Record(ctx context.Context, pct float64)
}

type beholderSQLQueryTimeMetric struct {
	histogram metric.Float64Histogram
}

func newSQLQueryTimeMetric(lggr logger.Logger) sqlQueryTimeMetric {
	histogram, err := beholder.GetMeter().Float64Histogram(
		"sql_query_timeout_percent",
		metric.WithDescription("SQL query time as a percentage of timeout."),
		metric.WithUnit("1"),
		metric.WithExplicitBucketBoundaries(sqlQueryTimeBuckets...),
	)
	if err != nil {
		lggr.Errorw("Failed to create sql_query_timeout_percent beholder histogram; disabling beholder SQL query time metric", "err", err)
		return noopSQLQueryTimeMetric{}
	}
	return &beholderSQLQueryTimeMetric{histogram: histogram}
}

func (m *beholderSQLQueryTimeMetric) Record(ctx context.Context, pct float64) {
	m.histogram.Record(ctx, pct)
}

type noopSQLQueryTimeMetric struct{}

func (noopSQLQueryTimeMetric) Record(context.Context, float64) {}

var (
	sqlQueryTimeMetricOnce   sync.Once
	globalSQLQueryTimeMetric sqlQueryTimeMetric
)

func getSQLQueryTimeMetric(lggr logger.Logger) sqlQueryTimeMetric {
	sqlQueryTimeMetricOnce.Do(func() {
		globalSQLQueryTimeMetric = newSQLQueryTimeMetric(lggr)
	})
	return globalSQLQueryTimeMetric
}
