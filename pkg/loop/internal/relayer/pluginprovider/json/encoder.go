package json

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type loopEncoder struct {
	*json.Encoder
}

func newEncoder(w io.Writer) *loopEncoder {
	return &loopEncoder{json.NewEncoder(w)}
}

// Encode pre-processes the data to convert numbers and values.Value types
// to strings before passing it to the standard JSON encoder.
func (e *loopEncoder) Encode(v any) error {
	converted, err := e.deepConvertToStrings(v)
	if err != nil {
		return err
	}
	return e.Encoder.Encode(converted)
}

func (e *loopEncoder) deepConvertToStrings(v any) (any, error) {
	if v == nil {
		return nil, nil
	}

	// Direct handling for *big.Int
	if bigInt, ok := v.(*big.Int); ok {
		if bigInt == nil {
			return nil, nil
		}
		return bigInt.String(), nil
	}

	// Direct handling for json.Number
	if num, ok := v.(json.Number); ok {
		return num.String(), nil
	}

	// Direct handling for time.Time - preserve as is for proper JSON marshaling
	if _, ok := v.(time.Time); ok {
		return v, nil
	}

	// Direct handling for byte slices - preserve as is for base64 encoding
	if byteSlice, ok := v.([]byte); ok {
		return byteSlice, nil
	}

	// Direct handling for values.Value - serialize as string
	if vv, ok := v.(values.Value); ok {
		pbValue := values.Proto(vv)
		data, err := proto.Marshal(pbValue)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal values.Value: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}

	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return nil, nil
	}

	if _, ok := v.(json.Marshaler); ok {
		return v, nil
	}
	if _, ok := v.(encoding.TextMarshaler); ok {
		return v, nil
	}

	typ := val.Type()
	if typ.Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) ||
		typ.Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
		return v, nil
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}

		// Deep convert the dereferenced value
		converted, err := e.deepConvertToStrings(val.Elem().Interface())
		if err != nil {
			return nil, err
		}

		// Check if the conversion changed the type
		// If the type didn't change, we can preserve the pointer
		origType := val.Elem().Type()
		convertedType := reflect.TypeOf(converted)
		if convertedType == origType {
			convertedVal := reflect.ValueOf(converted)
			newPtr := reflect.New(convertedVal.Type())
			newPtr.Elem().Set(convertedVal)
			return newPtr.Interface(), nil
		}

		// Otherwise return the dereferenced, converted value
		return converted, nil
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", val.Float()), nil
	case reflect.Struct:
		return e.convertStructToStrings(val)
	case reflect.Slice, reflect.Array:
		return e.convertSliceToStrings(val)
	case reflect.Map:
		return e.convertMapToStrings(val)
	default:
		return v, nil
	}
}

func (e *loopEncoder) convertStructToStrings(val reflect.Value) (any, error) {
	result := make(map[string]any)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		tag := fieldType.Tag.Get("json")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		fieldName := parts[0]
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// Handle omitempty
		if len(parts) > 1 && parts[1] == "omitempty" && field.IsZero() {
			continue
		}

		converted, err := e.deepConvertToStrings(field.Interface())
		if err != nil {
			return nil, err
		}
		result[fieldName] = converted
	}
	return result, nil
}

func (e *loopEncoder) convertSliceToStrings(val reflect.Value) (any, error) {
	result := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		converted, err := e.deepConvertToStrings(val.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		result[i] = converted
	}
	return result, nil
}

func (e *loopEncoder) convertMapToStrings(val reflect.Value) (any, error) {
	result := make(map[string]any)
	for _, key := range val.MapKeys() {
		converted, err := e.deepConvertToStrings(val.MapIndex(key).Interface())
		if err != nil {
			return nil, err
		}
		result[key.String()] = converted
	}
	return result, nil
}
