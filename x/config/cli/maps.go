package cli

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

const mapKVSep = "="

// Map entries merge by key at every source, so each occurrence adds to the config file's and the default's. Keys parse
// like values, so two texts that parse to one key, such as 16 and 0x10, are an error (see mapKey).
type textMapValue struct {
	// m holds only the entries given; decoding merges them over the file's and the default's.
	m    reflect.Value
	def  reflect.Value
	seen map[any]string
}

func (m *textMapValue) Set(s string) error {
	entries, err := parseList(s)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		k, v, ok := strings.Cut(entry, mapKVSep)
		if !ok {
			return fmt.Errorf("%q must be formatted as key%svalue", entry, mapKVSep)
		}

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

	return nil
}

func (m *textMapValue) String() string { return writeCSV(textMapOf(mergeMaps(m.m.Type(), m.def, m.m))) }

// Type is syntax rather than a Go type, so --help shows how to write it.
func (m *textMapValue) Type() string { return "key=value,..." }

// Sorted for stable --help output.
func textMapOf(v reflect.Value) []string {
	pairs := make([]string, 0, v.Len())
	for key, val := range v.Seq2() {
		pairs = append(pairs, textOf(key)+mapKVSep+textOf(val))
	}

	slices.Sort(pairs)
	return pairs
}

// A map read whole, by the markup or as a flag's text, is one value, so a later source replaces it.
func mergesByKey(t reflect.Type, lang markup.Markup) bool {
	return t.Kind() == reflect.Map && !lang.IsLeaf(t) && !readsText(t)
}

func mergeMaps(t reflect.Type, srcs ...reflect.Value) reflect.Value {
	out := reflect.MakeMap(t)
	for _, m := range srcs {
		for key, val := range m.Seq2() {
			out.SetMapIndex(key, val)
		}
	}

	return out
}
