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

type fieldMeta struct {
	key     string
	fileKey []string
	field   reflect.StructField
	// owner holds field's DocComments, even when field is promoted.
	owner  reflect.Type
	elem   reflect.Value
	goPath []string
	// inOptional: a pointer section may be absent as a whole, so its rules don't mark flags required.
	inOptional bool
}

// A flag or env var carries only text.
func readsText(t reflect.Type) bool {
	return t.Implements(textUnmarshaler) || reflect.PointerTo(t).Implements(textUnmarshaler)
}

var textUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

func canText(t reflect.Type) bool {
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

type walkScope struct {
	key        string
	fileKey    []string
	goPath     []string
	inOptional bool
	// ancestors stops recursion into self-referential types.
	ancestors []reflect.Type
}

func (e *targetEntry) walk(v reflect.Value, scope walkScope) error {
	for i := range v.NumField() {
		if err := e.walkField(v, i, scope); err != nil {
			return err
		}
	}

	return nil
}

func (e *targetEntry) walkField(v reflect.Value, i int, scope walkScope) error {
	field := v.Type().Field(i)
	key, ok := e.b.opts.Markup.Key(field)
	// The markup's decoder never reads it, so no flag or env var may either.
	if !ok {
		return nil
	}

	if key == "-" {
		return fmt.Errorf("%s: %s is named %q, which can't be a flag; give it another name", v.Type(), field.Name, "-")
	}

	if !field.IsExported() && key != "" {
		return nil
	}

	next := scope.child(field, key)
	elem := derefOrZero(v.Field(i))
	if e.isLeaf(field, elem, scope) {
		return bindLeafFlag(e, fieldMeta{next.key, next.fileKey, field, v.Type(), elem, next.goPath, scope.inOptional})
	}

	// reflect can't allocate an unexported embedded pointer, so none of its fields could be set.
	if field.Anonymous && !field.IsExported() && field.Type.Kind() == reflect.Pointer {
		return fmt.Errorf("embedded *%s is unexported, so nothing can allocate it; "+
			"export the type or embed it by value",
			elem.Type().Name())
	}

	next.inOptional = scope.inOptional || field.Type.Kind() == reflect.Pointer
	next.ancestors = append(slices.Clone(scope.ancestors), elem.Type())
	return e.walk(elem, next)
}

// A flattened field, whose key is "", adds nothing to the key paths.
func (s walkScope) child(field reflect.StructField, key string) walkScope {
	next := s
	next.goPath = append(slices.Clone(s.goPath), field.Name)
	if key != "" {
		// Kebab case, so a rename reads as if it were the name.
		next.key = strings.TrimPrefix(s.key+"."+strcase.ToKebab(key), ".")
		next.fileKey = append(slices.Clone(s.fileKey), key)
	}

	return next
}

// A nil pointer section still gets its flags, so its zero value stands in for it.
func derefOrZero(f reflect.Value) reflect.Value {
	for f.Kind() == reflect.Pointer {
		if f.IsNil() {
			return reflect.Zero(commentparsing.DerefType(f.Type()))
		}

		f = f.Elem()
	}

	return f
}

// A struct already being walked is a leaf, callers are expected to have dereferenced the field already.
func (e *targetEntry) isLeaf(field reflect.StructField, elem reflect.Value, scope walkScope) bool {
	lang := e.b.opts.Markup
	return elem.Kind() != reflect.Struct || lang.IsLeaf(field.Type) || lang.IsLeaf(elem.Type()) ||
		slices.Contains(scope.ancestors, elem.Type())
}
