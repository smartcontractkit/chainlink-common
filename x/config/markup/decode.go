package markup

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/smartcontractkit/chainlink-common/x/config/internal/decode"
)

// Decode decodes one document into v strictly, as the cli reads a config file, then checks its validate tags. Map keys
// are parsed as text, so a key type the language can't read itself still works, and two texts for one key are an
// error. Failures are named by config key.
//
// A key the document leaves out leaves v's own value alone, so defaults can stay in code; only pointer sections on the
// way to a set field are allocated, so an untouched optional *struct stays nil. A map's entries merge by key over the
// field's own value, as they do over a default at every cli source.
func Decode(lang Markup, data []byte, v any) error {
	if lang == nil {
		return Err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("target must be a non-nil pointer to struct")
	}

	if rv = rv.Elem(); rv.Kind() != reflect.Struct {
		return errors.New("target must be a non-nil pointer to struct")
	}

	keys := &decode.KeyNode{}
	var leaves []fileLeaf
	if err := walkKeys(rv.Type(), lang, nil, nil, []reflect.Type{rv.Type()}, keys, &leaves); err != nil {
		return err
	}

	// Pointers tell a key the document left out from one set to the zero value.
	fileValues := reflect.New(keys.FileValuesType(lang))
	if err := lang.Unmarshal(data, fileValues.Interface()); err != nil {
		return err
	}

	for _, leaf := range leaves {
		raw, ok := keys.Lookup(fileValues.Elem(), leaf.fileKey)
		if !ok {
			continue
		}

		val, err := decode.FromFile(leaf.goType, raw)
		if err != nil {
			return fmt.Errorf("%s: %w", strings.Join(leaf.fileKey, "."), err)
		}

		f := decode.AllocateField(rv, leaf.goPath)
		for f.Kind() == reflect.Pointer {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		if decode.MergesByKey(f.Type(), lang) {
			f.Set(decode.MergeMaps(f.Type(), f, val))
		} else {
			f.Set(val)
		}
	}

	err := decode.NewValidator(lang).Struct(rv.Addr().Interface())
	fieldErrs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err
	}

	errs := make([]error, len(fieldErrs))
	for i, fe := range fieldErrs {
		errs[i] = decode.RuleError(lang, "", rv.Type(), fe)
	}

	return errors.Join(errs...)
}

// fileLeaf is one settable field of v: fileKey names it in a document, goPath and goType name it in v.
type fileLeaf struct {
	fileKey []string
	goPath  []string
	goType  reflect.Type
}

// walkKeys builds t's key tree in keys and its leaves in leaves, as the cli builds one for a command's structs: a field
// the markup flattens adds no key segment, a field it never reads is left out, and two fields one document key would
// set are an error.
func walkKeys(t reflect.Type, lang Markup, fileKey, goPath []string, ancestors []reflect.Type, keys *decode.KeyNode, leaves *[]fileLeaf) error {
	for f := range t.Fields() {
		key, read := lang.Key(f)
		// The markup's decoder never reads it, so nothing else may either.
		if !read || (!f.IsExported() && key != "") {
			continue
		}

		elem := decode.DerefType(f.Type)
		fieldPath := append(slices.Clone(goPath), f.Name)
		childKey := fileKey
		if key != "" {
			childKey = append(slices.Clone(fileKey), key)
		}

		// A struct already being walked is a leaf, so a recursive type can't walk forever.
		if elem.Kind() != reflect.Struct || lang.IsLeaf(f.Type) || lang.IsLeaf(elem) ||
			slices.Contains(ancestors, elem) {
			fileType, err := decode.FileValueType(elem, lang, map[reflect.Type]bool{})
			if err != nil {
				return fmt.Errorf("%s: %w", f.Name, err)
			}

			if err = keys.Add(childKey, fileType, f.Tag, lang); err != nil {
				return err
			}

			*leaves = append(*leaves, fileLeaf{childKey, fieldPath, elem})
			continue
		}

		// reflect can't allocate an unexported embedded pointer, so none of its fields could be set.
		if f.Anonymous && !f.IsExported() && f.Type.Kind() == reflect.Pointer {
			return fmt.Errorf("embedded *%s is unexported, so nothing can allocate it; "+
				"export the type or embed it by value", elem.Name())
		}

		if err := walkKeys(elem, lang, childKey, fieldPath, append(slices.Clone(ancestors), elem), keys, leaves); err != nil {
			return err
		}
	}

	return nil
}
