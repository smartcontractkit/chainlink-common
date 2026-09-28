package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

type targetEntry struct {
	b         *Binder
	cmd       *cobra.Command
	namespace string
	target    any
	keys      []leafKey

	// suppliedFields addresses of every leaf a source set, allowing us to differentiate a decoded default from unset.
	suppliedFields map[uintptr]bool

	docs         fieldDocs
	undocumented []string

	// sections maps each nested struct's dotted Go key to its config key, to name it in errors.
	sections map[string]string
}

// leafKey ties a field to where each source holds it.
type leafKey struct {
	// key is the dotted kebab-case key the flag and env vars are named from.
	key string

	configPath []string

	// flagName is key with underscores as dashes.
	flagName string

	// flag is nil for a config-file-only field, which has no env var either. value is its parsed value.
	flag  *pflag.Flag
	value func() reflect.Value

	// goPath and goType locate and type the field without re-walking the struct.
	goPath []string
	goType reflect.Type
}

func (e *targetEntry) envVars(k leafKey) []string {
	names := make([]string, len(e.b.opts.Prefixes))
	for i, prefix := range e.b.opts.Prefixes {
		names[i] = envVarName(prefix, k.key)
	}
	return names
}

func (e *targetEntry) namespaced(path []string) []string {
	if e.namespace == "" {
		return path
	}
	return append(strings.Split(e.namespace, "."), path...)
}

// sources returns k's value from each source that set it, highest precedence first: changed flag,
// the first env var set, config file. Each is of k's field type through pointers; an env var is
// parsed as the flag is. None leaves the struct's default.
func (e *targetEntry) sources(k leafKey) ([]reflect.Value, error) {
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
	if raw, ok := e.b.keys.lookup(e.b.config, k.configPath); ok {
		val, err := fromFile(commentparsing.DerefType(k.goType), raw)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
	return vals, nil
}

// setTag requires that a source supplied the field.
const setTag = "set"

// isSupplied is the `set` rule. The validator hands it a field but not its path, so the field is
// matched by address.
func (e *targetEntry) isSupplied(fl validator.FieldLevel) bool {
	field := fl.Field()
	return field.CanAddr() && e.suppliedFields[field.Addr().Pointer()]
}

// validate runs the `validate` tags against the decoded target, after a full decode so everything
// is populated. A failing field is named by its config key, and a leaf by the sources that set it
// too, not by its Go path.
func (e *targetEntry) validate() error {
	v := validator.New()
	// Fails only if setTag collides with a built-in rule.
	if err := v.RegisterValidation(setTag, e.isSupplied, true); err != nil {
		return err
	}

	err := v.Struct(e.target)
	fieldErrs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err
	}

	leaves := make(map[string]leafKey, len(e.keys))
	for _, k := range e.keys {
		leaves[strings.Join(k.goPath, ".")] = k
	}
	errs := make([]error, len(fieldErrs))
	for i, fe := range fieldErrs {
		// StructNamespace starts with the root struct's type name, which goKey doesn't hold.
		_, goKey, _ := strings.Cut(fe.StructNamespace(), ".")
		errs[i] = fe
		if k, ok := leaves[goKey]; ok {
			var sources []string
			if k.flag != nil {
				sources = append([]string{"--" + k.flagName}, e.envVars(k)...)
			}
			sources = append(sources, strings.Join(k.configPath, ".")+" in a config file")
			errs[i] = fmt.Errorf("%w; set it with %s", e.ruleError(goKey, fe), strings.Join(sources, ", "))
		} else if _, ok := e.keyOf(goKey); ok {
			errs[i] = e.ruleError(goKey, fe)
		}
	}
	return errors.Join(errs...)
}

// ruleError names the field at goKey, and any sibling a cross-field rule names, by config key.
func (e *targetEntry) ruleError(goKey string, fe validator.FieldError) error {
	rule := fe.Tag()
	if params := strings.Fields(fe.Param()); len(params) > 0 {
		parent := goKey[:max(0, strings.LastIndex(goKey, "."))]
		for i, p := range params {
			if key, ok := e.keyOf(strings.TrimPrefix(parent+"."+p, ".")); ok {
				params[i] = key
			}
		}
		rule += "=" + strings.Join(params, " ")
	}
	key, _ := e.keyOf(goKey)
	return fmt.Errorf("%s failed on the '%s' tag", key, rule)
}

// keyOf is the config key for goKey, the validator's path to a field. For a field inside a leaf or
// section, such as a list element, it is that leaf's or section's config key followed by the path
// below it, indices as segments the way the core node names them (Nodes.1.Name, not Nodes[1].Name).
func (e *targetEntry) keyOf(goKey string) (string, bool) {
	for prefix := goKey; ; {
		if key, ok := e.sections[prefix]; ok {
			return key + e.elementKey(nil, goKey[len(prefix):]), true
		}
		for _, k := range e.keys {
			if strings.Join(k.goPath, ".") == prefix {
				return strings.Join(k.configPath, ".") + e.elementKey(k.goType, goKey[len(prefix):]), true
			}
		}
		i := strings.LastIndexAny(prefix, ".[")
		if i < 0 {
			return "", false
		}
		prefix = prefix[:i]
	}
}

// elementKey turns below, the validator's path under a value of type t such as "[0].SecretName",
// into config key segments such as ".0.secret_name": an index or map key is a segment, a field is
// the markup's key for it, and a flattened field adds nothing. Where t is nil or can't be followed,
// fields keep their Go names.
func (e *targetEntry) elementKey(t reflect.Type, below string) string {
	var out strings.Builder
	for below != "" {
		if t != nil {
			t = commentparsing.DerefType(t)
		}
		if below[0] == '[' {
			end := strings.IndexByte(below, ']')
			out.WriteString("." + below[1:end])
			below = below[end+1:]
			if t != nil && (t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map) {
				t = t.Elem()
			} else {
				t = nil
			}
			continue
		}
		below = below[1:]
		end := strings.IndexAny(below, ".[")
		if end < 0 {
			end = len(below)
		}
		name := below[:end]
		below = below[end:]
		f, ok := reflect.StructField{}, false
		if t != nil && t.Kind() == reflect.Struct {
			f, ok = t.FieldByName(name)
		}
		if !ok {
			out.WriteString("." + name)
			t = nil
			continue
		}
		if key, _ := e.b.opts.Markup.Key(f); key != "" {
			out.WriteString("." + key)
		}
		t = f.Type
	}
	return out.String()
}
