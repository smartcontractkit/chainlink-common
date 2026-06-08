package solana

import "encoding/hex"

// KVs implements capabilities.MonitoringLabels on *WriteReportRequest.
// The returned key-value pairs are appended to every lifecycle log line
// emitted by the generated server (initiated / succeeded / failed).
func (r *WriteReportRequest) KVs() []any {
	return []any{
		"receiver", hex.EncodeToString(r.GetReceiver()),
	}
}
