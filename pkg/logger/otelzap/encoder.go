package otelzap

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

// otelAttrEncoder implements zapcore.ObjectEncoder to encode zap fields into OpenTelemetry attributes
type otelAttrEncoder struct {
	attributes []attribute.KeyValue
}

func (e *otelAttrEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	// Not implemented
	return nil
}

func (e *otelAttrEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	// Not implemented
	return nil
}

func (e *otelAttrEncoder) AddBinary(key string, value []byte) {
	e.attributes = append(e.attributes, attribute.String(key, string(value)))
}

func (e *otelAttrEncoder) AddBool(key string, value bool) {
	e.attributes = append(e.attributes, attribute.Bool(key, value))
}

func (e *otelAttrEncoder) AddByteString(key string, value []byte) {
	e.attributes = append(e.attributes, attribute.String(key, string(value)))
}

func (e *otelAttrEncoder) AddComplex128(key string, value complex128) {
	// Not implemented
}

func (e *otelAttrEncoder) AddComplex64(key string, value complex64) {
	// Not implemented
}

func (e *otelAttrEncoder) AddDuration(key string, value time.Duration) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddFloat64(key string, value float64) {
	e.attributes = append(e.attributes, attribute.Float64(key, value))
}

func (e *otelAttrEncoder) AddFloat32(key string, value float32) {
	e.attributes = append(e.attributes, attribute.Float64(key, float64(value)))
}

func (e *otelAttrEncoder) AddInt(key string, value int) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddInt64(key string, value int64) {
	e.attributes = append(e.attributes, attribute.Int64(key, value))
}

func (e *otelAttrEncoder) AddInt32(key string, value int32) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddInt16(key string, value int16) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddInt8(key string, value int8) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddString(key, value string) {
	e.attributes = append(e.attributes, attribute.String(key, value))
}

func (e *otelAttrEncoder) AddTime(key string, value time.Time) {
	e.attributes = append(e.attributes, attribute.String(key, value.Format(time.RFC3339)))
}

func (e *otelAttrEncoder) AddUint(key string, value uint) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddUint64(key string, value uint64) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddUint32(key string, value uint32) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddUint16(key string, value uint16) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddUint8(key string, value uint8) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddUintptr(key string, value uintptr) {
	e.attributes = append(e.attributes, attribute.Int64(key, int64(value)))
}

func (e *otelAttrEncoder) AddReflected(key string, value any) error {
	// Not implemented
	return nil
}

func (e *otelAttrEncoder) OpenNamespace(key string) {
	// Not implemented
}
