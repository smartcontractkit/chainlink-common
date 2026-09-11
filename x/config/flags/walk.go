package flags

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

// fieldMeta describes one field discovered while walking a config struct.
type fieldMeta struct {
	// keyPath is the chain of toml/mapstructure keys from the root struct to this field.
	// Squashed and anonymous fields contribute no key segment of their own.
	keyPath []string
	// field is the raw struct field (field.Type may be a pointer).
	field reflect.StructField
	// elem is field's value with one level of pointer dereferenced (or the zero value of
	// elemType, if the pointer is nil).
	elem reflect.Value
	// elemType is field.Type with one level of pointer stripped.
	elemType reflect.Type
	// isTextUnmarshaler marks a field that is an opaque leaf value even though its Kind may be
	// Struct, because it (or a pointer to it) implements encoding.TextUnmarshaler.
	isTextUnmarshaler bool
	// isStructList marks the field that opens a list of structs - the branch whose element
	// struct the walk descends into. See nested.go.
	isStructList bool
	// listDepth is how many lists of structs this field sits inside: 0 for an ordinary field, 1
	// for a field of a []Node's element, 2 for a field of a []Node's element reached through a
	// []Chain. It is the number of list levels the field's flag carries.
	listDepth int
	// listRoot is the outermost list-of-structs value this field lives inside (valid only when
	// listDepth > 0) and listPath the chain of Go field names from that list's element struct
	// down to the field. Together they recover the field's compiled-in defaults column-wise,
	// through however many further lists lie between - see collectColumn.
	listRoot reflect.Value
	listPath []string
	// listKeyPath is the config key path of that outermost list, which is the key the config
	// file's array of tables arrives under.
	listKeyPath []string
}

// key joins keyPath into a dotted key, e.g. "chain.id".
func (f fieldMeta) key() string { return strings.Join(f.keyPath, ".") }

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// tagKey returns field's config key and whether it is squashed (contributing its own fields to
// the parent rather than a nested table), per the configured tag name and squash option. It falls
// back to the Go field name in kebab-case when the field carries no usable tag, so a struct only
// needs a tag where its key differs from its field name.
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
	// Only a struct can be squashed - mapstructure rejects the option on anything else, and a
	// TOML encoder reads the same `,inline` as "write this value on one line", which several
	// config structs carry on plain scalars. Honouring it there would drop those fields' key
	// segments and collapse every one of them onto their parent's name.
	if isSquashableType(field.Type) {
		for _, p := range parts[1:] {
			if p == o.squashOption() {
				squash = true
			}
		}
	}

	// An embedded struct is squashed without any tag, but only when the decoder is configured to
	// do so - mirroring mapstructure's own `d.config.Squash && v.Kind() == reflect.Struct &&
	// f.Anonymous`. Squashing here when the decoder won't (or vice versa) would name the flag and
	// the decoded key differently, and the value would silently never arrive.
	if o.DecoderConfig.Squash && field.Anonymous && field.Type.Kind() == reflect.Struct {
		squash = true
	}

	return key, squash
}

func isSquashableType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

func implementsTextUnmarshaler(t reflect.Type) bool {
	return t.Implements(textUnmarshalerType) || reflect.PointerTo(t).Implements(textUnmarshalerType)
}

// isTextValueType reports whether a value of type t can be carried as a single piece of text - a
// primitive, or a type that unmarshals itself from text (time.Duration, config.Duration, a
// net.IP, ...). These are exactly the types that survive the round trip through a flag string or
// an env var, so they're what a map's keys and values must be for the map to bind as one.
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
// piece of text, or a list of them ("primary=a,b"). A []byte is excluded despite being a list of
// a primitive - it is text, not a list of numbers, and the two decode differently.
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
	// branch is optionally called for every non-leaf (nested struct) field before it is recursed
	// into. Returning skip=true prevents descending into it.
	branch func(fieldMeta) (skip bool, err error)
}

// walkStruct recursively visits every field of target (a struct or pointer to a struct) in
// declaration order. Keys are derived using opts, so the walk agrees with how opts' decoder will
// later map the same fields.
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
	return walkStructValue(v, v.Type(), walkScope{ancestors: []reflect.Type{v.Type()}}, opts, visitor)
}

// walkScope is where the walk currently is: the paths accumulated to here, plus the list-of-
// structs context those paths run through (see fieldMeta's listDepth/listRoot/listPath).
type walkScope struct {
	keyPath     []string
	listDepth   int
	listRoot    reflect.Value
	listPath    []string
	listKeyPath []string
	// ancestors is the chain of struct types the walk is currently inside. A struct that contains
	// itself - directly, through a pointer, or through a list of itself - describes an unbounded
	// set of keys, so the recursion stops the second time a type appears rather than running out
	// of stack.
	ancestors []reflect.Type
}

func (s walkScope) descendsInto(t reflect.Type) bool {
	for _, a := range s.ancestors {
		if a == t {
			return true
		}
	}
	return false
}

func walkStructValue(v reflect.Value, t reflect.Type, scope walkScope, opts Options, visitor structVisitor) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		key, squash := opts.tagKey(field)

		fieldKeyPath := scope.keyPath
		if !squash {
			fieldKeyPath = append(append([]string{}, scope.keyPath...), key)
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
			keyPath:           fieldKeyPath,
			field:             field,
			elem:              elemVal,
			elemType:          elemType,
			isTextUnmarshaler: isTextUnmarshaler,
			listDepth:         scope.listDepth,
			listRoot:          scope.listRoot,
			listPath:          append(append([]string{}, scope.listPath...), field.Name),
			listKeyPath:       scope.listKeyPath,
		}

		switch {
		// A list of structs is a branch like a nested struct, one list level deeper: its element
		// struct's fields are walked, and each of their flags carries a list with one entry per
		// element. See nested.go.
		case isStructList(elemType) && !isTextUnmarshaler:
			meta.isStructList = true
			skip, err := visitBranch(visitor, meta)
			if err != nil {
				return err
			}
			if skip {
				continue
			}
			if err := walkStructList(elemVal, elemType, meta, scope, fieldKeyPath, opts, visitor); err != nil {
				return err
			}

		case elemType.Kind() == reflect.Struct && !isTextUnmarshaler:
			skip, err := visitBranch(visitor, meta)
			if err != nil {
				return err
			}
			if skip || scope.descendsInto(elemType) {
				continue
			}
			next := walkScope{
				keyPath:     fieldKeyPath,
				listDepth:   scope.listDepth,
				listRoot:    scope.listRoot,
				listPath:    meta.listPath,
				listKeyPath: scope.listKeyPath,
				ancestors:   append(append([]reflect.Type{}, scope.ancestors...), elemType),
			}
			if err := walkStructValue(elemVal, elemType, next, opts, visitor); err != nil {
				return err
			}

		default:
			if visitor.leaf != nil {
				if err := visitor.leaf(meta); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func visitBranch(visitor structVisitor, m fieldMeta) (bool, error) {
	if visitor.branch == nil {
		return false, nil
	}
	return visitor.branch(m)
}

func walkStructList(elemVal reflect.Value, elemType reflect.Type, m fieldMeta, scope walkScope, keyPath []string, opts Options, visitor structVisitor) error {
	inner := elemType.Elem()
	for inner.Kind() == reflect.Pointer {
		inner = inner.Elem()
	}
	if scope.descendsInto(inner) {
		return nil
	}

	next := walkScope{
		keyPath:     keyPath,
		listDepth:   scope.listDepth + 1,
		listRoot:    scope.listRoot,
		listPath:    m.listPath,
		listKeyPath: scope.listKeyPath,
		ancestors:   append(append([]reflect.Type{}, scope.ancestors...), inner),
	}
	if scope.listDepth == 0 {
		// The outermost list is where a column's defaults are read from, and the key its config
		// file rows arrive under: collectColumn walks down through every list below it on its own.
		next.listRoot, next.listPath, next.listKeyPath = elemVal, nil, keyPath
	}

	// There is no single element to read defaults from, so the walk carries the first one where
	// there is one, and a zero value otherwise. Per-field defaults come from the column (see
	// collectColumn), not from this value.
	exemplar := reflect.Zero(inner)
	if elemVal.IsValid() && elemVal.Kind() != reflect.Pointer && elemVal.Len() > 0 {
		first := elemVal.Index(0)
		for first.Kind() == reflect.Pointer {
			if first.IsNil() {
				first = reflect.Zero(inner)
				break
			}
			first = first.Elem()
		}
		exemplar = first
	}
	return walkStructValue(exemplar, inner, next, opts, visitor)
}
