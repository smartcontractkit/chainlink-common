package cli

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/internal/decode"
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

func canText(t reflect.Type) bool {
	t = commentparsing.DerefType(t)
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		return t.Elem().Kind() == reflect.Uint8 || decode.ReadsText(t)
	default:
		return decode.ReadsText(t)
	}
}

func bindStruct[T any](entry *typedEntry[T]) error {
	v := reflect.ValueOf(entry.dst)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return errors.New("target pointer cannot be nil")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("target must be a struct or pointer to struct")
	}

	scope := walkScope{ancestors: []reflect.Type{v.Type()}}
	if ns := entry.ns; ns != "" {
		scope.key, scope.fileKey = kebab(ns), []string{ns}
	}

	return entry.walk(v, scope)
}

type walkScope struct {
	key        string
	fileKey    []string
	goPath     []string
	inOptional bool
	// ancestors stops recursion into self-referential types.
	ancestors []reflect.Type
	// entry is set inside a map's entries, where goPath restarts at each entry.
	entry *entryScope
}

func (e *typedEntry[T]) walk(v reflect.Value, scope walkScope) error {
	for i := range v.NumField() {
		if err := e.walkField(v, i, scope); err != nil {
			return err
		}
	}

	return nil
}

func (e *typedEntry[T]) walkField(v reflect.Value, i int, scope walkScope) error {
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
		doc := fieldDoc(v.Type(), field.Name)
		if scope.entry != nil {
			_, err := e.bindEntryLeaf(next, field.Type, doc)
			return err
		}

		leaf, err := bindLeafFlag(e, fieldMeta{next.key, next.fileKey, field, v.Type(), elem, next.goPath, scope.inOptional}, doc)
		if err != nil {
			return err
		}

		if leaf.each, err = e.bindEntries(&leaf, field.Type, next, doc); err != nil {
			return err
		}

		e.addKey(leaf)
		return nil
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
		next.key = strings.TrimPrefix(s.key+"."+kebab(key), ".")
		next.fileKey = append(slices.Clone(s.fileKey), key)
	}

	return next
}

// kebab is name in kebab case. Digits stay with the word before them, and a lone s after an acronym is its plural,
// so V2Bootstrappers is v2-bootstrappers and OnchainKeyBundleIDs is onchain-key-bundle-ids.
func kebab(name string) string {
	r := []rune(name)
	var b strings.Builder
	for i, c := range r {
		if c == '_' || c == '-' {
			b.WriteByte('-')
			continue
		}

		if i > 0 && unicode.IsUpper(c) && startsWord(r, i) {
			b.WriteByte('-')
		}

		b.WriteRune(unicode.ToLower(c))
	}

	return b.String()
}

// startsWord reports whether the capital at r[i] begins a word: after a lowercase letter or a digit, or as the last
// capital of an acronym before a lowercase word, as the S in HTTPServer. A lone s after an acronym is its plural.
func startsWord(r []rune, i int) bool {
	prev := r[i-1]
	if unicode.IsLower(prev) || unicode.IsDigit(prev) {
		return true
	}

	if !unicode.IsUpper(prev) || i+1 >= len(r) || !unicode.IsLower(r[i+1]) {
		return false
	}

	plural := r[i+1] == 's' && (i+2 == len(r) || !unicode.IsLower(r[i+2]))
	return !plural
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
func (e *typedEntry[T]) isLeaf(field reflect.StructField, elem reflect.Value, scope walkScope) bool {
	lang := e.b.opts.Markup
	return decode.IsDiscard(field.Type) || elem.Kind() != reflect.Struct || lang.IsLeaf(field.Type) ||
		lang.IsLeaf(elem.Type()) || slices.Contains(scope.ancestors, elem.Type())
}
