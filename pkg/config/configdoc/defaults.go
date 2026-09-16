package configdoc

import (
	"bufio"
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// DefaultsOnly reads only the default values from a docs TOML file and decodes in to cfg.
// Fields without defaults will set to zero values.
// Arrays of tables are ignored.
func DefaultsOnly(r io.Reader, cfg any, decode func(io.Reader, any) error) error {
	pr, pw := io.Pipe()
	defer pr.Close()
	go writeDefaults(r, pw)
	if err := decode(pr, cfg); err != nil {
		return fmt.Errorf("failed to decode default core configuration: %w", err)
	}
	// replace niled examples with zero values.
	nilToZero(reflect.ValueOf(cfg))
	return nil
}

// writeDefaults writes default lines from defaultsTOML to w.
func writeDefaults(r io.Reader, w *io.PipeWriter) {
	defer w.Close()
	s := bufio.NewScanner(r)
	var skipTable string
	for s.Scan() {
		t := s.Text()
		// Indentation is ignored in TOML, and nested tables and their keys are conventionally indented.
		trimmed := strings.TrimSpace(t)

		if skipTable != "" {
			if !strings.HasPrefix(trimmed, "[") {
				// TOML does not end a table at a blank line, only at the next table header,
				// so keys - and blank lines - still belong to the array of tables.
				continue
			}
			if name := tableName(trimmed); name == skipTable || strings.HasPrefix(name, skipTable+".") {
				continue
			}
			skipTable = ""
		}

		// Skip arrays of tables
		if strings.HasPrefix(trimmed, "[[") {
			skipTable = tableName(trimmed)
			continue
		}

		// Skip comments and examples (which become zero values)
		if strings.HasPrefix(trimmed, "#") || strings.HasSuffix(trimmed, FieldExample) {
			continue
		}
		if _, err := io.WriteString(w, t); err != nil {
			w.CloseWithError(err)
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			w.CloseWithError(err)
		}
	}
	if err := s.Err(); err != nil {
		w.CloseWithError(fmt.Errorf("failed to scan core defaults: %w", err))
	}
}

var textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

func nilToZero(val reflect.Value) {
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			t := val.Type().Elem()
			val.Set(reflect.New(t))
		}
		if val.Type().Implements(textUnmarshalerType) {
			return // don't descend inside - leave whole zero value
		}
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Struct:
		if val.Type().Implements(textUnmarshalerType) {
			return // skip values unmarshaled from strings
		}
		for _, f := range val.Fields() {
			nilToZero(f)
		}
		return
	case reflect.Map:
		if !val.IsNil() {
			for _, k := range val.MapKeys() {
				nilToZero(val.MapIndex(k))
			}
		}
		return
	case reflect.Slice:
		if !val.IsNil() {
			for i := 0; i < val.Len(); i++ {
				nilToZero(val.Index(i))
			}
		}
		return
	default:
		return
	}
}
