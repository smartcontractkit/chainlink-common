package solana

import "encoding/hex"

// LogKVs and MetricKVs implement monitoring.MonitoringLabels on *WriteReportRequest.
// Receiver is high-cardinality and is included in logs only.
func (r *WriteReportRequest) LogKVs() []any {
	return []any{
		"receiver", hex.EncodeToString(r.GetReceiver()),
	}
}

func (r *WriteReportRequest) MetricKVs() []any {
	return nil
}
