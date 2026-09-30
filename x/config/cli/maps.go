package cli

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// Map entries merge by key at every source, so each occurrence adds to the config file's and the default's. Keys parse
// like values, so two texts that parse to one key, such as 16 and 0x10, are an error (see mapKey).
const mapKVSep = "="

// Not pflag's StringToString, which can't quote ',' or '=' and has only string values.
type textMapValue struct {
	// m holds only the entries given; decoding merges them over the file's and the default's.
	m       reflect.Value
	entries map[string]string
	seen    map[any]string
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

// String is sorted for stable --help output.
func (m *textMapValue) String() string {
	pairs := make([]string, 0, len(m.entries))
	for k, v := range m.entries {
		pairs = append(pairs, k+mapKVSep+v)
	}

	slices.Sort(pairs)
	return writeCSV(pairs)
}

// Type shows syntax, not a Go type, is what --help should show.
func (m *textMapValue) Type() string { return "key=value,..." }

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

func textMapOf(v reflect.Value) map[string]string {
	out := map[string]string{}
	if v.IsValid() && v.Kind() == reflect.Map {
		for iter := v.MapRange(); iter.Next(); {
			out[textOf(iter.Key())] = textOf(iter.Value())
		}
	}

	return out
}

func mergeMaps(t reflect.Type, maps ...reflect.Value) reflect.Value {
	out := reflect.MakeMap(t)
	for _, m := range maps {
		for iter := m.MapRange(); iter.Next(); {
			out.SetMapIndex(iter.Key(), iter.Value())
		}
	}

	return out
}
