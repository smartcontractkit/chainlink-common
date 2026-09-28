package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// durationType tells time.Duration apart from int64, which shares its Kind.
var durationType = reflect.TypeFor[time.Duration]()

func bindLeafFlag(entry *targetEntry, m fieldMeta) error {
	flags := entry.cmd.PersistentFlags()
	key := strings.Join(entry.namespaced([]string{m.key}), ".")
	leaf := leafKey{
		key:        key,
		configPath: entry.namespaced(m.configPath),
		flagName:   strings.ReplaceAll(key, "_", "-"),
		goPath:     m.goPath,
		goType:     m.field.Type,
	}
	// pflag panics on a redefinition.
	if flags.Lookup(leaf.flagName) != nil {
		return fmt.Errorf("%s: flag --%s is already defined on %s; register one of the structs under another namespace or command",
			m.key, leaf.flagName, entry.cmd.Name())
	}

	doc := entry.docs.usage(m.owner, m.field.Name)
	usage := doc
	// Under a pointer section the rule applies only once the section is configured.
	if isRequired(m.field) && !m.inOptional {
		usage = strings.TrimSpace(usage + " (required)")
	}
	if envs := entry.envVars(leaf); len(envs) > 0 {
		usage = strings.TrimSpace(usage + " [env " + strings.Join(envs, ", ") + "]")
	}

	f, get := newFlag(leaf.flagName, m.field.Type, m.elem, usage)
	if f != nil {
		flags.AddFlag(f)
		leaf.flag, leaf.value = f, get
		if doc == "" {
			entry.undocumented = append(entry.undocumented, leaf.key)
		}
	}
	entry.keys = append(entry.keys, leaf)
	return nil
}

// isRequired reports whether field's `validate` tag demands a value: `required`, or `set`.
func isRequired(field reflect.StructField) bool {
	for rule := range strings.SplitSeq(field.Tag.Get("validate"), ",") {
		switch rule {
		case "required", setTag:
			return true
		case "dive":
			// The rules after it are a list's or map's elements' (and, between keys and endkeys, a
			// map's keys'), not the field's own: `dive,required` requires every element, not the list.
			return false
		}
	}
	return false
}

// newFlag builds, unregistered, the flag for a field of type t with default def, and a getter for
// its parsed value, of t through its pointers. It returns nil for a type with no text form, which is config file only.
func newFlag(name string, t reflect.Type, def reflect.Value, usage string) (*pflag.Flag, func() reflect.Value) {
	t = commentparsing.DerefType(t)
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	var get func() reflect.Value
	switch {
	case readsText(t):
		v := &textValue{value: reflect.New(t).Elem()}
		if def.IsValid() && def.Type() == t {
			v.value.Set(def)
		}
		fs.Var(v, name, usage)
		get = func() reflect.Value { return v.value }
	case t == durationType:
		get = typed(fs.DurationVar, name, time.Duration(def.Int()), usage)
	case t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8:
		// []byte is text (JSON, PEM), not a list of numbers.
		text := typed(fs.StringVar, name, string(def.Bytes()), usage)
		get = func() reflect.Value { return reflect.ValueOf([]byte(text().String())) }
	case t.Kind() == reflect.Slice:
		if !isText(t.Elem()) {
			return nil, nil
		}
		l := &textListValue{list: reflect.MakeSlice(t, 0, 0), elems: textListOf(def)}
		fs.Var(l, name, usage)
		get = func() reflect.Value { return l.list }
	case t.Kind() == reflect.Map:
		// Only text keys and values have a flag form (see maps.go).
		if !isText(t.Key()) || !isText(t.Elem()) {
			return nil, nil
		}
		m := &textMapValue{m: reflect.MakeMap(t), entries: textMapOf(def), seen: map[any]string{}}
		fs.Var(m, name, usage)
		get = func() reflect.Value { return m.m }
	default:
		//nolint:gosec // G115: each conversion is to def's own kind, so none truncates
		switch t.Kind() {
		case reflect.String:
			get = typed(fs.StringVar, name, def.String(), usage)
		case reflect.Bool:
			get = typed(fs.BoolVar, name, def.Bool(), usage)
		case reflect.Int:
			get = typed(fs.IntVar, name, int(def.Int()), usage)
		case reflect.Int8:
			get = typed(fs.Int8Var, name, int8(def.Int()), usage)
		case reflect.Int16:
			get = typed(fs.Int16Var, name, int16(def.Int()), usage)
		case reflect.Int32:
			get = typed(fs.Int32Var, name, int32(def.Int()), usage)
		case reflect.Int64:
			get = typed(fs.Int64Var, name, def.Int(), usage)
		case reflect.Uint:
			get = typed(fs.UintVar, name, uint(def.Uint()), usage)
		case reflect.Uint8:
			get = typed(fs.Uint8Var, name, uint8(def.Uint()), usage)
		case reflect.Uint16:
			get = typed(fs.Uint16Var, name, uint16(def.Uint()), usage)
		case reflect.Uint32:
			get = typed(fs.Uint32Var, name, uint32(def.Uint()), usage)
		case reflect.Uint64:
			get = typed(fs.Uint64Var, name, def.Uint(), usage)
		case reflect.Float32:
			get = typed(fs.Float32Var, name, float32(def.Float()), usage)
		case reflect.Float64:
			get = typed(fs.Float64Var, name, def.Float(), usage)
		default:
			return nil, nil
		}
		// pflag prints the raw value, but a type's own MarshalText may redact it, as
		// pkg/config.SecretString does. A zero default stays blank rather than looking set.
		if !def.IsZero() {
			if text, ok := marshalText(def); ok {
				fs.Lookup(name).DefValue = text
			}
		}
	}
	// A named type, such as type Port int, takes pflag's value converted.
	return fs.Lookup(name), func() reflect.Value { return get().Convert(t) }
}

// typed defines a flag with one of pflag's XxxVar methods and returns a getter for its value.
func typed[V any](define func(*V, string, V, string), name string, def V, usage string) func() reflect.Value {
	p := new(V)
	define(p, name, def, usage)
	return func() reflect.Value { return reflect.ValueOf(*p) }
}

// textValue is a text type's flag: the type parses its own text, as the flag is parsed.
type textValue struct{ value reflect.Value }

func (v *textValue) Set(s string) error {
	p := reflect.New(v.value.Type())
	if err := p.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s)); err != nil {
		return err
	}
	v.value.Set(p.Elem())
	return nil
}

// String renders the default with the type's own marshaller, so it parses back; %v can print struct fields.
func (v *textValue) String() string { return textOf(v.value) }

func (v *textValue) Type() string { return "string" }

// parseText parses s as a flag of type t would, into a value of t through its pointers. A type with no text form is an error.
func parseText(t reflect.Type, s string) (reflect.Value, error) {
	f, get := newFlag("value", t, reflect.Zero(commentparsing.DerefType(t)), "")
	if f == nil {
		return reflect.Value{}, fmt.Errorf("%s has no text form", t)
	}
	if err := f.Value.Set(s); err != nil {
		return reflect.Value{}, err
	}
	return get(), nil
}
