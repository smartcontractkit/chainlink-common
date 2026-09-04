package pg

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type dbStatsBeholderMetrics struct {
	connsMax     metric.Int64Gauge
	connsOpen    metric.Int64Gauge
	connsInUse   metric.Int64Gauge
	waitCount    metric.Int64Gauge
	waitDuration metric.Float64Gauge
}

func newDBStatsBeholderMetrics(lggr logger.Logger) *dbStatsBeholderMetrics {
	meter := beholder.GetMeter()
	m := &dbStatsBeholderMetrics{
		connsMax:     noop.Int64Gauge{},
		connsOpen:    noop.Int64Gauge{},
		connsInUse:   noop.Int64Gauge{},
		waitCount:    noop.Int64Gauge{},
		waitDuration: noop.Float64Gauge{},
	}

	if g, err := meter.Int64Gauge("db_conns_max",
		metric.WithDescription("Maximum number of open connections to the database.")); err != nil {
		lggr.Errorw("Failed to create db_conns_max beholder gauge", "err", err)
	} else {
		m.connsMax = g
	}

	if g, err := meter.Int64Gauge("db_conns_open",
		metric.WithDescription("The number of established connections both in use and idle.")); err != nil {
		lggr.Errorw("Failed to create db_conns_open beholder gauge", "err", err)
	} else {
		m.connsOpen = g
	}

	if g, err := meter.Int64Gauge("db_conns_used",
		metric.WithDescription("The number of connections currently in use.")); err != nil {
		lggr.Errorw("Failed to create db_conns_used beholder gauge", "err", err)
	} else {
		m.connsInUse = g
	}

	if g, err := meter.Int64Gauge("db_wait_count",
		metric.WithDescription("The total number of connections waited for.")); err != nil {
		lggr.Errorw("Failed to create db_wait_count beholder gauge", "err", err)
	} else {
		m.waitCount = g
	}

	if g, err := meter.Float64Gauge("db_wait_time_seconds",
		metric.WithDescription("The total time blocked waiting for a new connection."),
		metric.WithUnit("s")); err != nil {
		lggr.Errorw("Failed to create db_wait_time_seconds beholder gauge", "err", err)
	} else {
		m.waitDuration = g
	}

	return m
}

func (m *dbStatsBeholderMetrics) record(ctx context.Context, stats sql.DBStats) {
	m.connsMax.Record(ctx, int64(stats.MaxOpenConnections))
	m.connsOpen.Record(ctx, int64(stats.OpenConnections))
	m.connsInUse.Record(ctx, int64(stats.InUse))
	m.waitCount.Record(ctx, stats.WaitCount)
	m.waitDuration.Record(ctx, stats.WaitDuration.Seconds())
}
