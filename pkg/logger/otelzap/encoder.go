package otelzap

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

// otelAttrEncoder implements zapcore.ObjectEncoder to encode zap fields into OpenTelemetry attributes
type otelAttrEncoder struct {
	attributes []attribute.KeyValue
	namespace  string
}

// otelArrayEncoder implements zapcore.ArrayEncoder to collect array elements as strings
type otelArrayEncoder struct {
	elements []string
}

func (a *otelArrayEncoder) AppendString(v string) { a.elements = append(a.elements, v) }
func (a *otelArrayEncoder) AppendInt64(v int64) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendInt(v int) { a.elements = append(a.elements, fmt.Sprintf("%d", v)) }
func (a *otelArrayEncoder) AppendInt32(v int32) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendInt16(v int16) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendInt8(v int8) { a.elements = append(a.elements, fmt.Sprintf("%d", v)) }
func (a *otelArrayEncoder) AppendUint64(v uint64) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendUint32(v uint32) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendUint16(v uint16) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendUint8(v uint8) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendUint(v uint) { a.elements = append(a.elements, fmt.Sprintf("%d", v)) }
func (a *otelArrayEncoder) AppendUintptr(v uintptr) {
	a.elements = append(a.elements, fmt.Sprintf("%d", v))
}
func (a *otelArrayEncoder) AppendFloat64(v float64) {
	a.elements = append(a.elements, fmt.Sprintf("%g", v))
}
func (a *otelArrayEncoder) AppendFloat32(v float32) {
	a.elements = append(a.elements, fmt.Sprintf("%g", v))
}
func (a *otelArrayEncoder) AppendBool(v bool) { a.elements = append(a.elements, fmt.Sprintf("%t", v)) }
func (a *otelArrayEncoder) AppendArray(zapcore.ArrayMarshaler) error {
	a.elements = append(a.elements, "[nested array]")
	return nil
}
func (a *otelArrayEncoder) AppendObject(zapcore.ObjectMarshaler) error {
	a.elements = append(a.elements, "[object]")
	return nil
}
func (a *otelArrayEncoder) AppendReflected(v any) error {
	a.elements = append(a.elements, fmt.Sprintf("%+v", v))
	return nil
}
func (a *otelArrayEncoder) AppendByteString(v []byte) { a.elements = append(a.elements, string(v)) }
func (a *otelArrayEncoder) AppendComplex128(v complex128) {
	a.elements = append(a.elements, fmt.Sprintf("%v", v))
}
func (a *otelArrayEncoder) AppendComplex64(v complex64) {
	a.elements = append(a.elements, fmt.Sprintf("%v", v))
}
func (a *otelArrayEncoder) AppendDuration(v time.Duration) {
	a.elements = append(a.elements, v.String())
}
func (a *otelArrayEncoder) AppendTime(v time.Time) {
	a.elements = append(a.elements, v.Format(time.RFC3339))
}

func (e *otelAttrEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	// Create a simple array encoder that converts everything to strings
	encoder := &otelArrayEncoder{}
	err := marshaler.MarshalLogArray(encoder)
	if err != nil {
		return err
	}

	e.attributes = append(e.attributes, attribute.StringSlice(e.prefixKey(key), encoder.elements))
	return nil
}

func (e *otelAttrEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	// Create a nested encoder for the object
	objectEncoder := &otelAttrEncoder{}
	err := marshaler.MarshalLogObject(objectEncoder)
	if err != nil {
		return err
	}

	// Add all attributes from the object with the key as prefix
	for _, attr := range objectEncoder.attributes {
		prefixedKey := key + "." + string(attr.Key)
		e.attributes = append(e.attributes, attribute.KeyValue{
			Key:   attribute.Key(prefixedKey),
			Value: attr.Value,
		})
	}
	return nil
}

func (e *otelAttrEncoder) AddBinary(key string, value []byte) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), string(value)))
}

func (e *otelAttrEncoder) AddBool(key string, value bool) {
	e.attributes = append(e.attributes, attribute.Bool(e.prefixKey(key), value))
}

func (e *otelAttrEncoder) AddByteString(key string, value []byte) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), string(value)))
}

func (e *otelAttrEncoder) AddComplex128(key string, value complex128) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), fmt.Sprintf("%v", value)))
}

func (e *otelAttrEncoder) AddComplex64(key string, value complex64) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), fmt.Sprintf("%v", value)))
}

func (e *otelAttrEncoder) AddDuration(key string, value time.Duration) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddFloat64(key string, value float64) {
	e.attributes = append(e.attributes, attribute.Float64(e.prefixKey(key), value))
}

func (e *otelAttrEncoder) AddFloat32(key string, value float32) {
	e.attributes = append(e.attributes, attribute.Float64(e.prefixKey(key), float64(value)))
}

func (e *otelAttrEncoder) AddInt(key string, value int) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddInt64(key string, value int64) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), value))
}

func (e *otelAttrEncoder) AddInt32(key string, value int32) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddInt16(key string, value int16) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddInt8(key string, value int8) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddString(key string, value string) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), value))
}

func (e *otelAttrEncoder) AddTime(key string, value time.Time) {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), value.Format(time.RFC3339)))
}

func (e *otelAttrEncoder) AddUint(key string, value uint) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddUint64(key string, value uint64) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddUint32(key string, value uint32) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddUint16(key string, value uint16) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddUint8(key string, value uint8) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddUintptr(key string, value uintptr) {
	e.attributes = append(e.attributes, attribute.Int64(e.prefixKey(key), int64(value)))
}

func (e *otelAttrEncoder) AddReflected(key string, value any) error {
	e.attributes = append(e.attributes, attribute.String(e.prefixKey(key), fmt.Sprintf("%+v", value)))
	return nil
}

func (e *otelAttrEncoder) OpenNamespace(key string) {
	if e.namespace == "" {
		e.namespace = key
	} else {
		e.namespace = e.namespace + "." + key
	}
}

// helper method to apply namespace prefix to keys
func (e *otelAttrEncoder) prefixKey(key string) string {
	if e.namespace == "" {
		return key
	}
	return e.namespace + "." + key
}
