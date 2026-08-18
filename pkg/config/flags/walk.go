package flags

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

// fieldMeta describes one field discovered while walking a config struct. Shared by the
// CLI/env/viper binding logic (bindLeafFlag) and the doc-generation logic (docs.go).
type fieldMeta struct {
	// goPath is the chain of Go field names from the root struct to this field.
	goPath []string
	// keyPath is the chain of toml/mapstructure keys from the root struct to this field.
	// Squashed and anonymous fields contribute no key segment of their own.
	keyPath []string
	// field is the raw struct field (field.Type may be a pointer).
	field reflect.StructField
	// parent is the struct type declaring field, used to resolve sibling field names (e.g.
	// the targets of cross-field `validate` rules) back to their config keys.
	parent reflect.Type
	// elem is field's value with one level of pointer dereferenced (or the zero value of
	// elemType, if the pointer is nil).
	elem reflect.Value
	// elemType is field.Type with one level of pointer stripped.
	elemType reflect.Type
	// isTextUnmarshaler is true if elemType (or a pointer to it) implements
	// encoding.TextUnmarshaler, meaning it should be treated as an opaque leaf value even
	// though its Kind may be Struct.
	isTextUnmarshaler bool
}

// key joins keyPath into a dotted key, e.g. "chain.id".
func (f fieldMeta) key() string { return strings.Join(f.keyPath, ".") }

// goName joins goPath into a dotted Go field path, e.g. "Chain.ID".
func (f fieldMeta) goName() string { return strings.Join(f.goPath, ".") }

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// tagKey returns field's config key and whether it is squashed (contributing its own fields to
// the parent rather than a nested table), per the configured tag name and squash option - e.g.
// `toml:"key,inline"` with Options.DecoderConfig.TagName "toml" and SquashTagOption "inline".
// Falls back to the Go field name in kebab-case when the field carries no usable tag, so a
// struct only needs a tag where its key differs from its field name.
func (o Options) tagKey(field reflect.StructField) (key string, squash bool) {
	tag := field.Tag.Get(o.tagName())
	if tag == "" || tag == "-" {
		tag = strcase.ToKebab(field.Name)
	}

	parts := strings.Split(tag, ",")
	key = parts[0]
	if key == "" {
		key = strcase.ToKebab(field.Name)
	}
	for _, p := range parts[1:] {
		if p == o.squashOption() {
			squash = true
		}
	}

	// An embedded struct is squashed without any tag, but only when the decoder is configured
	// to do so - mirroring mapstructure's own `d.config.Squash && v.Kind() == reflect.Struct &&
	// f.Anonymous`. Squashing here when the decoder won't (or vice versa) would name the flag
	// and the decoded key differently, and the value would silently never arrive.
	if o.DecoderConfig.Squash && field.Anonymous && field.Type.Kind() == reflect.Struct {
		squash = true
	}

	return key, squash
}

// implementsTextUnmarshaler reports whether t (or *t) implements encoding.TextUnmarshaler.
func implementsTextUnmarshaler(t reflect.Type) bool {
	return t.Implements(textUnmarshalerType) || reflect.PointerTo(t).Implements(textUnmarshalerType)
}

// isTextValueType reports whether a value of type t can be carried as a single piece of text -
// a primitive, or a type that unmarshals itself from text (time.Duration, config.Duration, a
// net.IP, ...). These are exactly the types that survive the round trip through a flag string
// or an env var, so they're what a map's keys and values must be for the map to bind as one.
func isTextValueType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if implementsTextUnmarshaler(t) {
		return true
	}
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// isMapValueType reports whether a map's values can be written on the command line: a single
// piece of text, or a list of them ("primary=a,b"). A []byte is excluded despite being a list
// of a primitive - it is text, not a list of numbers, and the two decode differently.
func isMapValueType(t reflect.Type) bool {
	if isTextValueType(t) {
		return true
	}
	if k := t.Kind(); k != reflect.Slice && k != reflect.Array {
		return false
	}
	return t.Elem().Kind() != reflect.Uint8 && isTextValueType(t.Elem())
}

// structVisitor holds the callbacks invoked while walking a struct.
type structVisitor struct {
	// leaf is called for every leaf field (scalar, or struct implementing
	// encoding.TextUnmarshaler).
	leaf func(fieldMeta) error
	// branch is optionally called for every non-leaf (nested struct) field before it is
	// recursed into. Returning skip=true prevents descending into it (no further leaf/branch
	// calls for anything under it).
	branch func(fieldMeta) (skip bool, err error)
}

// walkStruct recursively visits every field of target (a struct or pointer to a struct) in
// declaration order, calling visitor.leaf for scalar fields and visitor.branch (if set) for
// nested struct fields. Keys are derived using opts, so the walk agrees with how opts'
// decoder will later map the same fields.
func walkStruct(target any, opts Options, visitor structVisitor) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("target pointer cannot be nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct or pointer to struct")
	}
	return walkStructValue(v, v.Type(), nil, nil, opts, visitor)
}

func walkStructValue(v reflect.Value, t reflect.Type, goPath, keyPath []string, opts Options, visitor structVisitor) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // unexported
		}

		key, squash := opts.tagKey(field)

		fieldGoPath := append(append([]string{}, goPath...), field.Name)
		fieldKeyPath := keyPath
		if !squash {
			fieldKeyPath = append(append([]string{}, keyPath...), key)
		}

		fieldVal := v.Field(i)
		elemType := field.Type
		elemVal := fieldVal
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
			if fieldVal.IsNil() {
				elemVal = reflect.Zero(elemType)
			} else {
				elemVal = fieldVal.Elem()
			}
		}

		isTextUnmarshaler := implementsTextUnmarshaler(field.Type) || implementsTextUnmarshaler(elemType)

		meta := fieldMeta{
			goPath:            fieldGoPath,
			keyPath:           fieldKeyPath,
			field:             field,
			parent:            t,
			elem:              elemVal,
			elemType:          elemType,
			isTextUnmarshaler: isTextUnmarshaler,
		}

		if elemType.Kind() == reflect.Struct && !isTextUnmarshaler {
			skip := false
			if visitor.branch != nil {
				var err error
				skip, err = visitor.branch(meta)
				if err != nil {
					return err
				}
			}
			if skip {
				continue
			}
			if err := walkStructValue(elemVal, elemType, fieldGoPath, fieldKeyPath, opts, visitor); err != nil {
				return err
			}
			continue
		}

		if visitor.leaf != nil {
			if err := visitor.leaf(meta); err != nil {
				return err
			}
		}
	}
	return nil
}
