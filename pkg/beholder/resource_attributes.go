package beholder

import "go.opentelemetry.io/otel/attribute"

// resourceAttributesToStringMap converts OTel resource attributes into a plain string map,
// using attribute.Value.Emit for canonical stringification of any value type. This feeds the
// gRPC metadata headers sent to ChipIngress via chipingress.WithResourceAttributeHeaders —
// resource attributes are not stamped as CloudEvent extensions.
func resourceAttributesToStringMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
