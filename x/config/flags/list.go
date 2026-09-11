package flags

import (
	"reflect"
	"strings"
)

// The textual form of a list, on the command line and in an env var. A list of any depth has one,
// so a list of lists is writable without a config file:
//
//	--ports 8080,8081             []int, both elements at once
//	--ports 8080 --ports 8081     []int, one element per occurrence
//	--urls 'a,b' --urls 'c'       [][]string, one inner list per occurrence
//	--urls '[a,b],[c]'            [][]string, the whole value at once
//	URLS='[a,b],[c]'              the same, as an env var, which can't be repeated
//
// Elements are separated by mapListSep (','), the same separator a map's list value uses, and a
// nested list is written in brackets. Repeating the flag accumulates, the way pflag's own
// StringSlice does: the first occurrence replaces the struct's compiled-in default and later ones
// append.
//
// An element that contains a separator, a bracket or a quote is written in double quotes -
// --tags '"a,b",c' is two elements - which is also how a scalar keeps a comma that would otherwise
// split it.
const (
	listOpen  = '['
	listClose = ']'
)

// isTextLeafType reports whether a value of type t is carried as a single element of a list: a
// text value, or a byte slice, which is text (a JSON blob, a PEM key) rather than a list of
// numbers and is converted from a string wholesale by the decoder's weak typing.
func isTextLeafType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if isTextValueType(t) {
		return true
	}
	return (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) && t.Elem().Kind() == reflect.Uint8
}

// listDepthOf reports how many list levels t has above the text values at the bottom, and whether
// it bottoms out in text values at all. A []string is depth 1, a [][]config.Duration depth 2, a
// string depth 0; a []SomeStruct or a []map[string]string is not a text list, since its elements
// have no single-element text form (see nested.go for how a list of structs is carried instead).
func listDepthOf(t reflect.Type) (int, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if isTextLeafType(t) {
		return 0, true
	}
	if k := t.Kind(); k == reflect.Slice || k == reflect.Array {
		if d, ok := listDepthOf(t.Elem()); ok {
			return d + 1, true
		}
	}
	return 0, false
}

// textListValue is the pflag.Value backing a list flag of any depth. Like textMapValue it holds
// the value as text and leaves converting the elements to their real Go types to the decoder, so a
// flag, an env var and a config file all reach that decoder the same way.
//
// pflag's own StringSlice is deliberately not used, even for a plain []string: viper special-cases
// that flag type and re-parses its value as CSV, which has no way to express a nested list, so the
// one flag kind would behave differently at depth 1 and depth 2.
type textListValue struct {
	// depth is the number of list levels this flag's target has - 1 for []string, 2 for a
	// [][]string or for a []string leaf reached through a list of structs.
	depth int
	// elems holds one tree per top-level element, each of depth-1 levels: a string at the bottom,
	// a []any above it.
	elems []any
	// replaced tracks whether Set has been called, so the first occurrence replaces the
	// compiled-in default and later ones append to it.
	replaced bool
}

func newTextListValue(def any, depth int) *textListValue {
	l := &textListValue{depth: depth}
	if elems, ok := def.([]any); ok {
		l.elems = elems
	}
	return l
}

// Set appends the occurrence's elements. What counts as one element depends on how deeply the
// occurrence is written: at depth 1 a comma separates elements, so --tags a,b adds two, while
// deeper the comma separates the elements of a single nested list, so --urls a,b adds the one
// element [a,b] and only the explicitly nested --urls '[a,b],[c]' adds two. That way repeating the
// flag builds a list of lists one inner list at a time, and writing the whole thing at once - the
// only form an env var can take - still works.
func (l *textListValue) Set(s string) error {
	if !l.replaced {
		l.elems, l.replaced = nil, true
	}
	l.elems = append(l.elems, splitListOccurrence(s, l.depth)...)
	return nil
}

// String renders the elements in the form parseList reads back, since viper hands a flag's own
// string to the decoder for any flag type it doesn't parse itself.
func (l *textListValue) String() string { return renderList(l.elems) }

// Type is what --help prints after the flag name: the syntax rather than a Go type. It must not be
// one of the names viper parses itself ("stringSlice", "intSlice", ...) - see the type doc.
func (l *textListValue) Type() string {
	inner := "value"
	for i := 1; i < l.depth; i++ {
		inner = "[" + inner + ",...]"
	}
	return inner + ",..."
}

// splitListOccurrence splits one flag occurrence into the top-level elements it contributes to a
// list of the given depth, which is decided by how deeply the occurrence itself is nested: written
// as deep as the target it is the whole value, and written one level shallower it is a single
// element of it. That is what lets the two forms coexist - at depth 2 "a,b" and "[a,b]" each add
// the one element [a,b] while "[a,b],[c]" adds both - and it stays unambiguous at depth 3, where a
// single element is itself a list of lists.
func splitListOccurrence(s string, depth int) []any {
	s = strings.TrimSpace(s)
	if s == "" || s == string(listOpen)+string(listClose) {
		return nil
	}
	if nestingOf(s) >= depth {
		inner, _ := trimListBrackets(s)
		return parseListParts(splitList(inner), depth-1)
	}
	return []any{parseList(s, depth-1)}
}

// nestingOf reports how many list levels the text of s spells out: 0 for a bare element, 1 for
// "a,b" or "[a,b]", 2 for "[a,b],[c]". A separator inside brackets or quotes belongs to a level
// further down and doesn't count here.
func nestingOf(s string) int {
	s = strings.TrimSpace(s)
	inner, wrapped := trimListBrackets(s)
	parts := splitList(inner)
	if !wrapped && len(parts) < 2 {
		return 0
	}
	deepest := 0
	for _, p := range parts {
		if d := nestingOf(p); d > deepest {
			deepest = d
		}
	}
	return deepest + 1
}

func parseListParts(parts []string, depth int) []any {
	if len(parts) == 1 && strings.TrimSpace(parts[0]) == "" {
		return []any{}
	}
	out := make([]any, len(parts))
	for i, p := range parts {
		out[i] = parseList(p, depth)
	}
	return out
}

// parseList parses s as a list value of the given depth, into the tree of strings the rest of this
// file works with: a string at depth 0, a []any above it. One enclosing bracket pair is optional
// at every level, so both "a,b" and "[a,b]" read as the same list - the bracketed form is what
// renderList writes, and what viper hands back as a flag's own default.
func parseList(s string, depth int) any {
	s = strings.TrimSpace(s)
	if depth == 0 {
		return unquoteListElem(s)
	}
	inner, _ := trimListBrackets(s)
	return parseListParts(splitList(inner), depth-1)
}

// renderList renders a tree of strings back into the bracketed text form parseList reads.
func renderList(x any) string {
	elems, ok := x.([]any)
	if !ok {
		s, _ := x.(string)
		return quoteListElem(s)
	}
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = renderList(e)
	}
	return string(listOpen) + strings.Join(parts, mapListSep) + string(listClose)
}

// treeOf converts a Go value holding depth list levels into the same tree of strings, so a
// struct's compiled-in default survives into the flag's default and into --help.
func treeOf(v reflect.Value, depth int) any {
	if depth == 0 {
		return textOf(v)
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return []any{}
		}
		v = v.Elem()
	}
	if !v.IsValid() || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return []any{}
	}
	out := make([]any, v.Len())
	for i := range out {
		out[i] = treeOf(v.Index(i), depth-1)
	}
	return out
}

// trimListBrackets strips one enclosing bracket pair from s, reporting whether the whole string
// was wrapped in one - "[a],[b]" is not, even though it starts with '[' and ends with ']'.
func trimListBrackets(s string) (string, bool) {
	if len(s) < 2 || s[0] != listOpen || s[len(s)-1] != listClose {
		return s, false
	}
	depth := 0
	var quoted, escaped bool
	for i, r := range s {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			quoted = !quoted
		case quoted:
		case r == listOpen:
			depth++
		case r == listClose:
			depth--
			// A close that balances the opening bracket before the end means the string is a
			// sequence of lists rather than one list.
			if depth == 0 && i != len(s)-1 {
				return s, false
			}
		}
	}
	if depth != 0 {
		return s, false
	}
	return s[1 : len(s)-1], true
}

// splitList splits s on mapListSep at the top level, ignoring a separator nested inside brackets
// or inside a quoted element.
func splitList(s string) []string {
	var (
		parts   []string
		cur     strings.Builder
		depth   int
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
		case quoted:
		case r == listOpen:
			depth++
		case r == listClose:
			depth--
		case depth == 0 && string(r) == mapListSep:
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteRune(r)
	}
	return append(parts, cur.String())
}

// quoteListElem wraps an element that would otherwise read back as something else, since this text
// is re-parsed: viper hands a flag's own string to the decoder rather than the elements behind it.
func quoteListElem(s string) string {
	if !strings.ContainsAny(s, mapListSep+`"`+string(listOpen)+string(listClose)) && s == strings.TrimSpace(s) {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func unquoteListElem(s string) string {
	if !isQuoted(s) {
		return s
	}
	return strings.ReplaceAll(s[1:len(s)-1], `\"`, `"`)
}
