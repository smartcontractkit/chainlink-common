package json

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type loopDecoder struct {
	*json.Decoder
}

func newDecoder(r io.Reader) *loopDecoder {
	decoder := &loopDecoder{json.NewDecoder(r)}
	decoder.UseNumber()
	return decoder
}

func (d *loopDecoder) Decode(v any) error {
	targetVal := reflect.ValueOf(v)

	var rawData any
	if err := d.Decoder.Decode(&rawData); err != nil {
		return err
	}

	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("cannot decode into a non-pointer or nil value")
	}

	// Handling for null -> map or slice
	if rawData == nil && targetVal.Elem().Kind() == reflect.Map {
		// Set map to nil
		targetVal.Elem().Set(reflect.Zero(targetVal.Elem().Type()))
		return nil
	}
	if rawData == nil && targetVal.Elem().Kind() == reflect.Slice {
		// Set slice to nil
		targetVal.Elem().Set(reflect.Zero(targetVal.Elem().Type()))
		return nil
	}

	return d.unmarshal(reflect.ValueOf(rawData), targetVal)
}

func (d *loopDecoder) unmarshal(from reflect.Value, to reflect.Value) error {
	if !from.IsValid() {
		return nil
	}

	// Handle JSON `null` values
	if from.Kind() == reflect.Interface && from.IsNil() {
		if to.Elem().CanSet() {
			elemType := to.Elem().Type()
			to.Elem().Set(reflect.Zero(elemType))
		}
		return nil
	}

	if from.Kind() == reflect.Interface {
		from = from.Elem()
	}

	if to.Type().Elem() == reflect.TypeOf((*big.Int)(nil)) {
		bigInt := new(big.Int)
		var success bool

		switch v := from.Interface().(type) {
		case json.Number:
			_, success = bigInt.SetString(v.String(), 10)
		case string:
			_, success = bigInt.SetString(v, 10)
		default:
			return fmt.Errorf("cannot unmarshal %T into *big.Int", from.Interface())
		}

		if !success {
			return fmt.Errorf("failed to parse '%v' as *big.Int", from.Interface())
		}
		if to.Elem().CanSet() {
			to.Elem().Set(reflect.ValueOf(bigInt))
		}
		return nil
	}

	valueType := reflect.TypeOf((*values.Value)(nil)).Elem()
	if to.Type().Elem().Implements(valueType) {
		if str, ok := from.Interface().(string); ok {
			protoBytes, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return fmt.Errorf("failed to decode base64 for values.Value: %w", err)
			}

			pbValue := &pb.Value{}
			if err := proto.Unmarshal(protoBytes, pbValue); err != nil {
				return fmt.Errorf("failed to unmarshal protobuf for values.Value: %w", err)
			}

			val, err := values.FromProto(pbValue)
			if err != nil {
				return fmt.Errorf("failed to convert from proto to values.Value: %w", err)
			}

			if to.Elem().CanSet() {
				to.Elem().Set(reflect.ValueOf(val))
			}
			return nil
		}
	}

	toElem := to.Elem()

	if toElem.Type() == reflect.TypeOf(time.Time{}) {
		if str, ok := from.Interface().(string); ok {
			var t time.Time
			if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, str)), &t); err != nil {
				return fmt.Errorf("failed to unmarshal time.Time: %w", err)
			}
			toElem.Set(reflect.ValueOf(t))
			return nil
		}
	}

	switch toElem.Kind() {
	case reflect.Struct:
		return d.unmarshalStruct(from, toElem)
	case reflect.Slice:
		return d.unmarshalSlice(from, toElem)
	case reflect.Array:
		return d.unmarshalArray(from, toElem)
	case reflect.Map:
		return d.unmarshalMap(from, toElem)
	case reflect.Ptr:
		// Allocate a new pointer if needed
		if toElem.IsNil() {
			toElem.Set(reflect.New(toElem.Type().Elem()))
		}
		return d.unmarshal(from, toElem)
	default:
		return d.unmarshalPrimitive(from, toElem)
	}
}

func (d *loopDecoder) unmarshalStruct(from reflect.Value, to reflect.Value) error {
	fromMap, ok := from.Interface().(map[string]any)
	if !ok {
		return fmt.Errorf("expected object for struct, but got %T", from.Interface())
	}

	for i := 0; i < to.NumField(); i++ {
		fieldVal := to.Field(i)
		fieldType := to.Type().Field(i)
		if !fieldType.IsExported() || !fieldVal.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("json")
		if tag == "-" {
			continue
		}
		fieldName := strings.Split(tag, ",")[0]
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		if mapVal, ok := fromMap[fieldName]; ok {
			if err := d.unmarshal(reflect.ValueOf(mapVal), fieldVal.Addr()); err != nil {
				return fmt.Errorf("field %s: %w", fieldName, err)
			}
		}
	}
	return nil
}

func (d *loopDecoder) unmarshalSlice(from reflect.Value, to reflect.Value) error {
	// Special handling for []byte (base64 encoded string)
	if to.Type().Elem().Kind() == reflect.Uint8 {
		if str, ok := from.Interface().(string); ok {
			// The string is likely base64 encoded, let json handle it
			var b []byte
			if err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, str)), &b); err != nil {
				return fmt.Errorf("failed to decode base64 string: %w", err)
			}
			to.Set(reflect.ValueOf(b))
			return nil
		}
	}

	fromSlice, ok := from.Interface().([]any)
	if !ok {
		return fmt.Errorf("expected array for slice, but got %T", from.Interface())
	}
	newSlice := reflect.MakeSlice(to.Type(), len(fromSlice), len(fromSlice))
	for i := 0; i < len(fromSlice); i++ {
		if err := d.unmarshal(reflect.ValueOf(fromSlice[i]), newSlice.Index(i).Addr()); err != nil {
			return fmt.Errorf("slice index %d: %w", i, err)
		}
	}
	to.Set(newSlice)
	return nil
}

func (d *loopDecoder) unmarshalArray(from reflect.Value, to reflect.Value) error {
	// Special handling for byte arrays (base64 encoded in JSON)
	if to.Type().Elem().Kind() == reflect.Uint8 {
		// Check if we got a base64 string
		if str, ok := from.Interface().(string); ok {
			// Decode base64 into a temporary byte slice
			var tempSlice []byte
			err := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, str)), &tempSlice)
			if err != nil {
				return fmt.Errorf("failed to decode base64 for byte array: %w", err)
			}

			// Check length matches
			arrayLen := to.Type().Len()
			if len(tempSlice) != arrayLen {
				return fmt.Errorf("byte array size mismatch: expected %d bytes, got %d", arrayLen, len(tempSlice))
			}

			// Copy bytes to array
			for i := 0; i < arrayLen; i++ {
				to.Index(i).SetUint(uint64(tempSlice[i]))
			}
			return nil
		}
	}

	// Regular array handling
	fromSlice, ok := from.Interface().([]any)
	if !ok {
		return fmt.Errorf("expected array for array type, but got %T", from.Interface())
	}

	arrayLen := to.Type().Len()
	if len(fromSlice) != arrayLen {
		return fmt.Errorf("array size mismatch: expected %d, got %d", arrayLen, len(fromSlice))
	}

	for i := 0; i < arrayLen; i++ {
		if err := d.unmarshal(reflect.ValueOf(fromSlice[i]), to.Index(i).Addr()); err != nil {
			return fmt.Errorf("array index %d: %w", i, err)
		}
	}
	return nil
}

func (d *loopDecoder) unmarshalMap(from reflect.Value, to reflect.Value) error {
	fromMap, ok := from.Interface().(map[string]any)
	if !ok {
		return fmt.Errorf("expected object for map, but got %T", from.Interface())
	}
	newMap := reflect.MakeMap(to.Type())
	for key, val := range fromMap {
		newKey := reflect.ValueOf(key).Convert(to.Type().Key())
		newVal := reflect.New(to.Type().Elem())
		if err := d.unmarshal(reflect.ValueOf(val), newVal); err != nil {
			return err
		}
		newMap.SetMapIndex(newKey, newVal.Elem())
	}
	to.Set(newMap)
	return nil
}

func (d *loopDecoder) unmarshalPrimitive(from reflect.Value, to reflect.Value) error {
	// Special handling for byte slices (base64 encoded in JSON)
	if to.Kind() == reflect.Slice && to.Type().Elem().Kind() == reflect.Uint8 {
		// []byte is handled specially by json package
		if str, ok := from.Interface().(string); ok {
			// The string is likely base64 encoded, let json handle it
			return json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, str)), to.Addr().Interface())
		}
	}

	// Special handling for strings that might be numbers
	if from.Kind() == reflect.String {
		strVal := from.String()

		// First, try to handle numeric types
		switch to.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i64, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				// If it's not a number string but target is a string, just set it
				if to.Kind() == reflect.String {
					to.SetString(strVal)
					return nil
				}
				return fmt.Errorf("could not parse string \"%s\" to int: %w", strVal, err)
			}
			to.SetInt(i64)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u64, err := strconv.ParseUint(strVal, 10, 64)
			if err != nil {
				// If it's not a number string but target is a string, just set it
				if to.Kind() == reflect.String {
					to.SetString(strVal)
					return nil
				}
				return fmt.Errorf("could not parse string \"%s\" to uint: %w", strVal, err)
			}
			to.SetUint(u64)
			return nil
		case reflect.Float32, reflect.Float64:
			f64, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				// If it's not a number string but target is a string, just set it
				if to.Kind() == reflect.String {
					to.SetString(strVal)
					return nil
				}
				return fmt.Errorf("could not parse string \"%s\" to float: %w", strVal, err)
			}
			to.SetFloat(f64)
			return nil
		case reflect.Bool:
			b, err := strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("could not parse string \"%s\" to bool: %w", strVal, err)
			}
			to.SetBool(b)
			return nil
		case reflect.String:
			to.SetString(strVal)
			return nil
		}
	}

	// Handle json.Number to numeric types
	if num, ok := from.Interface().(json.Number); ok {
		switch to.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i64, err := num.Int64()
			if err != nil {
				return fmt.Errorf("could not convert number %s to int: %w", num, err)
			}
			to.SetInt(i64)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// json.Number doesn't have Uint64(), so parse string
			u64, err := strconv.ParseUint(num.String(), 10, 64)
			if err != nil {
				return fmt.Errorf("could not convert number %s to uint: %w", num, err)
			}
			to.SetUint(u64)
			return nil
		case reflect.Float32, reflect.Float64:
			f64, err := num.Float64()
			if err != nil {
				return fmt.Errorf("could not convert number %s to float: %w", num, err)
			}
			to.SetFloat(f64)
			return nil
		}
	}

	// Handle direct assignment (e.g., string to string, bool to bool)
	if from.IsValid() && from.Type().ConvertibleTo(to.Type()) {
		to.Set(from.Convert(to.Type()))
		return nil
	}

	return fmt.Errorf("unsupported conversion from %v (%s) to %s", from, from.Type(), to.Type())
}
