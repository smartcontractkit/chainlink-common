package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// flattenedMarker names an embedded field the markup flattens, so configKey can drop its segment. It can be any string
// that is non-empty (the validator replaces "" with the Go name), is never a config key, and has no '.', '[' or ']'
// (configKey rewrites and splits on those before comparing, so a marker with them would never match).
const flattenedMarker = "\x00"

type targetEntry interface {
	binder() *Binder
	command() *cobra.Command
	keys() []leafKey
	addKey(k leafKey)
	undocumented() []string
	addUndocumented(key string)
	envVars(k leafKey) []string

	// Only pointer sections on the way to a set field are allocated, so an untouched optional *struct stays nil for
	// `required_without` and `excluded_with`.
	decode(cc commandConfig) error

	// Runs after the full decode so cross-field rules see every value.
	validate() error
}

type leafKey struct {
	// key names the flag and env vars, such as server.listen-addr.
	key string

	fileKey []string

	flagName string

	// flag is nil for a config-file-only field, which has no env var either.
	flag  *pflag.Flag
	value func() reflect.Value

	// goPath and goType save walking the struct again.
	goPath []string
	goType reflect.Type
}

type typedEntry[T any] struct {
	b      *Binder
	cmd    *cobra.Command
	dst    *T
	leaves []leafKey

	undocumentedKeys []string
}

func (e *typedEntry[T]) binder() *Binder { return e.b }

func (e *typedEntry[T]) command() *cobra.Command { return e.cmd }

func (e *typedEntry[T]) keys() []leafKey { return e.leaves }

func (e *typedEntry[T]) addKey(k leafKey) { e.leaves = append(e.leaves, k) }

func (e *typedEntry[T]) undocumented() []string { return e.undocumentedKeys }

func (e *typedEntry[T]) addUndocumented(key string) {
	e.undocumentedKeys = append(e.undocumentedKeys, key)
}

func (e *typedEntry[T]) envVars(k leafKey) []string {
	name := strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(k.key))
	if len(e.b.opts.Prefixes) == 0 {
		return []string{name}
	}

	names := make([]string, len(e.b.opts.Prefixes))
	for i, prefix := range e.b.opts.Prefixes {
		names[i] = name
		if prefix != "" {
			names[i] = strings.TrimSuffix(strings.ToUpper(prefix), "_") + "_" + name
		}
	}

	return names
}

// Highest precedence first. An env var parses as the flag would, so both accept the same input.
func (e *typedEntry[T]) sources(k leafKey, cc commandConfig) ([]reflect.Value, error) {
	var vals []reflect.Value
	if k.flag != nil {
		if k.flag.Changed {
			vals = append(vals, k.value())
		}

		for _, name := range e.envVars(k) {
			// Empty counts as unset, so a stray export can't blank a file value.
			if s := os.Getenv(name); s != "" {
				val, err := parseText(k.goType, s)
				if err != nil {
					return nil, fmt.Errorf("invalid value %q for %s: %w", s, name, err)
				}

				vals = append(vals, val)
				break
			}
		}
	}

	if raw, ok := cc.keys.lookup(cc.fileValues, k.fileKey); ok {
		vals = append(vals, raw)
	}

	return vals, nil
}

func (e *typedEntry[T]) validate() error {
	v := validator.New()
	// With config key names, the validator's namespaces differ from config keys only by the root type's name, [i]
	// indices, and flattened embeds, which configKey fixes.
	lang := e.b.opts.Markup
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

	err := v.Struct(e.dst)
	fieldErrs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err
	}

	leaves := make(map[string]leafKey, len(e.leaves))
	for _, k := range e.leaves {
		leaves[strings.Join(k.goPath, ".")] = k
	}

	errs := make([]error, len(fieldErrs))
	for i, fe := range fieldErrs {
		errs[i] = e.ruleError(fe)
		// StructNamespace starts with the root struct's type name, which goPath doesn't hold.
		_, goKey, _ := strings.Cut(fe.StructNamespace(), ".")
		if k, ok := leaves[goKey]; ok {
			var sources []string
			if k.flag != nil {
				sources = append([]string{"--" + k.flagName}, e.envVars(k)...)
			}

			sources = append(sources, strings.Join(k.fileKey, ".")+" in a config file")
			errs[i] = fmt.Errorf("%w; set it with %s", errs[i], strings.Join(sources, ", "))
		}
	}

	return errors.Join(errs...)
}

// Names fields in the error by config key rather than Go path. For example, a field renamed by its tag, which also
// renames its flag, env vars, and config key, appears in the message under its new name.
func (e *typedEntry[T]) ruleError(fe validator.FieldError) error {
	rule := fe.Tag()
	if params := strings.Fields(fe.Param()); len(params) > 0 {
		ns := fe.Namespace()
		parent := ns[:strings.LastIndex(ns, ".")]
		siblings := e.parentType(fe.StructNamespace())
		for i, p := range params {
			if f, ok := siblings.FieldByName(p); ok {
				if key, _ := e.b.opts.Markup.Key(f); key != "" {
					params[i] = configKey(parent + "." + key)
				}
			}
		}

		rule += "=" + strings.Join(params, " ")
	}

	return fmt.Errorf("%s failed on the '%s' tag", configKey(fe.Namespace()), rule)
}

func (e *typedEntry[T]) parentType(structNS string) reflect.Type {
	t := commentparsing.DerefType(reflect.TypeFor[T]())
	segments := strings.Split(structNS, ".")
	for _, segment := range segments[1 : len(segments)-1] {
		name, rest, indexed := strings.Cut(segment, "[")
		f, _ := t.FieldByName(name)
		t = commentparsing.DerefType(f.Type)
		for ; indexed; _, rest, indexed = strings.Cut(rest, "[") {
			t = commentparsing.DerefType(t.Elem())
		}
	}

	return t
}

func (e *typedEntry[T]) decode(cc commandConfig) error {
	dst := reflect.ValueOf(e.dst).Elem()
	for _, k := range e.leaves {
		vals, err := e.sources(k, cc)
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

		f.Set(vals[0])
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

// List indices and map keys become segments, as chainlink-common's pkg/config.Validate names them: Nodes.1.Name.
func configKey(ns string) string {
	_, ns, _ = strings.Cut(ns, ".")
	segments := strings.Split(strings.NewReplacer("[", ".", "]", "").Replace(ns), ".")
	return strings.Join(slices.DeleteFunc(segments, func(s string) bool { return s == flattenedMarker }), ".")
}
