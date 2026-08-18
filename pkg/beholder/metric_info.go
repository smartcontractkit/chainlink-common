package beholder

import "go.opentelemetry.io/otel/metric"

type MetricInfo struct {
	Name        string
	Unit        string
	Description string
}

// NewInt64Counter creates a new Int64Counter metric
func (m MetricInfo) NewInt64Counter(meter metric.Meter) (metric.Int64Counter, error) {
	return meter.Int64Counter(
		m.Name,
		metric.WithUnit(m.Unit),
		metric.WithDescription(m.Description),
	)
}

// NewInt64Gauge creates a new Int64Gauge metric
func (m MetricInfo) NewInt64Gauge(meter metric.Meter) (metric.Int64Gauge, error) {
	return meter.Int64Gauge(
		m.Name,
		metric.WithUnit(m.Unit),
		metric.WithDescription(m.Description),
	)
}

// NewInt64Histogram creates a new Int64Histogram metric
func (m MetricInfo) NewInt64Histogram(meter metric.Meter) (metric.Int64Histogram, error) {
	return meter.Int64Histogram(
		m.Name,
		metric.WithUnit(m.Unit),
		metric.WithDescription(m.Description),
	)
}

// NewFloat64Gauge creates a new Float64Gauge metric
func (m MetricInfo) NewFloat64Gauge(meter metric.Meter) (metric.Float64Gauge, error) {
	return meter.Float64Gauge(
		m.Name,
		metric.WithUnit(m.Unit),
		metric.WithDescription(m.Description),
	)
}

// NewFloat64Histogram creates a new Float64Histogram metric
func (m MetricInfo) NewFloat64Histogram(meter metric.Meter) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(
		m.Name,
		metric.WithUnit(m.Unit),
		metric.WithDescription(m.Description),
	)
}
