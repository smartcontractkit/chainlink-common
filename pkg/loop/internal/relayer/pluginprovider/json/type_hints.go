package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// defaultMaxTypeHintDepth is the default maximum recursion depth for GetTypeHint
const defaultMaxTypeHintDepth = 80

// getTypeHint returns a standard Go type hint for the given value.
// It uses the default maximum recursion depth to prevent infinite recursion.
// See getTypeHintWithDepth for more details.
func getTypeHint(v any) (string, error) {
	return getTypeHintWithDepth(v, defaultMaxTypeHintDepth)
}

// getTypeHintWithDepth returns a standard Go type hint for the given value with a maximum recursion depth.
// The depth parameter prevents infinite recursion in circular references.
// When depth reaches 0, complex types return "..."
//
// Only returns hints for standard types that any implementation can understand.
// Returns an error if the type cannot be determined.
//
// Supported type hints:
//   - Basic types: "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64"
//   - Pointer types: "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64"
//   - Float types: "float32", "float64", "*float32", "*float64"
//   - Other basics: "bool", "string", "*bool", "*string"
//   - Byte arrays: "[]byte" (normalized from []uint8)
//   - Slices: "[]int", "[]string", "[]bool", etc.
//   - Slices with element hints: "[]any{int,string,bool}"
//   - Maps: "map[string]any", "map[string]string", "map[string]int", etc.
//   - Maps with field hints: "map[string]any{field1=int64,field2=[]byte,nested=map[string]any{inner=uint64}}"
//   - Special types: "*big.Int", "time.Time"
//   - Nil: "nil" (for nil values)
//   - Depth exceeded: "..." (when recursion depth is exceeded)
//
// The type hint system resolves JSON ambiguities:
//   - Numbers: Without hints, all numbers become json.Number or float64
//   - []byte: Without hints, base64 strings remain as strings
//   - Type precision: int32 vs int64 vs uint64 cannot be distinguished without hints
//   - Pointers: JSON doesn't distinguish between values and pointers, type hints preserve this information
func getTypeHintWithDepth(v any, depth int) (string, error) {
	// Check depth limit
	if depth <= 0 {
		return "", errors.New("max recursion depth exceeded")
	}

	t := reflect.TypeOf(v)
	// this is an `any` that is nil.
	if t == nil {
		return "nil", nil
	}

	// Check for special types first (before pointer handling)
	switch v.(type) {
	case *big.Int:
		return "*big.Int", nil
	case big.Int:
		return "*big.Int", nil // Always treat big.Int as *big.Int
	case time.Time:
		return "time.Time", nil
	case *time.Time:
		return "time.Time", nil // Always treat *time.Time as time.Time
	case values.Value:
		// Even if nil, it's still a values.Value type
		return "values.Value", nil
	}

	// Handle pointers - preserve pointer information
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()

		// Check if pointer to values.Value interface - this is an error
		valueType := reflect.TypeOf((*values.Value)(nil)).Elem()
		if elem.Implements(valueType) {
			return "", fmt.Errorf("pointer to values.Value interface is not supported - values.Value is already an interface type")
		}

		// Build pointer prefix (handle nested pointers like **int, ***int)
		pointerPrefix := ""
		currentType := t
		for currentType.Kind() == reflect.Ptr {
			pointerPrefix += "*"
			currentType = currentType.Elem()
		}

		// For nil pointer, we need to determine the type from the type itself
		if reflect.ValueOf(v).IsNil() {
			// Handle *any specially - for single pointer to any, return nil
			if elem.Kind() == reflect.Interface && elem.NumMethod() == 0 && pointerPrefix == "*" {
				return "nil", nil
			}
			// Handle nested pointers to any
			if elem.Kind() == reflect.Interface && elem.NumMethod() == 0 {
				return pointerPrefix + "any", nil
			}
			
			// Create a zero value of the pointed-to type to get its hint
			zeroValue := reflect.Zero(elem).Interface()
			elemHint, err := getTypeHintWithDepth(zeroValue, depth-1)
			if err != nil {
				return "", fmt.Errorf("failed to get type hint for nil pointer element: %w", err)
			}
			return pointerPrefix + elemHint, nil
		}

		// For non-nil pointers, we can handle specific cases more accurately
		switch elem.Kind() {
		case reflect.Int:
			return pointerPrefix + "int", nil
		case reflect.Int8:
			return pointerPrefix + "int8", nil
		case reflect.Int16:
			return pointerPrefix + "int16", nil
		case reflect.Int32:
			return pointerPrefix + "int32", nil
		case reflect.Int64:
			return pointerPrefix + "int64", nil
		case reflect.Uint:
			return pointerPrefix + "uint", nil
		case reflect.Uint8:
			return pointerPrefix + "uint8", nil
		case reflect.Uint16:
			return pointerPrefix + "uint16", nil
		case reflect.Uint32:
			return pointerPrefix + "uint32", nil
		case reflect.Uint64:
			return pointerPrefix + "uint64", nil
		case reflect.Float32:
			return pointerPrefix + "float32", nil
		case reflect.Float64:
			return pointerPrefix + "float64", nil
		case reflect.Bool:
			return pointerPrefix + "bool", nil
		case reflect.String:
			return pointerPrefix + "string", nil
		case reflect.Slice, reflect.Array:
			// Handle pointer to slice/array
			zeroValue := reflect.Zero(elem).Interface()
			elemHint, err := getTypeHintWithDepth(zeroValue, depth-1)
			if err != nil {
				return "", fmt.Errorf("failed to get type hint for pointer to slice: %w", err)
			}
			return pointerPrefix + elemHint, nil
		case reflect.Interface:
			// For pointer to any (*any), handle specially
			if elem.NumMethod() == 0 {
				// For single pointer to any, dereference and get the type of the contained value
				if pointerPrefix == "*" {
					// Dereference the pointer and get the actual value inside the any
					deref := reflect.ValueOf(v).Elem().Interface()
					return getTypeHintWithDepth(deref, depth-1)
				}
				// For nested pointers to any (e.g., **any, ***any), return the full type
				return pointerPrefix + "any", nil
			}
			// For other interfaces, dereference and get the actual value
			deref := reflect.ValueOf(v).Elem().Interface()
			return getTypeHintWithDepth(deref, depth-1)
		case reflect.Struct:
			// For pointer to struct, dereference and handle as struct
			deref := reflect.ValueOf(v).Elem().Interface()
			return getTypeHintWithDepth(deref, depth-1)
		case reflect.Ptr:
			// Nested pointer - we've already counted all the pointer levels
			// Get the base type (non-pointer)
			baseType := currentType
			
			// Create a zero value of the base type
			zeroValue := reflect.Zero(baseType).Interface()
			baseHint, err := getTypeHintWithDepth(zeroValue, depth-1)
			if err != nil {
				return "", fmt.Errorf("failed to get type hint for nested pointer: %w", err)
			}
			return pointerPrefix + baseHint, nil
		default:
			// For any other pointer types, get the element type hint
			zeroValue := reflect.Zero(elem).Interface()
			elemHint, err := getTypeHintWithDepth(zeroValue, depth-1)
			if err != nil {
				return "", fmt.Errorf("failed to get type hint for pointer element: %w", err)
			}
			return pointerPrefix + elemHint, nil
		}
	}

	// Handle non-pointer types
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		// Arrays and slices are treated the same for type hints
		elem := t.Elem()

		// Special case: normalize []uint8 to []byte
		if elem.Kind() == reflect.Uint8 {
			return "[]byte", nil
		}

		// Special case for []any - include element type hints
		if elem.Kind() == reflect.Interface && elem.NumMethod() == 0 {
			s, ok := v.([]any)
			if !ok {
				return "", fmt.Errorf("type assertion to []any failed: %T", elem)
			}
			if len(s) == 0 {
				return "[]any{}", nil
			}
			elementHints := make([]string, 0, len(s))
			for _, elem := range s {
				hint, err := getTypeHintWithDepth(elem, depth-1)
				if err != nil {
					return "", fmt.Errorf("failed to get type hint for slice element: %w", err)
				}
				elementHints = append(elementHints, hint)
			}
			return fmt.Sprintf("[]any{%s}", strings.Join(elementHints, ",")), nil
		}

		// General recursive handling for all other slice/array types
		// For nested slices of any, we need to check the actual values
		if isNestedAnySlice(t) {
			// This is [][]any, [][][]any, or deeper nesting
			return getNestedAnySliceTypeHint(v, t, depth)
		}

		// Get the element type hint directly from the type
		elemHint, err := getTypeHintFromType(elem)
		if err != nil {
			return "", fmt.Errorf("failed to get type hint for slice element type: %w", err)
		}
		return "[]" + elemHint, nil
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			elem := t.Elem()
			if elem.Kind() == reflect.Interface && elem.NumMethod() == 0 {
				if m, ok := v.(map[string]any); ok {
					// Try to build detailed type hint with field types
					fieldHints := make([]string, 0, len(m))
					for k, v := range m {
						hint, err := getTypeHintWithDepth(v, depth-1)
						if err != nil {
							return "", fmt.Errorf("failed to get type hint for map key %s: %w", k, err)
						}
						fieldHints = append(fieldHints, fmt.Sprintf("%s=%s", k, hint))
					}
					// Sort for consistent output
					sort.Strings(fieldHints)
					return fmt.Sprintf("map[string]any{%s}", strings.Join(fieldHints, ",")), nil
				}
			}
			// Handle other common map types
			switch elem.Kind() {
			case reflect.String:
				return "map[string]string", nil
			case reflect.Int:
				return "map[string]int", nil
			case reflect.Int64:
				return "map[string]int64", nil
			case reflect.Float64:
				return "map[string]float64", nil
			case reflect.Bool:
				return "map[string]bool", nil
			}
		}
	case reflect.Int:
		return "int", nil
	case reflect.Int8:
		return "int8", nil
	case reflect.Int16:
		return "int16", nil
	case reflect.Int32:
		return "int32", nil
	case reflect.Int64:
		return "int64", nil
	case reflect.Uint:
		return "uint", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint64:
		return "uint64", nil
	case reflect.Float32:
		return "float32", nil
	case reflect.Float64:
		return "float64", nil
	case reflect.Bool:
		return "bool", nil
	case reflect.String:
		return "string", nil
	case reflect.Struct:
		// Check depth before processing struct
		if depth <= 1 {
			return "", errors.New("max recursion depth exceeded")
		}

		// Convert struct to map[string]any with type hints for each field
		// We do this manually instead of with mapstructure in order to parse `json`
		// tags correctly, because the value accompanying this type hint will
		// be serialized using json and field renames would be applied.
		structValue := reflect.ValueOf(v)
		structType := structValue.Type()

		// Build the type hint string manually to preserve field types
		var fieldHintParts []string
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldValue := structValue.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get json tag if exists
			fieldName := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				// Handle json:"-" (ignore field)
				if jsonTag == "-" {
					continue
				}
				// Parse json tag (handle "name,omitempty" format)
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "omitempty" {
					fieldName = parts[0]
				}
				// Check for omitempty
				hasOmitEmpty := false
				for _, part := range parts {
					if part == "omitempty" {
						hasOmitEmpty = true
						break
					}
				}
				if hasOmitEmpty && fieldValue.IsZero() {
					continue // Skip zero values with omitempty
				}
			}

			// Get type hint for the field value
			var fieldTypeHint string

			// Special handling for interface fields
			if field.Type.Kind() == reflect.Interface {
				// Check if it's values.Value
				valueType := reflect.TypeOf((*values.Value)(nil)).Elem()
				if field.Type.Implements(valueType) {
					fieldTypeHint = "values.Value"
				} else if fieldValue.IsNil() {
					// For nil any, we can't determine the type, so it's just nil
					// This is different from typed interfaces like values.Value
					fieldTypeHint = "nil"
				} else {
					// Get hint from the actual value
					hint, err := getTypeHintWithDepth(fieldValue.Interface(), depth-1)
					if err != nil {
						return "", err
					}
					fieldTypeHint = hint
				}
			} else {
				// Get hint from the field value
				hint, err := getTypeHintWithDepth(fieldValue.Interface(), depth-1)
				if err != nil {
					return "", err
				}
				fieldTypeHint = hint
			}

			fieldHintParts = append(fieldHintParts, fmt.Sprintf("%s=%s", fieldName, fieldTypeHint))
		}

		// Sort for consistent output
		sort.Strings(fieldHintParts)
		return fmt.Sprintf("map[string]any{%s}", strings.Join(fieldHintParts, ",")), nil
	}

	// Cannot determine type hint for unsupported types
	return "", fmt.Errorf("unsupported type %T: no type hint available", v)
}

// parseMapTypeHint parses map type hints with field information.
// Format: "map[string]any{field1=type1,field2=type2,nested=map[string]any{inner=type3}}"
// Returns (fieldHints, true) if valid map hint with fields, or (nil, false) if not.
func parseMapTypeHint(hint string) (map[string]string, bool) {
	if !strings.HasPrefix(hint, "map[string]any{") || !strings.HasSuffix(hint, "}") {
		return nil, false
	}

	// Extract content between braces
	content := hint[len("map[string]any{") : len(hint)-1]
	if content == "" {
		return map[string]string{}, true // Empty map
	}

	fieldHints := make(map[string]string)

	// Parse field=type pairs, handling nested maps
	var field strings.Builder
	var typeHint strings.Builder
	braceCount := 0
	inType := false

	for i, ch := range content {
		switch ch {
		case '=':
			if braceCount == 0 && !inType {
				inType = true
				continue
			}
		case '{':
			braceCount++
		case '}':
			braceCount--
		case ',':
			if braceCount == 0 {
				// End of field
				if field.Len() > 0 && typeHint.Len() > 0 {
					fieldHints[field.String()] = typeHint.String()
				}
				field.Reset()
				typeHint.Reset()
				inType = false
				continue
			}
		}

		// Add character to field or type
		if inType {
			typeHint.WriteRune(ch)
		} else {
			field.WriteRune(ch)
		}

		// Handle last field
		if i == len(content)-1 && field.Len() > 0 && typeHint.Len() > 0 {
			fieldHints[field.String()] = typeHint.String()
		}
	}

	return fieldHints, true
}

// parseSliceTypeHint parses slice type hints with element information.
// Format: "[]any{type1,type2,nested=[]any{type3,type4}}" or "[][]any{[]any{int},[]any{bool}}"
// Returns (elementHints, true) if valid slice hint with elements, or (nil, false) if not.
func parseSliceTypeHint(hint string) ([]string, bool) {
	// Check for nested []any patterns (e.g., [][]any{...}, [][][]any{...})
	prefix := ""
	remaining := hint
	for strings.HasPrefix(remaining, "[]") {
		prefix += "[]"
		remaining = remaining[2:]
	}

	if !strings.HasPrefix(remaining, "any{") || !strings.HasSuffix(remaining, "}") {
		return nil, false
	}

	// Extract content between braces
	content := remaining[len("any{") : len(remaining)-1]
	if content == "" {
		return []string{}, true // Empty slice
	}

	elementHints := []string{}

	// Parse type elements, handling nested structures
	var currentType strings.Builder
	braceCount := 0

	for i, ch := range content {
		switch ch {
		case '{':
			braceCount++
			currentType.WriteRune(ch)
		case '}':
			braceCount--
			currentType.WriteRune(ch)
		case ',':
			if braceCount == 0 {
				// End of element
				if currentType.Len() > 0 {
					elementHints = append(elementHints, currentType.String())
					currentType.Reset()
				}
				continue
			}
			currentType.WriteRune(ch)
		default:
			currentType.WriteRune(ch)
		}

		// Handle last element
		if i == len(content)-1 && currentType.Len() > 0 {
			elementHints = append(elementHints, currentType.String())
		}
	}

	return elementHints, true
}

// UnmarshalWithHint unmarshals JSON data using the provided type hint.
// Returns unmarshaled data as any for unknown hints or complex types.
//
// Type hint handling:
//   - Known types are unmarshaled to their exact Go types
//   - Pointer types: "*int32", "*string", etc. unmarshal to pointers with nil handling
//   - Slices with element hints: "[]any{type1,type2,...}" preserves exact types
//   - Maps with field hints: "map[string]any{field1=type1,field2=type2}"
//   - Null values are handled appropriately (nil for pointers/slices/maps, zero for values)
//   - Unknown hints fall back to generic unmarshal with UseNumber() for numeric precision
//   - All unmarshaling uses UseNumber() to preserve numeric precision
//
// Special cases:
//   - "*big.Int": null becomes nil, not zero value
//   - Pointer types: null JSON values become nil pointers, non-null values become pointers to values
//   - "nil" hint: null returns nil, non-null unmarshals as any
//   - Nested maps: "map[string]any{nested=map[string]any{inner=int64}}"
//   - Nested slices: "[]any{[]any{int,int},string}"
func UnmarshalWithHint(data []byte, hint string) (any, error) {
	if hint == "" {
		return nil, errors.New("empty type hint")
	}
	// Check if this is a map type hint with field information
	if fieldHints, ok := parseMapTypeHint(hint); ok {
		// First unmarshal to generic map
		var m map[string]any
		if err := UnmarshalJson(data, &m); err != nil {
			return nil, err
		}

		// Apply field type hints
		for field, fieldHint := range fieldHints {
			if val, exists := m[field]; exists {
				// Handle nil values - they should remain nil
				if val == nil {
					continue // Keep nil as nil
				}

				// Special handling for *big.Int when value is json.Number
				if fieldHint == "*big.Int" {
					if num, ok := val.(json.Number); ok {
						// Convert json.Number directly to *big.Int
						bi := new(big.Int)
						if _, success := bi.SetString(num.String(), 10); success {
							m[field] = bi
							continue
						}
					}
				}

				// Re-marshal the field value
				fieldData, err := MarshalJson(val)
				if err != nil {
					continue // Skip fields that can't be marshaled
				}

				// Check if the marshaled data is "null"
				if string(fieldData) == "null" {
					m[field] = nil
					continue
				}

				// Unmarshal with the field's type hint
				typedValue, err := UnmarshalWithHint(fieldData, fieldHint)
				if err == nil {
					m[field] = typedValue
				}
			}
		}

		return m, nil
	}

	// Check if this is a slice type hint with element information
	if elementHints, ok := parseSliceTypeHint(hint); ok {
		// First unmarshal to generic slice
		var s []any
		if err := UnmarshalJson(data, &s); err != nil {
			return nil, err
		}

		// Apply element type hints
		for i := 0; i < len(s) && i < len(elementHints); i++ {
			if s[i] == nil {
				continue // Keep nil as nil
			}

			// Re-marshal the element value
			elemData, err := MarshalJson(s[i])
			if err != nil {
				continue // Skip elements that can't be marshaled
			}

			// Check if the marshaled data is "null"
			if string(elemData) == "null" {
				s[i] = nil
				continue
			}

			// Unmarshal with the element's type hint
			typedValue, err := UnmarshalWithHint(elemData, elementHints[i])
			if err == nil {
				s[i] = typedValue
			}
		}

		// For nested []any types, we need to return the properly typed slice
		// Extract the nesting level from the hint
		nesting := 0
		remaining := hint
		for strings.HasPrefix(remaining, "[]") {
			nesting++
			remaining = remaining[2:]
		}

		// If it's a nested []any (more than one level), create the proper type
		if nesting > 1 && strings.HasPrefix(remaining, "any{") {
			return createNestedAnySlice(s, nesting)
		}

		return s, nil
	}

	switch hint {
	case "nil":
		// For nil params, we expect null as the only accepted value.
		if string(data) != "null" {
			return nil, fmt.Errorf("unexpected value for null type hint: %s", data)
		}
		return nil, nil
	case "[]byte":
		var b []byte
		err := UnmarshalJson(data, &b)
		return b, err
	// Pointer types
	case "*int":
		if string(data) == "null" {
			return (*int)(nil), nil
		}
		var i int
		err := UnmarshalJson(data, &i)
		return &i, err
	case "*int8":
		if string(data) == "null" {
			return (*int8)(nil), nil
		}
		var i int8
		err := UnmarshalJson(data, &i)
		return &i, err
	case "*int16":
		if string(data) == "null" {
			return (*int16)(nil), nil
		}
		var i int16
		err := UnmarshalJson(data, &i)
		return &i, err
	case "*int32":
		if string(data) == "null" {
			return (*int32)(nil), nil
		}
		var i int32
		err := UnmarshalJson(data, &i)
		return &i, err
	case "*int64":
		if string(data) == "null" {
			return (*int64)(nil), nil
		}
		var i int64
		err := UnmarshalJson(data, &i)
		return &i, err
	case "*uint":
		if string(data) == "null" {
			return (*uint)(nil), nil
		}
		var u uint
		err := UnmarshalJson(data, &u)
		return &u, err
	case "*uint8":
		if string(data) == "null" {
			return (*uint8)(nil), nil
		}
		var u uint8
		err := UnmarshalJson(data, &u)
		return &u, err
	case "*uint16":
		if string(data) == "null" {
			return (*uint16)(nil), nil
		}
		var u uint16
		err := UnmarshalJson(data, &u)
		return &u, err
	case "*uint32":
		if string(data) == "null" {
			return (*uint32)(nil), nil
		}
		var u uint32
		err := UnmarshalJson(data, &u)
		return &u, err
	case "*uint64":
		if string(data) == "null" {
			return (*uint64)(nil), nil
		}
		var u uint64
		err := UnmarshalJson(data, &u)
		return &u, err
	case "*float32":
		if string(data) == "null" {
			return (*float32)(nil), nil
		}
		var f float32
		err := UnmarshalJson(data, &f)
		return &f, err
	case "*float64":
		if string(data) == "null" {
			return (*float64)(nil), nil
		}
		var f float64
		err := UnmarshalJson(data, &f)
		return &f, err
	case "*bool":
		if string(data) == "null" {
			return (*bool)(nil), nil
		}
		var b bool
		err := UnmarshalJson(data, &b)
		return &b, err
	case "*string":
		if string(data) == "null" {
			return (*string)(nil), nil
		}
		var s string
		err := UnmarshalJson(data, &s)
		return &s, err
	// Non-pointer types
	case "int":
		var i int
		err := UnmarshalJson(data, &i)
		return i, err
	case "int8":
		var i int8
		err := UnmarshalJson(data, &i)
		return i, err
	case "int16":
		var i int16
		err := UnmarshalJson(data, &i)
		return i, err
	case "int32":
		var i int32
		err := UnmarshalJson(data, &i)
		return i, err
	case "int64":
		var i int64
		err := UnmarshalJson(data, &i)
		return i, err
	case "uint":
		var u uint
		err := UnmarshalJson(data, &u)
		return u, err
	case "uint8":
		var u uint8
		err := UnmarshalJson(data, &u)
		return u, err
	case "uint16":
		var u uint16
		err := UnmarshalJson(data, &u)
		return u, err
	case "uint32":
		var u uint32
		err := UnmarshalJson(data, &u)
		return u, err
	case "uint64":
		var u uint64
		err := UnmarshalJson(data, &u)
		return u, err
	case "float32":
		var f float32
		err := UnmarshalJson(data, &f)
		return f, err
	case "float64":
		var f float64
		err := UnmarshalJson(data, &f)
		return f, err
	case "bool":
		var b bool
		err := UnmarshalJson(data, &b)
		return b, err
	case "string":
		var s string
		err := UnmarshalJson(data, &s)
		return s, err
	case "*big.Int":
		if string(data) == "null" {
			return (*big.Int)(nil), nil
		}
		var bi *big.Int
		if err := UnmarshalJson(data, &bi); err == nil {
			return bi, nil
		}
		return nil, fmt.Errorf("failed to unmarshal %s as *big.Int", string(data))
	case "time.Time":
		var t time.Time
		err := UnmarshalJson(data, &t)
		return t, err
	case "map[string]string":
		var m map[string]string
		err := UnmarshalJson(data, &m)
		return m, err
	case "map[string]int":
		var m map[string]int
		err := UnmarshalJson(data, &m)
		return m, err
	case "map[string]int64":
		var m map[string]int64
		err := UnmarshalJson(data, &m)
		return m, err
	case "map[string]float64":
		var m map[string]float64
		err := UnmarshalJson(data, &m)
		return m, err
	case "map[string]bool":
		var m map[string]bool
		err := UnmarshalJson(data, &m)
		return m, err
	case "values.Value":
		// Just unmarshal to values.Value, the decoder will handle the special format
		var val values.Value
		err := UnmarshalJson(data, &val)
		return val, err
	default:
		// Check if it's a slice type (starts with [])
		if strings.HasPrefix(hint, "[]") {
			// Extract the element type hint
			elemHint := hint[2:]

			// First unmarshal to []any
			var slice []any
			if err := UnmarshalJson(data, &slice); err != nil {
				return nil, err
			}

			// Create a typed slice based on the element hint
			result, err := createTypedSlice(slice, elemHint)
			if err != nil {
				return nil, err
			}
			return result, nil
		}

		// Check if it's a pointer type (starts with *)
		if strings.HasPrefix(hint, "*") {
			// Check if data is null
			if string(data) == "null" {
				// Create a nil pointer of the correct type
				return createNilPointer(hint)
			}

			// Extract the pointed-to type
			innerHint := hint[1:]
			
			// Unmarshal the inner value
			innerValue, err := UnmarshalWithHint(data, innerHint)
			if err != nil {
				return nil, err
			}

			// Create a pointer to the value
			return createPointer(innerValue, hint)
		}
	}

	return nil, fmt.Errorf("unknown type hint: %s", hint)
}

// createTypedSlice converts a []any to a properly typed slice based on the element type hint
func createTypedSlice(slice []any, elemHint string) (any, error) {
	// First, process all elements to get their typed values
	typedElements := make([]any, len(slice))
	for i, elem := range slice {
		// Handle nil elements
		if elem == nil {
			typedElements[i] = nil
			continue
		}

		// Re-marshal the element
		elemData, err := MarshalJson(elem)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal slice element %d: %w", i, err)
		}

		// Unmarshal with the element type hint
		typedElem, err := UnmarshalWithHint(elemData, elemHint)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal slice element %d with hint %s: %w", i, elemHint, err)
		}
		typedElements[i] = typedElem
	}

	// Now create the properly typed slice using reflection
	return createSliceFromHint(typedElements, elemHint)
}

// createSliceFromHint creates a typed slice from elements and a type hint
func createSliceFromHint(elements []any, hint string) (any, error) {
	// Determine the slice type from the hint
	var sliceType reflect.Type

	switch hint {
	case "byte":
		sliceType = reflect.TypeOf([]byte{})
	case "int":
		sliceType = reflect.TypeOf([]int{})
	case "int32":
		sliceType = reflect.TypeOf([]int32{})
	case "int64":
		sliceType = reflect.TypeOf([]int64{})
	case "uint32":
		sliceType = reflect.TypeOf([]uint32{})
	case "uint64":
		sliceType = reflect.TypeOf([]uint64{})
	case "float32":
		sliceType = reflect.TypeOf([]float32{})
	case "float64":
		sliceType = reflect.TypeOf([]float64{})
	case "bool":
		sliceType = reflect.TypeOf([]bool{})
	case "string":
		sliceType = reflect.TypeOf([]string{})
	default:
		// Handle nested slices
		if strings.HasPrefix(hint, "[]") {
			// This is a nested slice, we need to create the type dynamically
			if len(elements) == 0 {
				// For empty nested slices, we need to construct the type
				return createEmptyNestedSlice(hint)
			}

			// Get the type from the first non-nil element
			var elemType reflect.Type
			for _, elem := range elements {
				if elem != nil {
					elemType = reflect.TypeOf(elem)
					break
				}
			}

			if elemType == nil {
				// All elements are nil, create empty slice of the correct type
				return createEmptyNestedSlice(hint)
			}

			sliceType = reflect.SliceOf(elemType)
		} else if strings.HasPrefix(hint, "*") {
			// Handle pointer types like *int, *[]byte, etc.
			// Get the type from the hint
			elemType, err := getTypeFromHint(hint)
			if err != nil {
				return nil, fmt.Errorf("failed to get type for slice element hint %s: %w", hint, err)
			}
			sliceType = reflect.SliceOf(elemType)
		} else {
			// Unknown type, keep as []any
			return elements, nil
		}
	}

	// Create the slice
	result := reflect.MakeSlice(sliceType, len(elements), len(elements))

	// Set the elements
	for i, elem := range elements {
		if elem != nil {
			result.Index(i).Set(reflect.ValueOf(elem))
		}
	}

	return result.Interface(), nil
}

// createEmptyNestedSlice creates an empty slice of the correct nested type
func createEmptyNestedSlice(hint string) (any, error) {
	// Count the nesting level
	nesting := 0
	for strings.HasPrefix(hint, "[]") {
		nesting++
		hint = hint[2:]
	}

	// Determine the base type
	var baseType reflect.Type
	switch hint {
	case "byte":
		baseType = reflect.TypeOf(byte(0))
	case "int":
		baseType = reflect.TypeOf(int(0))
	case "int32":
		baseType = reflect.TypeOf(int32(0))
	case "int64":
		baseType = reflect.TypeOf(int64(0))
	case "uint32":
		baseType = reflect.TypeOf(uint32(0))
	case "uint64":
		baseType = reflect.TypeOf(uint64(0))
	case "float32":
		baseType = reflect.TypeOf(float32(0))
	case "float64":
		baseType = reflect.TypeOf(float64(0))
	case "bool":
		baseType = reflect.TypeOf(bool(false))
	case "string":
		baseType = reflect.TypeOf(string(""))
	default:
		// Unknown type, use any
		baseType = reflect.TypeOf((*any)(nil)).Elem()
	}

	// Build the nested slice type
	sliceType := baseType
	for i := 0; i < nesting; i++ {
		sliceType = reflect.SliceOf(sliceType)
	}

	// Create an empty slice of that type
	return reflect.MakeSlice(sliceType, 0, 0).Interface(), nil
}

// MarshalWithHint marshals a value to JSON and returns both the JSON bytes and the type hint.
// This ensures the type hint is always generated for the exact value being marshaled.
//
// Returns:
//   - data: JSON-encoded bytes using the custom encoder (UseNumber, etc.)
//   - hint: Type hint string for the value
//   - error: Any error from marshaling or type hint generation
//
// This is the preferred way to marshal values that will later be unmarshaled with UnmarshalWithHint.
func MarshalWithHint(v any) ([]byte, string, error) {
	// Get the type hint first
	hint, err := getTypeHint(v)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get type hint: %w", err)
	}

	// Marshal the value
	data, err := MarshalJson(v)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal value: %w", err)
	}

	return data, hint, nil
}

// isNestedAnySlice checks if a type represents a nested slice of any (e.g., [][]any, [][][]any)
func isNestedAnySlice(t reflect.Type) bool {
	// Must be a slice/array
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
		return false
	}

	elem := t.Elem()
	// Check if element is also a slice/array
	if elem.Kind() != reflect.Slice && elem.Kind() != reflect.Array {
		return false
	}

	// Keep going down until we find the base element
	for elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
		elem = elem.Elem()
	}

	// Check if the base element is any
	return elem.Kind() == reflect.Interface && elem.NumMethod() == 0
}

// getNestedAnySliceTypeHint generates type hints for nested slices of any
func getNestedAnySliceTypeHint(v any, t reflect.Type, depth int) (string, error) {
	// Count the nesting level
	nesting := 0
	elem := t
	for elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
		nesting++
		elem = elem.Elem()
	}

	// Build the prefix (e.g., "[][]" for 2 levels of nesting)
	prefix := strings.Repeat("[]", nesting)

	// Get the actual slice value
	slice := reflect.ValueOf(v)
	if slice.Len() == 0 {
		return prefix + "any{}", nil
	}

	// Get type hints for each element
	elementHints := make([]string, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		elemVal := slice.Index(i).Interface()
		hint, err := getTypeHintWithDepth(elemVal, depth-1)
		if err != nil {
			return "", fmt.Errorf("failed to get type hint for nested any slice element: %w", err)
		}
		elementHints = append(elementHints, hint)
	}

	return fmt.Sprintf("%sany{%s}", prefix, strings.Join(elementHints, ",")), nil
}

func createNestedAnySlice(elements []any, nesting int) (any, error) {
	if nesting <= 1 {
		return elements, nil
	}

	sliceType := reflect.TypeOf((*interface{})(nil)).Elem()

	for i := 0; i < nesting; i++ {
		sliceType = reflect.SliceOf(sliceType)
	}

	result := reflect.MakeSlice(sliceType, len(elements), len(elements))

	for i, elem := range elements {
		if elem != nil {
			result.Index(i).Set(reflect.ValueOf(elem))
		}
	}

	return result.Interface(), nil
}

// createNilPointer creates a nil pointer of the specified type from a type hint
func createNilPointer(hint string) (any, error) {
	// Count pointer depth
	pointerDepth := 0
	remaining := hint
	for strings.HasPrefix(remaining, "*") {
		pointerDepth++
		remaining = remaining[1:]
	}

	// Get the base type
	baseType, err := getTypeFromHint(remaining)
	if err != nil {
		return nil, fmt.Errorf("failed to get base type for nil pointer: %w", err)
	}

	// Build up the pointer type
	ptrType := baseType
	for i := 0; i < pointerDepth; i++ {
		ptrType = reflect.PtrTo(ptrType)
	}

	// Return a nil pointer of that type
	return reflect.Zero(ptrType).Interface(), nil
}

// createPointer creates a pointer to the given value, handling nested pointers
func createPointer(value any, hint string) (any, error) {
	// Just create a single pointer to the value
	// The value has already been unmarshaled with the inner hint
	ptr := reflect.New(reflect.TypeOf(value))
	ptr.Elem().Set(reflect.ValueOf(value))
	return ptr.Interface(), nil
}

// getTypeHintFromType generates a type hint string from a reflect.Type
func getTypeHintFromType(t reflect.Type) (string, error) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		pointerPrefix := ""
		currentType := t
		for currentType.Kind() == reflect.Ptr {
			pointerPrefix += "*"
			currentType = currentType.Elem()
		}
		
		// Get the base type hint
		baseHint, err := getTypeHintFromType(currentType)
		if err != nil {
			return "", err
		}
		return pointerPrefix + baseHint, nil
	}

	// Handle slices and arrays
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		elem := t.Elem()
		// Special case: normalize []uint8 to []byte
		if elem.Kind() == reflect.Uint8 {
			return "[]byte", nil
		}
		
		// Recursively get element type hint
		elemHint, err := getTypeHintFromType(elem)
		if err != nil {
			return "", err
		}
		return "[]" + elemHint, nil
	}

	// Handle basic types
	switch t.Kind() {
	case reflect.Int:
		return "int", nil
	case reflect.Int8:
		return "int8", nil
	case reflect.Int16:
		return "int16", nil
	case reflect.Int32:
		return "int32", nil
	case reflect.Int64:
		return "int64", nil
	case reflect.Uint:
		return "uint", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint64:
		return "uint64", nil
	case reflect.Float32:
		return "float32", nil
	case reflect.Float64:
		return "float64", nil
	case reflect.Bool:
		return "bool", nil
	case reflect.String:
		return "string", nil
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return "any", nil
		}
		// Check for special interface types
		if t.Implements(reflect.TypeOf((*values.Value)(nil)).Elem()) {
			return "values.Value", nil
		}
		return "", fmt.Errorf("unsupported interface type: %v", t)
	case reflect.Struct:
		// Check for special struct types
		if t == reflect.TypeOf(time.Time{}) {
			return "time.Time", nil
		}
		if t == reflect.TypeOf(big.Int{}) {
			return "*big.Int", nil
		}
		// For other structs, we can't determine the hint without a value
		return "struct", nil
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			elem := t.Elem()
			switch elem.Kind() {
			case reflect.String:
				return "map[string]string", nil
			case reflect.Int:
				return "map[string]int", nil
			case reflect.Int64:
				return "map[string]int64", nil
			case reflect.Bool:
				return "map[string]bool", nil
			case reflect.Float64:
				return "map[string]float64", nil
			case reflect.Interface:
				if elem.NumMethod() == 0 {
					return "map[string]any", nil
				}
			}
		}
		return "", fmt.Errorf("unsupported map type: %v", t)
	default:
		return "", fmt.Errorf("unsupported type: %v", t)
	}
}

// getTypeFromHint returns a reflect.Type from a type hint string
func getTypeFromHint(hint string) (reflect.Type, error) {
	switch hint {
	case "nil":
		return reflect.TypeOf((*interface{})(nil)).Elem(), nil
	case "any":
		return reflect.TypeOf((*interface{})(nil)).Elem(), nil
	case "int":
		return reflect.TypeOf(int(0)), nil
	case "int8":
		return reflect.TypeOf(int8(0)), nil
	case "int16":
		return reflect.TypeOf(int16(0)), nil
	case "int32":
		return reflect.TypeOf(int32(0)), nil
	case "int64":
		return reflect.TypeOf(int64(0)), nil
	case "uint":
		return reflect.TypeOf(uint(0)), nil
	case "uint8":
		return reflect.TypeOf(uint8(0)), nil
	case "uint16":
		return reflect.TypeOf(uint16(0)), nil
	case "uint32":
		return reflect.TypeOf(uint32(0)), nil
	case "uint64":
		return reflect.TypeOf(uint64(0)), nil
	case "float32":
		return reflect.TypeOf(float32(0)), nil
	case "float64":
		return reflect.TypeOf(float64(0)), nil
	case "bool":
		return reflect.TypeOf(bool(false)), nil
	case "string":
		return reflect.TypeOf(string("")), nil
	case "[]byte":
		return reflect.TypeOf([]byte{}), nil
	case "*big.Int":
		return reflect.TypeOf((*big.Int)(nil)), nil
	case "time.Time":
		return reflect.TypeOf(time.Time{}), nil
	case "values.Value":
		return reflect.TypeOf((*values.Value)(nil)).Elem(), nil
	default:
		// Check if it's a slice type
		if strings.HasPrefix(hint, "[]") {
			elemHint := hint[2:]
			elemType, err := getTypeFromHint(elemHint)
			if err != nil {
				return nil, err
			}
			return reflect.SliceOf(elemType), nil
		}
		
		// Check if it's a pointer type  
		if strings.HasPrefix(hint, "*") {
			elemHint := hint[1:]
			elemType, err := getTypeFromHint(elemHint)
			if err != nil {
				return nil, err
			}
			return reflect.PtrTo(elemType), nil
		}

		// Check if it's a map type
		if strings.HasPrefix(hint, "map[string]") {
			switch hint {
			case "map[string]string":
				return reflect.TypeOf(map[string]string{}), nil
			case "map[string]int":
				return reflect.TypeOf(map[string]int{}), nil
			case "map[string]int64":
				return reflect.TypeOf(map[string]int64{}), nil
			case "map[string]bool":
				return reflect.TypeOf(map[string]bool{}), nil
			case "map[string]float64":
				return reflect.TypeOf(map[string]float64{}), nil
			default:
				if strings.Contains(hint, "map[string]any{") {
					return reflect.TypeOf(map[string]any{}), nil
				}
			}
		}

		return nil, fmt.Errorf("unknown type hint: %s", hint)
	}
}
