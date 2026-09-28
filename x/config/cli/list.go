package cli

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
)

// The text form of a list, for flags and env vars:
//
//	--ports 8080,8081             []int, both elements at once
//	--ports 8080 --ports 8081     []int, one element per occurrence
//	PORTS=8080,8081               the same, as an env var
//
// The first occurrence replaces the default; later ones append. Elements are CSV, as in pflag's
// slice flags and maps.go: --tags '"a,b",c' is two elements. Each is parsed as a flag of the
// element type would be, so --ports 80,x fails as the flag is parsed.

// textListValue is a list flag. Not pflag's StringSlice, so it shares quoting with textMapValue,
// parses any element type, and shows its syntax in --help.
type textListValue struct {
	// list is the elements parsed (see parseText), once set; elems is their text, for --help.
	list  reflect.Value
	elems []string
	// replaced: the first Set replaces the default, later ones append.
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

// String renders the --help default as it is typed. pflag's slice flags wrap theirs in brackets
// they don't accept.
func (l *textListValue) String() string { return writeCSV(l.elems) }

// Type spells out the syntax for --help rather than naming a Go type.
func (l *textListValue) Type() string { return "value,..." }

// parseList splits the text form of a list into its elements.
func parseList(s string) ([]string, error) {
	if s = strings.TrimSpace(s); s == "" {
		return nil, nil
	}
	return csv.NewReader(strings.NewReader(s)).Read()
}

func writeCSV(fields []string) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	// Writing to a strings.Builder cannot fail, so neither can this.
	_ = w.Write(fields)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

// textListOf renders v's elements as text, for the flag default.
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
