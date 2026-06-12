package capabilities

// MonitoringLabels is an optional interface that request proto types may implement
// to contribute method-specific fields to structured log lines emitted by the
// generated server (--with-monitoring). Returning nil or an empty slice is valid.
//
// Example (in chainlink-common, package solana):
//
//	func (r *WriteReportRequest) KVs() []any {
//	    return []any{"receiver", hex.EncodeToString(r.GetReceiver())}
//	}
//
// The generated server checks any(input).(capabilities.MonitoringLabels) at call
// time; if the request type does not implement it, monitoring still works — log
// lines just carry the method name and execution context.
type MonitoringLabels interface {
	KVs() []any
}
