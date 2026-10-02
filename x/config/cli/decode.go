package cli

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

// keyNode is the key tree Binder.commandConfig builds; see there for an example.
type keyNode struct {
	leaf     reflect.Type
	children map[string]*keyNode
	index    int
}

// A key the markup can't tell from another, such as Value and VALUE when it matches case-insensitively, would set both
// fields, so it is rejected too.
func (n *keyNode) add(path []string, leaf reflect.Type, lang markup.Markup) error {
	node := n
	for i, segment := range path {
		at := strings.Join(path[:i+1], ".")
		for name := range node.children {
			if name != segment && lang.MatchesKey(segment, name) {
				return fmt.Errorf("%s and %s differ only in case, so a config file key would set both",
					strings.Join(append(slices.Clone(path[:i]), name), "."), at)
			}
		}

		isLeaf := i == len(path)-1
		child, exists := node.children[segment]
		switch {
		case !exists:
			child = &keyNode{children: map[string]*keyNode{}}
			if isLeaf {
				child.leaf = leaf
			}

			node.children[segment] = child
		case isLeaf || child.leaf != nil:
			return fmt.Errorf("two fields hold %s, so a config file key would set both", at)
		}

		node = child
	}

	return nil
}

// Pointers tell a key the file left out from one set to the zero value. Tags, not names, match keys, so the F<n> names
// only show in the markup's errors.
func (n *keyNode) fileValuesType(lang markup.Markup) reflect.Type {
	names := slices.Sorted(maps.Keys(n.children))
	fields := make([]reflect.StructField, len(names))
	for i, name := range names {
		child := n.children[name]
		child.index = i
		t := child.leaf
		if t == nil {
			t = child.fileValuesType(lang)
		}

		fields[i] = reflect.StructField{
			Name: fmt.Sprintf("F%d", i),
			Type: reflect.PointerTo(t),
			Tag:  lang.RenameTag(name),
		}
	}

	return reflect.StructOf(fields)
}

// A list replaces rather than merges, or a later file could never shorten it.
func (n *keyNode) overlay(dst, src reflect.Value) {
	for _, child := range n.children {
		s, d := src.Field(child.index), dst.Field(child.index)
		switch {
		case s.IsNil():
		case d.IsNil():
			d.Set(s)
		case child.leaf == nil:
			child.overlay(d.Elem(), s.Elem())
		default:
			d.Set(s)
		}
	}
}

func (n *keyNode) lookup(v reflect.Value, path []string) (reflect.Value, bool) {
	node := n
	for _, segment := range path {
		node = node.children[segment]
		if v = v.Field(node.index); v.IsNil() {
			return reflect.Value{}, false
		}

		v = v.Elem()
	}

	return v, true
}

// Only pointer sections on the way to a set field are allocated, so an untouched optional *struct stays nil for
// `required_without` and `excluded_with`.
func decodeEntry(entry *targetEntry, cc commandConfig) error {
	dst := reflect.Indirect(reflect.ValueOf(entry.target))
	for _, k := range entry.keys {
		vals, err := entry.sources(k, cc)
		if err != nil {
			return fmt.Errorf("%s: %w", k.key, err)
		}

		if len(vals) == 0 {
			continue
		}

		f := allocateField(dst, k.goPath)
		for f.Kind() == reflect.Pointer {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		f.Set(vals[0])
	}

	return nil
}

// Call only to write, so sections are allocated only when something in them is set.
func allocateField(v reflect.Value, path []string) reflect.Value {
	for _, p := range path {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				if !v.CanSet() {
					return reflect.Value{}
				}

				v.Set(reflect.New(v.Type().Elem()))
			}

			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return reflect.Value{}
		}

		if v = v.FieldByName(p); !v.IsValid() {
			return reflect.Value{}
		}
	}

	return v
}
