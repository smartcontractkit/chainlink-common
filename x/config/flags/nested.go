package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// A list of structs has no single-element text form, so it is not carried on the command line the
// way a list of strings is. It is carried column-wise instead: one flag per leaf field of the
// element struct, each holding a list with one entry per element.
//
//	type Node struct{ Name string; URL string }
//	type Cfg  struct{ Nodes []Node }
//
//	--nodes.name a,b --nodes.url http://a,http://b   =>  [{a http://a} {b http://b}]
//
// The lists line up by position, so the i'th entry of every column belongs to the i'th struct, and
// a column that runs short simply leaves that field at its zero value. Nesting adds a list level
// rather than a new kind of key, so a list inside a list of structs is a list of lists:
//
//	--evm.nodes.name '[a,b],[c]'                     =>  chain 0 has nodes a and b, chain 1 c
//
// A config file still writes the same field row-wise, as an array of tables ([[nodes]]), which is
// what a config file is good at; the two are merged by pivotColumns, with the columns (a flag or
// an env var) overriding the row they land on.

// columns is one list-of-structs field's value as the sources delivered it: whatever row-wise
// value the config file held, plus one entry per leaf key bound underneath it.
type columns struct {
	// rows is the config file's array-of-tables value for the field, if it had one.
	rows any
	// cols holds the columnar values, nested by key path under the list - for
	// "shards.nodes.name" bound under "sharded-dons", cols is {"shards": {"nodes": {"name": ...}}}.
	// A value is a list one level deeper than the field it describes, one entry per struct at each
	// list level it passes through.
	cols map[string]any
}

func newColumns() *columns { return &columns{cols: map[string]any{}} }

// isStructList reports whether t is a list whose elements are structs bound field by field - the
// shape this file's columnar form describes. A list that unmarshals itself from text, or whose
// elements do, is text and is bound as a list value instead.
func isStructList(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if k := t.Kind(); k != reflect.Slice && k != reflect.Array {
		return false
	}
	if implementsTextUnmarshaler(t) {
		return false
	}
	el := t.Elem()
	for el.Kind() == reflect.Pointer {
		el = el.Elem()
	}
	return el.Kind() == reflect.Struct && !implementsTextUnmarshaler(el)
}

// columnsToRowsHookFunc turns the columnar form back into the rows a list of structs decodes from,
// so the decoder itself never has to know the flags were columns. It fires again for every list
// level, since indexing a level's columns leaves the level below still columnar.
func columnsToRowsHookFunc() mapstructure.DecodeHookFuncType {
	return func(from, to reflect.Type, data any) (any, error) {
		if !isStructList(to) {
			return data, nil
		}
		switch v := data.(type) {
		case *columns:
			return pivotColumns(v.rows, v.cols)
		case map[string]any:
			return pivotColumns(nil, v)
		default:
			return data, nil
		}
	}
}

// pivotColumns merges a row-wise value (from a config file) with columnar ones (from flags and env
// vars) into the rows a list of structs decodes from. The result has as many rows as the longest
// source; a column that is shorter leaves the rows past its end alone.
func pivotColumns(rowsVal any, cols map[string]any) (any, error) {
	base, err := rowsOf(rowsVal)
	if err != nil {
		return nil, err
	}

	n := len(base)
	if c := columnsLen(cols); c > n {
		n = c
	}

	out := make([]any, n)
	for i := range out {
		var row map[string]any
		if i < len(base) {
			row = base[i]
		}
		out[i] = mergeKeys(row, indexColumns(cols, i))
	}
	return out, nil
}

// rowsOf normalizes a config file's array-of-tables value into one map per row.
func rowsOf(v any) ([]map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(v)
	if k := rv.Kind(); k != reflect.Slice && k != reflect.Array {
		return nil, fmt.Errorf("expected a list of tables, got %T - set the individual fields (key.field=v1,v2) instead", v)
	}
	out := make([]map[string]any, rv.Len())
	for i := range out {
		el := rv.Index(i).Interface()
		m, ok := el.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("list element %d: expected a table, got %T", i, el)
		}
		out[i] = m
	}
	return out, nil
}

// columnsLen is the number of rows the columns describe: the longest column at this level,
// counting through the nested levels of a list of structs inside a list of structs.
func columnsLen(cols map[string]any) int {
	n := 0
	for _, v := range cols {
		l := 0
		if sub, ok := v.(map[string]any); ok {
			l = columnsLen(sub)
		} else if list, ok := asList(v); ok {
			l = len(list)
		}
		if l > n {
			n = l
		}
	}
	return n
}

// indexColumns takes the i'th row's slice of every column, stripping one list level: what is left
// describes row i alone, and is itself columnar wherever the row holds a further list of structs.
// A column with no i'th entry contributes no key at all, so it doesn't blank out a value the
// config file supplied for that row.
func indexColumns(cols map[string]any, i int) map[string]any {
	out := make(map[string]any, len(cols))
	for k, v := range cols {
		if sub, ok := v.(map[string]any); ok {
			out[k] = indexColumns(sub, i)
			continue
		}
		list, ok := asList(v)
		if !ok || i >= len(list) {
			continue
		}
		out[k] = list[i]
	}
	return out
}

// asList reads one list level off a column's value, whichever form its source delivered: already a
// list, from a config file, or the bracketed text a flag and an env var carry.
func asList(v any) ([]any, bool) {
	switch t := v.(type) {
	case nil:
		return nil, false
	case []any:
		return t, true
	case string:
		inner, _ := trimListBrackets(strings.TrimSpace(t))
		if strings.TrimSpace(inner) == "" {
			return nil, true
		}
		parts := splitList(inner)
		out := make([]any, len(parts))
		for i, p := range parts {
			out[i] = strings.TrimSpace(p)
		}
		return out, true
	}

	rv := reflect.ValueOf(v)
	if k := rv.Kind(); k != reflect.Slice && k != reflect.Array {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

// mergeKeys overlays one row's columnar values on the row the config file supplied, recursing into
// nested tables so a flag can override a single field of one without discarding the rest.
//
// Keys are matched the way the decoder matches them to fields (see matchKeyToFieldName), since
// viper lowercases what it reads from a config file while the columns are keyed by the tag: left
// unmatched, "chainid" and "chain-id" would both reach the decoder and one would silently win over
// the other.
func mergeKeys(base, overlay map[string]any) map[string]any {
	if len(overlay) == 0 {
		return base
	}
	out := make(map[string]any, len(base)+len(overlay))
	byName := make(map[string]string, len(base))
	for k, v := range base {
		out[k] = v
		byName[stripSeparators(strings.ToLower(k))] = k
	}

	for k, v := range overlay {
		norm := stripSeparators(strings.ToLower(k))
		prev, matched := byName[norm]
		if matched {
			delete(out, prev)
		}
		sub, isSub := v.(map[string]any)
		if !isSub {
			out[k] = v
			continue
		}
		var prevVal any
		if matched {
			prevVal = base[prev]
		}
		switch pv := prevVal.(type) {
		case map[string]any:
			out[k] = mergeKeys(pv, sub)
		case []any:
			// The config file wrote this field as an array of tables while a flag addressed it
			// column-wise; hand both to the hook, which merges them the same way this row was.
			out[k] = &columns{rows: pv, cols: sub}
		default:
			out[k] = mergeKeys(nil, sub)
		}
	}
	return out
}

// collectColumn reads a leaf field's compiled-in defaults out of the list of structs it lives in,
// column-wise: the field's value within every element of listRoot, one list level deeper for every
// further list of structs on the way down to it. leafDepth is how many list levels the field's own
// type has, so a []string leaf inside a []Node yields [[a,b],[c]] rather than four loose strings.
func collectColumn(listRoot reflect.Value, goPath []string, leafDepth int) any {
	for listRoot.Kind() == reflect.Pointer {
		if listRoot.IsNil() {
			return emptyTree(leafDepth)
		}
		listRoot = listRoot.Elem()
	}
	if !listRoot.IsValid() {
		return emptyTree(leafDepth)
	}

	if len(goPath) == 0 {
		return treeOf(listRoot, leafDepth)
	}

	if k := listRoot.Kind(); k == reflect.Slice || k == reflect.Array {
		out := make([]any, listRoot.Len())
		for i := range out {
			out[i] = collectColumn(listRoot.Index(i), goPath, leafDepth)
		}
		return out
	}

	if listRoot.Kind() != reflect.Struct {
		return emptyTree(leafDepth)
	}
	return collectColumn(listRoot.FieldByName(goPath[0]), goPath[1:], leafDepth)
}

// emptyTree is the value of a column that runs out before it reaches its field - a nil pointer on
// the way down, say - at the depth that field's own type has.
func emptyTree(depth int) any {
	if depth == 0 {
		return ""
	}
	return []any{}
}
