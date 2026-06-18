package monitoring

// MonitoringLabels is an optional interface that request proto types may implement
// to contribute method-specific fields to structured logs and metrics emitted by the
// generated server (--with-monitoring). Returning nil or an empty slice is valid.
//
// LogKVs are appended to lifecycle log lines (initiated / succeeded / failed).
// High-cardinality values (addresses, IDs, etc.) belong here.
//
// MetricKVs are converted to OTel attributes and recorded by ActionMetrics on lifecycle
// events. Metrics are exported via Beholder OTLP and appear in Prometheus after collector
// ingestion. Keep MetricKVs low-cardinality to avoid overloading metrics storage.
//
// Example (in chainlink-common, package solana):
//
//	func (r *WriteReportRequest) LogKVs() []any {
//	    return []any{"receiver", hex.EncodeToString(r.GetReceiver())}
//	}
//
//	func (r *WriteReportRequest) MetricKVs() []any {
//	    return nil
//	}
//
// The generated server checks any(input).(monitoring.MonitoringLabels) at call
// time; if the request type does not implement it, monitoring still works — log
// lines just carry the method name and request metadata fields.
type MonitoringLabels interface {
	LogKVs() []any
	MetricKVs() []any
}
