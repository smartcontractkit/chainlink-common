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
func (n *keyNode) overlay(dst, src reflect.Value, lang markup.Markup) {
	for _, child := range n.children {
		s, d := src.Field(child.index), dst.Field(child.index)
		switch {
		case s.IsNil():
		case d.IsNil():
			d.Set(s)
		case child.leaf == nil:
			child.overlay(d.Elem(), s.Elem(), lang)
		case mergesByKey(child.leaf, lang):
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

// Config file map keys are text, so every key is read as a string and parsed as its flag would be (see fromFile), even
// one the markup could read itself, so two texts for one key are always caught.
//
// A recursive type that needs converting, such as type M map[uint32]*M, is an error: the inner M can't be converted.
// visiting holds the types being converted, false until one is reached again inside itself.
func fileValueType(t reflect.Type, lang markup.Markup, visiting map[reflect.Type]bool) (reflect.Type, error) {
	if _, ok := visiting[t]; ok {
		visiting[t] = true
		// Recursion itself is not an error, for example, nothing should stop the type below from being used.
		// type Endpoint struct {
		// 	 URL      string
		// 	 Fallback *Endpoint
		// }
		// We only prevent recursive definitions if we need to modify the type
		// type Endpoint struct {
		// 	 URLs      map[int32]string
		// 	 Fallback *Endpoint
		// }
		// because Go's reflection can't redefine it for the string parsing of the key as
		// <runtime type> struct {
		// 	 URLs      map[string]string
		// 	 Fallback *<runtime type>
		// }
		return t, nil
	}

	if lang.IsLeaf(t) {
		return t, nil
	}

	visiting[t] = false
	defer delete(visiting, t)
	out, err := convertFileValueType(t, lang, visiting)
	if err != nil {
		return nil, err
	}

	if visiting[t] && out != t {
		return nil, fmt.Errorf("%s holds itself and a map with non-string keys, which a config file can't set; "+
			"key the map by string", t)
	}

	return out, nil
}

func convertFileValueType(t reflect.Type, lang markup.Markup, visiting map[reflect.Type]bool) (reflect.Type, error) {
	var wrap func(reflect.Type) reflect.Type
	switch t.Kind() {
	case reflect.Map:
		if holdsIdentityKind(t.Key()) {
			return nil, fmt.Errorf("%s has keys that hold a pointer, channel or interface, which can compare by address, "+
				"so 2 and 2 could be different keys; key it by value", t)
		}

		wrap = func(elem reflect.Type) reflect.Type { return reflect.MapOf(stringType, elem) }
	case reflect.Slice:
		wrap = reflect.SliceOf
	case reflect.Array:
		wrap = func(elem reflect.Type) reflect.Type { return reflect.ArrayOf(t.Len(), elem) }
	case reflect.Pointer:
		wrap = reflect.PointerTo
	case reflect.Struct:
		return fileStructType(t, lang, visiting)
	default:
		return t, nil
	}

	elem, err := fileValueType(t.Elem(), lang, visiting)
	if err != nil {
		return nil, err
	}

	// !readsText ensures that a type backed by a string is still given a chance to detect duplicate parsing
	// to the same value. For example if type foo string has an UnmarshalText that lower cases input, we need to ensure
	// 'foo' and 'Foo' are considered collisions instead of silently replacing them.
	if elem == t.Elem() && (t.Kind() != reflect.Map || t.Key().Kind() == reflect.String && !readsText(t.Key())) {
		return t, nil
	}

	return wrap(elem), nil
}

// Unexported fields, which reflect.StructOf can't make, and fields the markup ignores are never set from a file, so
// they're left out. StructOf doesn't promote an embedded field's methods either, so a struct that embeds one can't be
// rebuilt.
func fileStructType(t reflect.Type, lang markup.Markup, visiting map[reflect.Type]bool) (reflect.Type, error) {
	var fields []reflect.StructField
	changed, embeds := false, false
	for f := range t.Fields() {
		if _, read := lang.Key(f); !read || (!f.IsExported() && !f.Anonymous) {
			continue
		}

		ft, err := fileValueType(f.Type, lang, visiting)
		if err != nil {
			return nil, err
		}

		changed = changed || ft != f.Type
		embeds = embeds || f.Anonymous
		fields = append(fields, reflect.StructField{Name: f.Name, Type: ft, Tag: f.Tag})
	}

	switch {
	case !changed:
		return t, nil
	case embeds:
		return nil, fmt.Errorf("%s holds a map with non-string keys and embeds a field, so it can't be rebuilt to read "+
			"the keys as text; name the embedded field", t)
	default:
		return reflect.StructOf(fields), nil
	}
}

// Turns a value decoded as fileValueType(t) back into t, parsing the map keys that were read as strings.
func fromFile(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	if v.Type() == t {
		return v, nil
	}

	switch t.Kind() {
	case reflect.Map:
		out := reflect.MakeMapWithSize(t, v.Len())
		seen := map[any]string{}
		// Sorted, so a collision is reported the same way every run.
		rawKeys := v.MapKeys()
		slices.SortFunc(rawKeys, func(a, b reflect.Value) int { return cmp.Compare(a.String(), b.String()) })
		for _, rawKey := range rawKeys {
			key, err := mapKey(t.Key(), rawKey.String(), seen)
			if err != nil {
				return reflect.Value{}, err
			}

			elem, err := fromFile(t.Elem(), v.MapIndex(rawKey))
			if err != nil {
				return reflect.Value{}, err
			}

			out.SetMapIndex(key, elem)
		}

		return out, nil
	case reflect.Slice, reflect.Array:
		out := reflect.New(t).Elem()
		if t.Kind() == reflect.Slice {
			out = reflect.MakeSlice(t, v.Len(), v.Len())
		}

		for i := range v.Len() {
			elem, err := fromFile(t.Elem(), v.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}

			out.Index(i).Set(elem)
		}

		return out, nil
	case reflect.Struct:
		out := reflect.New(t).Elem()
		for i := range t.NumField() {
			src := v.FieldByName(t.Field(i).Name)
			if !src.IsValid() {
				continue
			}

			field, err := fromFile(t.Field(i).Type, src)
			if err != nil {
				return reflect.Value{}, err
			}

			out.Field(i).Set(field)
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

// Pointers and channels compare by identity, so two parses of one text are different keys. An interface may hold
// either, or a value that can't be a key at all. Arrays and structs are the only composite key types, and neither can
// hold itself except through a pointer.
func holdsIdentityKind(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.UnsafePointer, reflect.Chan, reflect.Interface:
		return true
	case reflect.Array:
		return holdsIdentityKind(t.Elem())
	case reflect.Struct:
		for field := range t.Fields() {
			if holdsIdentityKind(field.Type) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

// seen makes two texts that parse to one key, such as 16 and 0x10, an error rather than one silently replacing the
// other.
func mapKey(keyType reflect.Type, keyText string, seen map[any]string) (reflect.Value, error) {
	key, err := parseText(keyType, keyText)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("key %q: %w", keyText, err)
	}

	// NaN never equals itself, nor does a key holding one, so each would be its own key and no lookup could find one.
	if !key.Equal(key) {
		return reflect.Value{}, fmt.Errorf("key %q never equals itself, as NaN doesn't, so it can't be looked up", keyText)
	}

	if other, dup := seen[key.Interface()]; dup && other != keyText {
		return reflect.Value{}, fmt.Errorf("keys %q and %q are both %v", other, keyText, key.Interface())
	}

	seen[key.Interface()] = keyText
	return key, nil
}
