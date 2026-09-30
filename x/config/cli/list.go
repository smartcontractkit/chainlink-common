package cli

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
)

// Lists are CSV, as pflag's slice flags are, so an element can be quoted: --tags '"a,b",c' is two elements. The first
// occurrence replaces the default and later ones append, so a list can be replaced or built up.

// Not pflag's StringSlice, so it quotes like map flags, takes any text element type, and shows its syntax in --help.
type textListValue struct {
	list     reflect.Value
	elems    []string
	replaced bool
}

func (l *textListValue) Set(s string) error {
	elems, err := parseList(s)
	if err != nil {
		return err
	}

	if !l.replaced {
		l.list, l.elems, l.replaced = reflect.MakeSlice(l.list.Type(), 0, len(elems)), nil, true
	}

	for _, elem := range elems {
		v, err := parseText(l.list.Type().Elem(), elem)
		if err != nil {
			return fmt.Errorf("element %q: %w", elem, err)
		}

		l.list = reflect.Append(l.list, v)
	}

	l.elems = append(l.elems, elems...)
	return nil
}

// String has no brackets, unlike pflag's slice flags, so the default reads as it would be typed.
func (l *textListValue) String() string { return writeCSV(l.elems) }

// Type shows syntax, not a Go type, is what --help should show.
func (l *textListValue) Type() string { return "value,..." }

func parseList(s string) ([]string, error) {
	if s = strings.TrimSpace(s); s == "" {
		return nil, nil
	}

	return csv.NewReader(strings.NewReader(s)).Read()
}

func writeCSV(fields []string) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write(fields)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

func textListOf(v reflect.Value) []string {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}

		v = v.Elem()
	}

	if !v.IsValid() || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return nil
	}

	out := make([]string, v.Len())
	for i := range out {
		out[i] = textOf(v.Index(i))
	}

	return out
}
