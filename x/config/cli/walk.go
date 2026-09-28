package cli

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// fieldMeta is one leaf at one path through a config struct.
type fieldMeta struct {
	// key is dotted from the root struct to this field, in kebab case. Flattened fields add no segment.
	key        string
	configPath []string
	field      reflect.StructField
	// owner declares field and so holds its DocComments, even when field is promoted.
	owner reflect.Type
	// elem is field's value one pointer level down, or the zero value if that pointer is nil.
	elem reflect.Value
	// goPath is the Go field names from the root, which decoding writes through and validation
	// errors are mapped back by.
	goPath []string
	// inOptional marks a field under a pointer section, which may be absent as a whole.
	inOptional bool
}

// readsText reports whether t, or a pointer to it, parses its own text, which is all a flag or env
// var carries.
func readsText(t reflect.Type) bool {
	return t.Implements(textUnmarshaler) || reflect.PointerTo(t).Implements(textUnmarshaler)
}

var textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

// isText reports whether t has a single text form, so it can be a list element or a map key or
// value on the command line: a scalar, a text type, or []byte.
func isText(t reflect.Type) bool {
	t = commentparsing.DerefType(t)
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		return t.Elem().Kind() == reflect.Uint8 || readsText(t)
	default:
		return readsText(t)
	}
}

// bindStruct binds every field of entry's target the markup reads, in declaration order.
func bindStruct(entry *targetEntry) error {
	v := reflect.ValueOf(entry.target)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return errors.New("target pointer cannot be nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("target must be a struct or pointer to struct")
	}
	return entry.walk(v, walkScope{ancestors: []reflect.Type{v.Type()}})
}

// walkScope is the walk's current position.
type walkScope struct {
	key        string
	configPath []string
	goPath     []string
	// inOptional is set once the walk passes through a pointer section.
	inOptional bool
	// ancestors stops recursion into self-referential types.
	ancestors []reflect.Type
}

// walk binds v's fields. A leaf is a scalar, a type the markup reads whole, or a struct the walk is
// already inside, which the config file decodes whole.
func (e *targetEntry) walk(v reflect.Value, scope walkScope) error {
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		key, ok := e.b.opts.Markup.Key(field)
		// The markup's decoder never reads it, so no flag or env var may either.
		if !ok {
			continue
		}
		if key == "-" {
			return fmt.Errorf("%s: %s is named %q, which can't be a flag; give it another name", t, field.Name, "-")
		}

		if !field.IsExported() && key != "" {
			continue
		}

		fieldKey, configPath := scope.key, scope.configPath
		if key != "" {
			// Kebab case, so a rename reads as if it were the name.
			fieldKey = strings.TrimPrefix(scope.key+"."+strcase.ToKebab(key), ".")
			configPath = append(slices.Clone(scope.configPath), key)
		}
		goPath := append(slices.Clone(scope.goPath), field.Name)

		elem := v.Field(i)
		if field.Type.Kind() == reflect.Pointer {
			if elem.IsNil() {
				elem = reflect.Zero(field.Type.Elem())
			} else {
				elem = elem.Elem()
			}
		}

		isLeaf := e.b.opts.Markup.IsLeaf(field.Type) || e.b.opts.Markup.IsLeaf(elem.Type())
		if elem.Kind() != reflect.Struct || isLeaf || slices.Contains(scope.ancestors, elem.Type()) {
			meta := fieldMeta{fieldKey, configPath, field, t, elem, goPath, scope.inOptional}
			if err := bindLeafFlag(e, meta); err != nil {
				return err
			}
			continue
		}

		// reflect can't allocate an unexported embedded pointer, so none of its fields could be set.
		if field.Anonymous && !field.IsExported() && field.Type.Kind() == reflect.Pointer {
			return fmt.Errorf("embedded *%s is unexported, so nothing can allocate it; export the type or embed it by value",
				elem.Type().Name())
		}
		e.sections[strings.Join(goPath, ".")] = strings.Join(e.namespaced(configPath), ".")
		next := walkScope{
			key:        fieldKey,
			configPath: configPath,
			goPath:     goPath,
			inOptional: scope.inOptional || field.Type.Kind() == reflect.Pointer,
			ancestors:  append(slices.Clone(scope.ancestors), elem.Type()),
		}
		if err := e.walk(elem, next); err != nil {
			return err
		}
	}
	return nil
}
