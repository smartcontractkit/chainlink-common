package cli

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Lists are CSV, as pflag's slice flags are, so an element can be quoted: --tags '"a,b",c' is two elements. The first
// occurrence replaces the default and later ones append, so a list can be replaced or built up. pflag has a slice flag
// only for a fixed set of element types; this takes any element type with a text form.
type textListValue struct {
	// list starts as the default, which only shows in --help: decoding reads it only once the flag is set.
	list     reflect.Value
	replaced bool
}

func (l *textListValue) Set(s string) error {
	elems, err := parseList(s)
	if err != nil {
		return err
	}

	if !l.replaced {
		l.list, l.replaced = reflect.MakeSlice(l.list.Type(), 0, len(elems)), true
	}

	for _, elem := range elems {
		v, err := parseText(l.list.Type().Elem(), elem)
		if err != nil {
			return fmt.Errorf("element %q: %w", elem, err)
		}

		l.list = reflect.Append(l.list, v)
	}

	return nil
}

// String has no brackets, unlike pflag's slice flags, so the default reads as it would be typed.
func (l *textListValue) String() string { return writeCSV(textListOf(l.list)) }

// Type is syntax rather than a Go type, so --help shows how to write it.
func (l *textListValue) Type() string { return "value,..." }

func parseList(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}

	r := csv.NewReader(strings.NewReader(s))
	fields, err := r.Read()
	if err != nil {
		return nil, err
	}

	// A second record would otherwise be dropped silently.
	if _, err = r.Read(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%q must be one line; quote an element that holds a newline", s)
	}

	return fields, nil
}

func writeCSV(fields []string) string {
	// A lone empty field would write as an empty line, which parseList reads as no elements.
	if len(fields) == 1 && fields[0] == "" {
		return `""`
	}

	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write(fields)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

func textListOf(v reflect.Value) []string {
	out := make([]string, v.Len())
	for i := range out {
		out[i] = textOf(v.Index(i))
	}

	return out
}
