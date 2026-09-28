package cli

import (
	"cmp"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

func envVarName(prefix, key string) string {
	suffix := strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(key))
	return strings.TrimSuffix(strings.ToUpper(prefix), "_") + "_" + suffix
}

// keyNode is one level of the keys a config file can hold, across every registered struct.
type keyNode struct {
	// leaf is the type a file holds a leaf as (see fileValueType); nil for a table.
	leaf     reflect.Type
	children map[string]*keyNode
	index    int
}

// add records path, holding a leaf of type leaf. It rejects a key that is a value in one struct and
// a table in another, and keys the markup can't tell apart, since one config file key would set
// both. Several commands may register the same key.
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
		case (child.leaf == nil) == isLeaf:
			return fmt.Errorf("%s is a value in one struct and a table in another", at)
		case isLeaf && child.leaf != leaf:
			return fmt.Errorf("%s is a %s in one struct and a %s in another", at, child.leaf, leaf)
		}
		node = child
	}
	return nil
}

// fileType is the struct a config file decodes into: a pointer per key, so nil is a key the file
// left out, named by a tag so the markup matches keys to fields exactly as it would the originals.
func (n *keyNode) fileType(lang markup.Markup) reflect.Type {
	names := slices.Sorted(maps.Keys(n.children))
	fields := make([]reflect.StructField, len(names))
	for i, name := range names {
		child := n.children[name]
		child.index = i
		t := child.leaf
		if t == nil {
			t = child.fileType(lang)
		}
		fields[i] = reflect.StructField{
			Name: fmt.Sprintf("F%d", i),
			Type: reflect.PointerTo(t),
			Tag:  lang.RenameTag(name),
		}
	}
	return reflect.StructOf(fields)
}

// overlay copies src's keys onto dst, both of n's fileType: tables and maps merge by key, and any
// other leaf src holds, a list included, replaces dst's.
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

// lookup is the value at path in v, of n's fileType, if a file held it.
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

// fileValueType is t as a config file holds it. A map keyed by a number or a bool is keyed by
// text there, which fromFile parses as the key's flag would; a string key, or one the markup reads
// whole, is the key as it is.
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

// fromFile converts v, of fileValueType(t), to t.
func fromFile(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	if v.Type() == t {
		return v, nil
	}
	switch t.Kind() {
	case reflect.Map:
		out := reflect.MakeMapWithSize(t, v.Len())
		seen := map[any]string{}
		// Sorted, so a collision is reported the same way every run.
		for _, fileKey := range slices.SortedFunc(slices.Values(v.MapKeys()), func(a, b reflect.Value) int { return cmp.Compare(a.String(), b.String()) }) {
			key := fileKey
			if key.Type() != t.Key() {
				var err error
				if key, err = mapKey(t.Key(), key.String(), seen); err != nil {
					return reflect.Value{}, err
				}
			}
			elem, err := fromFile(t.Elem(), v.MapIndex(fileKey))
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

// mapKey parses text as a map key of type t, as a flag of t would. seen holds the keys parsed so
// far, and the text of each, so two spellings of one key, such as 16 and 0x10, are an error rather
// than one silently replacing the other.
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

// decodeEntry sets the leaves a source supplied, allocating only their pointer sections, so unset
// fields keep the struct's defaults and untouched optional *struct sections stay nil for
// `required_without`/`excluded_with`.
func decodeEntry(entry *targetEntry) error {
	entry.suppliedFields = map[uintptr]bool{}
	dst := reflect.Indirect(reflect.ValueOf(entry.target))
	for _, k := range entry.keys {
		vals, err := entry.sources(k)
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
			// Over the struct's own entries, the lowest source first so the highest wins a key.
			slices.Reverse(vals)
			f.Set(mergeMaps(f.Type(), append([]reflect.Value{f}, vals...)...))
		}
		// By address, for the `set` validator, which sees only values.
		entry.suppliedFields[f.Addr().Pointer()] = true
	}
	return nil
}

// IsBuiltinCommand reports whether cmd is, or is under, cobra's help or completion commands.
// Decoding skips these; custom PreRunE checks should too, so help works without config.
func IsBuiltinCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return true
		}
	}
	return false
}

// allocateField walks path, allocating nil pointers; invalid where it can't.
// Called only before a write, so untouched optional sections stay nil.
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
