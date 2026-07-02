package beholder

import "go.opentelemetry.io/otel/attribute"

// resourceAttributesToStringMap converts OTel resource attributes into a plain string map,
// using attribute.Value.Emit for canonical stringification of any value type. This is the
// single source of truth used to derive both the gRPC metadata headers and the CloudEvent
// extension keys/values sent to ChipIngress, so both mechanisms stay consistent.
func resourceAttributesToStringMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
