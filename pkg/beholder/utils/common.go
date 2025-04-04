package utils

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// MetricInfo is a struct for metrics information
type MetricsInfoCapBasic struct {
	// common
	count             MetricInfo
	capTimestampStart MetricInfo
	capTimestampEmit  MetricInfo
	capDuration       MetricInfo // ts.emit - ts.start
}

// NewMetricsInfoCapBasic creates a new MetricsInfoCapBasic using the provided event/metric information
func NewMetricsInfoCapBasic(metricPrefix, eventRef string) MetricsInfoCapBasic {
	return MetricsInfoCapBasic{
		count: MetricInfo{
			Name:        fmt.Sprintf("%s_count", metricPrefix),
			Unit:        "",
			Description: fmt.Sprintf("The count of message: '%s' emitted", eventRef),
		},
		capTimestampStart: MetricInfo{
			Name:        fmt.Sprintf("%s_cap_timestamp_start", metricPrefix),
			Unit:        "ms",
			Description: fmt.Sprintf("The timestamp (local) at capability exec start that resulted in message: '%s' emit", eventRef),
		},
		capTimestampEmit: MetricInfo{
			Name:        fmt.Sprintf("%s_cap_timestamp_emit", metricPrefix),
			Unit:        "ms",
			Description: fmt.Sprintf("The timestamp (local) at message: '%s' emit", eventRef),
		},
		capDuration: MetricInfo{
			Name:        fmt.Sprintf("%s_cap_duration", metricPrefix),
			Unit:        "ms",
			Description: fmt.Sprintf("The duration (local) since capability exec start to message: '%s' emit", eventRef),
		},
	}
}

// MetricsCapBasic is a base struct for metrics related to a capability
type MetricsCapBasic struct {
	count             metric.Int64Counter
	capTimestampStart metric.Int64Gauge
	capTimestampEmit  metric.Int64Gauge
	capDuration       metric.Int64Gauge // ts.emit - ts.start
}

// NewMetricsCapBasic creates a new MetricsCapBasic using the provided MetricsInfoCapBasic
func NewMetricsCapBasic(info MetricsInfoCapBasic) (MetricsCapBasic, error) {
	meter := beholder.GetMeter()
	set := MetricsCapBasic{}

	// Create new metrics
	var err error

	set.count, err = info.count.NewInt64Counter(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create new counter: %w", err)
	}

	set.capTimestampStart, err = info.capTimestampStart.NewInt64Gauge(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create new gauge: %w", err)
	}

	set.capTimestampEmit, err = info.capTimestampEmit.NewInt64Gauge(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create new gauge: %w", err)
	}

	set.capDuration, err = info.capDuration.NewInt64Gauge(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create new gauge: %w", err)
	}

	return set, nil
}

func (m *MetricsCapBasic) RecordEmit(ctx context.Context, start, emit uint64, attrKVs ...attribute.KeyValue) {
	// Define attributes
	attrs := metric.WithAttributes(attrKVs...)

	// Count events
	m.count.Add(ctx, 1, attrs)

	// Timestamp events
	m.capTimestampStart.Record(ctx, int64(start), attrs)
	m.capTimestampEmit.Record(ctx, int64(emit), attrs)
	m.capDuration.Record(ctx, int64(emit-start), attrs)
}
