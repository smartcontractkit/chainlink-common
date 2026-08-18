package flags

import (
	"encoding"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// The textual form of a map value, on the command line and in an env var:
//
//	--labels 'env=prod;region=us'          map[string]string
//	--urls   'primary=a,b;backup=c'        map[string][]string
//
// Entries are separated by ';' and a list value by ','. This follows kong, the only widely used
// Go CLI library that binds maps of lists (its `mapsep` defaults to ';', its `sep` to ','); it
// is not pflag's own StringToString, which separates entries with ',' and so has no separator
// left for the elements of a list value. A ':' would collide with the values maps most often
// hold - a URL, a host:port, a timestamp - and no CLI uses it for this.
const (
	mapEntrySep = ";"
	mapKVSep    = "="
	// mapListSep splits a list value into its elements. It must stay in step with the
	// separator mapstructure's StringToSliceHookFunc is configured with in stringCoercionHooks,
	// which is what actually splits the value during decoding.
	mapListSep = ","
)

// textMapValue is the pflag.Value backing a map flag. It holds the entries as text and leaves
// converting them to the map's real key and value types to the decoder, so a flag, an env var
// and a config file all reach that decoder the same way.
//
// pflag's own StringToString is deliberately not used: viper special-cases that flag type and
// parses its value itself with a comma separator (see stringToStringConv), which would both
// disagree with the syntax here and mangle a list value.
type textMapValue struct {
	entries map[string]string
	// listValued records that each value is a list, so a comma in it separates elements rather
	// than being the mistake parseTextMap otherwise reports.
	listValued bool
	// replaced tracks whether Set has been called: the first call replaces the struct's
	// compiled-in default, and later ones merge into it, so repeating the flag accumulates
	// (--labels a=1 --labels b=2) the way repeating a slice flag does.
	replaced bool
}

func newTextMapValue(def map[string]string, listValued bool) *textMapValue {
	return &textMapValue{entries: def, listValued: listValued}
}

func (m *textMapValue) Set(s string) error {
	parsed, err := parseTextMap(s, m.listValued)
	if err != nil {
		return err
	}
	if !m.replaced {
		m.entries, m.replaced = map[string]string{}, true
	}
	for k, v := range parsed {
		m.entries[k] = v
	}
	return nil
}

// String renders the entries in the form parseTextMap reads back, since viper hands this string
// to the decoder verbatim for any flag type it doesn't know. Entries are sorted so --help and
// generated docs don't reorder between runs.
func (m *textMapValue) String() string {
	if len(m.entries) == 0 {
		return "[]"
	}
	pairs := make([]string, 0, len(m.entries))
	for k, v := range m.entries {
		pairs = append(pairs, k+mapKVSep+m.quote(v))
	}
	sort.Strings(pairs)
	return "[" + strings.Join(pairs, mapEntrySep) + "]"
}

// quote wraps a value that would otherwise read back as something else, since this text is
// re-parsed: viper hands a flag's own string to the decoder rather than the entries behind it.
// A comma only needs quoting where it isn't already the list separator.
func (m *textMapValue) quote(v string) string {
	if strings.ContainsAny(v, mapEntrySep+`"`) || (!m.listValued && strings.Contains(v, mapListSep)) {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}

// Type is what --help prints after the flag name, so it spells out the syntax rather than
// naming a Go type nobody has to know. It must not be one of the flag types viper parses
// itself ("stringToString", "stringSlice", ...) - see the type doc.
func (m *textMapValue) Type() string {
	if m.listValued {
		return "key=v1,v2;..."
	}
	return "key=value;..."
}

// parseTextMap parses the "k=v;k=v" text form of a map into its entries. Values are left as
// text; the decoder converts each one to the map's actual value type, including splitting a
// list value on mapListSep. listValued says whether the target's values are lists, which is
// what makes a comma meaningful inside a value rather than a mistake.
func parseTextMap(s string, listValued bool) (map[string]string, error) {
	s = strings.TrimSpace(s)
	// The bracketed form is what String() prints, so a value read back from a flag's own
	// default (which viper does when nothing else set the key) round-trips.
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}

	out := map[string]string{}
	if strings.TrimSpace(s) == "" {
		return out, nil
	}

	for _, entry := range splitEntries(s) {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		k, val, ok := strings.Cut(entry, mapKVSep)
		if !ok {
			return nil, fmt.Errorf("%q must be formatted as key%svalue", entry, mapKVSep)
		}
		k, val = strings.TrimSpace(k), strings.TrimSpace(val)

		switch {
		case isQuoted(val):
			// An explicitly quoted value is taken whole, which is the way out for a scalar
			// value that really does contain a comma.
			val = strings.ReplaceAll(val[1:len(val)-1], `\"`, `"`)
		case listValued:
			// A list may also be bracketed, as AWS CLI shorthand writes one, so that a value
			// reads as a list even where the flag's own syntax already implies it.
			if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
				val = val[1 : len(val)-1]
			}
		default:
			if err := checkNotCommaSeparatedPairs(entry, val); err != nil {
				return nil, err
			}
		}

		out[k] = val
	}
	return out, nil
}

// splitEntries splits the entries on mapEntrySep, ignoring a separator inside a quoted value
// so that a value holding one survives the trip out through String and back.
func splitEntries(s string) []string {
	var (
		entries []string
		cur     strings.Builder
		quoted  bool
		escaped bool
	)
	for _, r := range s {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			quoted = !quoted
		case !quoted && string(r) == mapEntrySep:
			entries = append(entries, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteRune(r)
	}
	return append(entries, cur.String())
}

// checkNotCommaSeparatedPairs rejects a scalar map value that looks like it holds further
// key=value pairs - "--labels env=prod,region=us", the comma-separated form pflag's
// StringToString and kubectl use. Splitting on the comma here would be wrong for a list-valued
// map, so the syntax can't just accept both; storing "prod,region=us" as one value silently
// would be worse than saying so.
func checkNotCommaSeparatedPairs(entry, val string) error {
	parts := strings.Split(val, mapListSep)
	for _, part := range parts[1:] {
		if !strings.Contains(part, mapKVSep) {
			continue
		}
		return fmt.Errorf("%q: separate map entries with %q, not %q (quote the value to keep a comma in it)",
			entry, mapEntrySep, mapListSep)
	}
	return nil
}

func isQuoted(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}

// textMapOf renders a map value as the entries a map flag holds, so a struct's compiled-in
// default survives into the flag's default and into --help. Returns nil for a nil or empty map,
// which prints as "[]".
func textMapOf(v reflect.Value) map[string]string {
	if !v.IsValid() || v.Kind() != reflect.Map || v.Len() == 0 {
		return nil
	}
	out := make(map[string]string, v.Len())
	for iter := v.MapRange(); iter.Next(); {
		out[textOf(iter.Key())] = textOf(iter.Value())
	}
	return out
}

// textOf renders a single map key or value in the form a flag carries it, which is the form the
// decoder reads back: a TextMarshaler's own text, a duration as "1m30s", a list as its elements
// joined by mapListSep, anything else as its default formatting.
func textOf(v reflect.Value) string {
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		return ""
	}

	// Map keys and values are not addressable, so a MarshalText declared on the pointer
	// receiver is invisible on the value itself; copy into an addressable pointer to find it.
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	if tm, ok := p.Interface().(encoding.TextMarshaler); ok {
		if b, err := tm.MarshalText(); err == nil {
			return string(b)
		}
	}

	if v.Type() == durationType {
		return time.Duration(v.Int()).String()
	}

	if k := v.Kind(); k == reflect.Slice || k == reflect.Array {
		elems := make([]string, v.Len())
		for i := range elems {
			elems[i] = textOf(v.Index(i))
		}
		return strings.Join(elems, mapListSep)
	}

	return fmt.Sprintf("%v", v.Interface())
}
