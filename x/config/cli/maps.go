package cli

import (
	"encoding"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// The text form of a map, for flags and env vars:
//
//	--labels env=prod,region=us          map[string]string
//	--labels env=prod --labels region=us the same, one entry per occurrence
//	--chains 1=mainnet                   map[uint32]string, keys parse like values
//
// Entries are list.go's CSV fields split on "=". A map merges by key at every source, so each
// occurrence adds its entries over the config file's and the default's, a later key winning.
// Quote a whole entry to include a comma: --labels '"env=a,b",region=us'.
//
// Keys and values are parsed as list elements are, so two spellings of one key, such as 16 and
// 0x10, are an error (see mapKey).
const mapKVSep = "="

// textMapValue is a map flag. Not pflag's StringToString, which can't quote ',' or '=' and has
// only string values.
type textMapValue struct {
	// m is the entries supplied, parsed (see parseText); the decode merges them over the others.
	m reflect.Value
	// entries is the default's text and each entry supplied, for --help.
	entries map[string]string
	// seen is each key parsed, and its text.
	seen map[any]string
}

func (m *textMapValue) Set(s string) error {
	parsed, err := parseTextMap(s)
	if err != nil {
		return err
	}
	// Sorted, so a collision is reported the same way every run.
	for _, k := range slices.Sorted(maps.Keys(parsed)) {
		v := parsed[k]
		key, err := mapKey(m.m.Type().Key(), k, m.seen)
		if err != nil {
			return err
		}
		val, err := parseText(m.m.Type().Elem(), v)
		if err != nil {
			return fmt.Errorf("value of %q: %w", k, err)
		}
		m.m.SetMapIndex(key, val)
	}
	maps.Copy(m.entries, parsed)
	return nil
}

// String renders the --help default as it is typed, sorted for stable output.
func (m *textMapValue) String() string {
	pairs := make([]string, 0, len(m.entries))
	for k, v := range m.entries {
		pairs = append(pairs, k+mapKVSep+v)
	}
	slices.Sort(pairs)
	return writeCSV(pairs)
}

// Type spells out the syntax for --help rather than naming a Go type.
func (m *textMapValue) Type() string { return "key=value,..." }

// parseTextMap parses "k=v,k=v" into text.
func parseTextMap(s string) (map[string]string, error) {
	entries, err := parseList(s)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		k, val, ok := strings.Cut(entry, mapKVSep)
		if !ok {
			return nil, fmt.Errorf("%q must be formatted as key%svalue", entry, mapKVSep)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(val)
	}
	return out, nil
}

// textMapOf renders v's entries as text, for the flag default.
func textMapOf(v reflect.Value) map[string]string {
	out := map[string]string{}
	if v.IsValid() && v.Kind() == reflect.Map {
		for iter := v.MapRange(); iter.Next(); {
			out[textOf(iter.Key())] = textOf(iter.Value())
		}
	}
	return out
}

// mergeMaps is a new map of type t holding every entry of maps, a later map's winning a key.
func mergeMaps(t reflect.Type, maps ...reflect.Value) reflect.Value {
	out := reflect.MakeMap(t)
	for _, m := range maps {
		for iter := m.MapRange(); iter.Next(); {
			out.SetMapIndex(iter.Key(), iter.Value())
		}
	}
	return out
}

// textOf renders v as source text, for a flag default. A type's own marshaller renders it so the
// default parses back; anything else falls back to %v.
func textOf(v reflect.Value) string {
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		return ""
	}
	if text, ok := marshalText(reflect.Indirect(v)); ok {
		return text
	}
	return fmt.Sprint(v.Interface())
}

// marshalText renders v with its type's MarshalText. v is copied somewhere addressable first, so a
// method on the pointer receiver is reachable on a map key or element.
func marshalText(v reflect.Value) (string, bool) {
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	m, ok := p.Interface().(encoding.TextMarshaler)
	if !ok {
		return "", false
	}
	text, err := m.MarshalText()
	return string(text), err == nil
}
