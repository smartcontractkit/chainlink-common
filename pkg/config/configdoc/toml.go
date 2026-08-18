package configdoc

import (
	"encoding"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TOML documents TOML configuration: `# comment`, `[table]`, `key = value # Default`.
type TOML struct{}

func (t TOML) Name() string {
	return "toml"
}

var _ Format = TOML{}

func (TOML) Comment(text string) string { return "# " + text }

func (TOML) Table(path string) string { return "[" + path + "]" }

func (TOML) Field(key, value, marker string) string {
	if marker == "" {
		return fmt.Sprintf("%s = %s", key, value)
	}
	return fmt.Sprintf("%s = %s %s", key, value, marker)
}

func (TOML) DefaultMarker() string { return FieldDefault }

func (TOML) ExampleMarker() string { return FieldExample }

func (TOML) DocsOnlyMarker() string { return FieldDocsOnly }

func (TOML) ParseLine(line string) Line {
	switch {
	case strings.HasPrefix(line, "#"):
		return Line{Kind: LineComment, Text: strings.TrimSpace(line[1:])}
	case strings.TrimSpace(line) == "":
		return Line{Kind: LineBlank}
	case strings.HasPrefix(line, "[["):
		return Line{Kind: LineArrayOfTables, Text: strings.Trim(strings.Trim(line, FieldExample), "[]")}
	case strings.HasPrefix(line, "["):
		return Line{Kind: LineTable, Text: strings.Trim(line, "[]")}
	default:
		name := strings.TrimSpace(line)
		if i := strings.Index(name, " "); i > -1 {
			name = name[:i]
		}
		return Line{Kind: LineField, Text: name}
	}
}

var textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

// Literal renders v as a TOML value literal. Values that marshal themselves to text (such as
// config.Duration) are written as their text form in quotes, so they round-trip through the
// same UnmarshalText that reads them back.
func (t TOML) Literal(val any) string {
	if val == nil {
		return "''"
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "''"
		}
		v = v.Elem()
	}

	if v.CanInterface() && v.Type().Implements(textMarshalerType) {
		if b, err := v.Interface().(encoding.TextMarshaler).MarshalText(); err == nil {
			return "'" + string(b) + "'"
		}
	}

	if v.Type() == reflect.TypeOf(time.Duration(0)) {
		return "'" + time.Duration(v.Int()).String() + "'"
	}

	switch v.Kind() {
	case reflect.String:
		return "'" + v.String() + "'"
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Slice, reflect.Array:
		parts := make([]string, v.Len())
		for i := range parts {
			parts[i] = t.Literal(v.Index(i).Interface())
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case reflect.Map:
		// An inline table, so the value stays on the field's own line the way every other
		// literal here does - a [table] header would have to be written somewhere else in the
		// document. Keys are sorted, or the generated document would differ run to run.
		if v.Len() == 0 {
			return "{}"
		}
		parts := make([]string, 0, v.Len())
		for _, k := range v.MapKeys() {
			// A TOML key is itself a string, and Literal quotes one, so the key renders the
			// same way the value does - 'env' = 'prod'.
			parts = append(parts, t.Literal(k.Interface())+" = "+t.Literal(v.MapIndex(k).Interface()))
		}
		sort.Strings(parts)
		return "{ " + strings.Join(parts, ", ") + " }"
	default:
		return fmt.Sprintf("'%v'", v.Interface())
	}
}
