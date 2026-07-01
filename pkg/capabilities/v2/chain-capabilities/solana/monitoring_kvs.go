package solana

import (
	"encoding/hex"

	"go.opentelemetry.io/otel/attribute"
)

// LogKVs and MetricKVs implement monitoring.MonitoringLabels on *WriteReportRequest.
// Receiver is high-cardinality and is included in logs only.
func (r *WriteReportRequest) LogKVs() []any {
	return []any{
		"receiver", hex.EncodeToString(r.GetReceiver()),
	}
}

func (r *WriteReportRequest) MetricKVs() []attribute.KeyValue {
	return nil
}
