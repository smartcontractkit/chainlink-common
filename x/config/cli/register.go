package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

// durationType tells time.Duration apart from int64, which shares its Kind.
var durationType = reflect.TypeFor[time.Duration]()

func bindLeafFlag(entry *targetEntry, m fieldMeta) error {
	flags := entry.cmd.PersistentFlags()
	leaf := leafKey{
		key:      m.key,
		fileKey:  m.fileKey,
		flagName: strings.ReplaceAll(m.key, "_", "-"),
		goPath:   m.goPath,
		goType:   m.field.Type,
	}
	// pflag panics on a redefinition.
	if flags.Lookup(leaf.flagName) != nil {
		return fmt.Errorf("%s: flag --%s is already defined on %s", m.key, leaf.flagName, entry.cmd.Name())
	}

	docs, _ := commentparsing.Lookup(m.owner)
	doc := strings.Join(strings.Fields(docs[m.field.Name].Comment), " ")
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

func isRequired(field reflect.StructField) bool {
	for rule := range strings.SplitSeq(field.Tag.Get("validate"), ",") {
		switch rule {
		case "required":
			return true
		case "dive":
			// The rules after it are a list's or map's elements' (and, between keys and endkeys, a map's keys'), not
			// the field's own: `dive,required` requires every element, not the list.
			return false
		}
	}

	return false
}

// A nil flag means the type has no text form, so the field is config file only.
func newFlag(name string, t reflect.Type, def reflect.Value, usage string) (*pflag.Flag, func() reflect.Value) {
	t = commentparsing.DerefType(t)
	var value pflag.Value
	var get func() reflect.Value
	switch {
	case canText(t):
		v := &textValue{value: reflect.New(t).Elem()}
		if def.IsValid() && def.Type() == t {
			v.value.Set(def)
		}

		value, get = v, func() reflect.Value { return v.value }
	case t.Kind() == reflect.Slice && canText(t.Elem()):
		l := &textListValue{list: reflect.MakeSlice(t, 0, 0), elems: textListOf(def)}
		value, get = l, func() reflect.Value { return l.list }
	case t.Kind() == reflect.Map && canText(t.Key()) && canText(t.Elem()):
		m := &textMapValue{m: reflect.MakeMap(t), entries: textMapOf(def), seen: map[any]string{}}
		value, get = m, func() reflect.Value { return m.m }
	default:
		return nil, nil
	}

	f := &pflag.Flag{Name: name, Usage: usage, Value: value, DefValue: value.String()}
	// So --flag alone sets it, as with pflag's own bool flags.
	if t.Kind() == reflect.Bool && !readsText(t) {
		f.NoOptDefVal = "true"
	}

	return f, get
}

// Parses into the field's own type, so a named type such as type Port int needs no conversion.
type textValue struct{ value reflect.Value }

func (v *textValue) Set(s string) error {
	p := reflect.New(v.value.Type())
	if err := setText(p.Elem(), s); err != nil {
		return err
	}

	v.value.Set(p.Elem())
	return nil
}

// Blank when zero, so it doesn't look set; otherwise the type's marshaller, so it parses back and
// pkg/config.SecretString stays redacted.
func (v *textValue) String() string {
	if v.value.IsZero() {
		return ""
	}

	return textOf(v.value)
}

// Named as pflag names its own flags, for consistent --help.
func (v *textValue) Type() string {
	switch t := v.value.Type(); {
	case readsText(t) || t.Kind() == reflect.Slice:
		return "string"
	case t == durationType:
		return "duration"
	default:
		return t.Kind().String()
	}
}

// Parsed as pflag would: any Go literal base, and sized, so 256 overflows a uint8. []byte is text (JSON, PEM), not a
// list of numbers.
func setText(dst reflect.Value, s string) error {
	t := dst.Type()
	switch {
	case readsText(t):
		return dst.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
	case t == durationType:
		d, err := time.ParseDuration(s)
		dst.SetInt(int64(d))
		return err
	}

	var err error
	switch t.Kind() {
	case reflect.String:
		dst.SetString(s)
	case reflect.Slice:
		dst.SetBytes([]byte(s))
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(s)
		dst.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		n, err = strconv.ParseInt(s, 0, t.Bits())
		dst.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var n uint64
		n, err = strconv.ParseUint(s, 0, t.Bits())
		dst.SetUint(n)
	case reflect.Float32, reflect.Float64:
		var n float64
		n, err = strconv.ParseFloat(s, t.Bits())
		dst.SetFloat(n)
	default:
		err = fmt.Errorf("%s has no text form", t)
	}

	return err
}

// Parsed through a flag, so env vars and config file map keys accept exactly what the flag does.
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

// A type's own marshaller, so the default parses back and secrets stay redacted.
func textOf(v reflect.Value) string {
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		return ""
	}

	v = reflect.Indirect(v)
	if text, ok := marshalText(v); ok {
		return text
	}

	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		return string(v.Bytes())
	}

	return fmt.Sprint(v.Interface())
}

// Copied somewhere addressable first, so a pointer-receiver MarshalText is reachable.
func marshalText(v reflect.Value) (string, bool) {
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	m, ok := p.Interface().(encoding.TextMarshaler)
	if !ok {
		return "", false
	}

	text, err := m.MarshalText()
	return string(text), err == nil
}
