package cli

import (
	"cmp"
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
		case child.leaf.Kind() == reflect.Map:
			d.Elem().Set(mergeMaps(child.leaf, d.Elem(), s.Elem()))
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

var stringType = reflect.TypeFor[string]()

// Config file map keys are text, so a number or bool key is read as a string and parsed as its flag would be
// (see fromFile).
func fileValueType(t reflect.Type, lang markup.Markup) reflect.Type {
	switch t.Kind() {
	case reflect.Map:
		key := t.Key()
		if key.Kind() != reflect.String && !lang.IsLeaf(key) {
			key = stringType
		}

		return reflect.MapOf(key, fileValueType(t.Elem(), lang))
	case reflect.Slice:
		return reflect.SliceOf(fileValueType(t.Elem(), lang))
	case reflect.Pointer:
		return reflect.PointerTo(fileValueType(t.Elem(), lang))
	default:
		return t
	}
}

func fromFile(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	if v.Type() == t {
		return v, nil
	}

	switch t.Kind() {
	case reflect.Map:
		out := reflect.MakeMapWithSize(t, v.Len())
		seen := map[any]string{}
		// Sorted, so a collision is reported the same way every run.
		byText := func(a, b reflect.Value) int { return cmp.Compare(a.String(), b.String()) }
		for _, rawKey := range slices.SortedFunc(slices.Values(v.MapKeys()), byText) {
			key := rawKey
			if key.Type() != t.Key() {
				var err error
				if key, err = mapKey(t.Key(), key.String(), seen); err != nil {
					return reflect.Value{}, err
				}
			}

			elem, err := fromFile(t.Elem(), v.MapIndex(rawKey))
			if err != nil {
				return reflect.Value{}, err
			}

			out.SetMapIndex(key, elem)
		}

		return out, nil
	case reflect.Slice:
		out := reflect.MakeSlice(t, v.Len(), v.Len())
		for i := range v.Len() {
			elem, err := fromFile(t.Elem(), v.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}

			out.Index(i).Set(elem)
		}

		return out, nil
	default: // reflect.Pointer, the only other kind fileValueType changes
		if v.IsNil() {
			return reflect.Zero(t), nil
		}

		elem, err := fromFile(t.Elem(), v.Elem())
		if err != nil {
			return reflect.Value{}, err
		}

		out := reflect.New(t.Elem())
		out.Elem().Set(elem)
		return out, nil
	}
}

// seen makes two texts that parse to one key, such as 16 and 0x10, an error rather than one silently replacing the
// other.
func mapKey(t reflect.Type, text string, seen map[any]string) (reflect.Value, error) {
	key, err := parseText(t, text)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("key %q: %w", text, err)
	}

	if other, dup := seen[key.Interface()]; dup && other != text {
		return reflect.Value{}, fmt.Errorf("keys %q and %q are both %v", other, text, key.Interface())
	}

	seen[key.Interface()] = text
	return key, nil
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

		if f.Kind() != reflect.Map {
			f.Set(vals[0])
		} else {
			// The struct's own entries first, then sources lowest precedence first, so the highest wins a key.
			slices.Reverse(vals)
			f.Set(mergeMaps(f.Type(), append([]reflect.Value{f}, vals...)...))
		}
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
