package stellar

import (
	"go.opentelemetry.io/otel/attribute"
)

// LogKVs and MetricKVs implement monitoring.MonitoringLabels on *WriteReportRequest.
// ContractId is high-cardinality and is included in logs only.
func (r *WriteReportRequest) LogKVs() []any {
	return []any{
		"contractID", r.GetContractId(),
	}
}

func (r *WriteReportRequest) MetricKVs() []attribute.KeyValue {
	return nil
}

// LogKVs and MetricKVs implement monitoring.MonitoringLabels on *ReadContractRequest.
// ContractId and SourceAccount are high-cardinality and are included in logs only.
func (r *ReadContractRequest) LogKVs() []any {
	return []any{
		"contractID", r.GetContractId(),
		"function", r.GetFunction(),
		"sourceAccount", r.GetSourceAccount(),
	}
}

func (r *ReadContractRequest) MetricKVs() []attribute.KeyValue {
	return nil
}
