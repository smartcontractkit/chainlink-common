package decode

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

// flattenedMarker names an embedded field the markup flattens, so configKey can drop its segment. It can be any string
// that is non-empty (the validator replaces "" with the Go name), is never a config key, and has no '.', '[' or ']'
// (configKey rewrites and splits on those before comparing, so a marker with them would never match).
const flattenedMarker = "\x00"

// With config key names, the validator's namespaces differ from config keys only by the root type's name, [i] indices,
// and flattened embeds, which configKey fixes.
func NewValidator(lang Markup) *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		key, read := lang.Key(f)
		switch {
		case !read:
			return f.Name
		case key == "":
			return flattenedMarker
		default:
			return key
		}
	})

	return v
}

// RuleError names fields in the error by config key rather than Go path: a field renamed by its tag appears in the
// message under its new name. root is the struct's own type, dereferenced; namespace, if set, is the first segment of
// every key, as the cli's RegisterInNamespace sets it.
func RuleError(lang Markup, namespace string, root reflect.Type, fe validator.FieldError) error {
	rule := fe.Tag()
	if params := strings.Fields(fe.Param()); len(params) > 0 {
		ns := fe.Namespace()
		parent := ns[:strings.LastIndex(ns, ".")]
		siblings := parentType(root, fe.StructNamespace())
		for i, p := range params {
			if f, ok := siblings.FieldByName(p); ok {
				if key, _ := lang.Key(f); key != "" {
					params[i] = configKey(namespace, parent+"."+key)
				}
			}
		}

		rule += "=" + strings.Join(params, " ")
	}

	return fmt.Errorf("%s failed on the '%s' tag", configKey(namespace, fe.Namespace()), rule)
}

func parentType(root reflect.Type, structNS string) reflect.Type {
	t := root
	segments := strings.Split(structNS, ".")
	for _, segment := range segments[1 : len(segments)-1] {
		name, rest, indexed := strings.Cut(segment, "[")
		f, _ := t.FieldByName(name)
		t = DerefType(f.Type)
		for ; indexed; _, rest, indexed = strings.Cut(rest, "[") {
			t = DerefType(t.Elem())
		}
	}

	return t
}

var indexReplacer = strings.NewReplacer("[", ".", "]", "")

// List indices and map keys become segments, as chainlink-common's pkg/config.Validate names them: Nodes.1.Name.
func configKey(namespace, ns string) string {
	_, ns, _ = strings.Cut(ns, ".")
	if namespace != "" {
		ns = namespace + "." + ns
	}

	segments := strings.Split(indexReplacer.Replace(ns), ".")
	return strings.Join(slices.DeleteFunc(segments, func(s string) bool { return s == flattenedMarker }), ".")
}
