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
	"github.com/smartcontractkit/chainlink-common/x/config/internal/decode"
)

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

	// fileType is goType with every map key a string, such as map[string]string for map[uint32]string. A config file
	// is decoded into it rather than goType, so decode.FromFile can parse each key as its flag would and catch two texts for
	// one key, such as 16 and 0x10. Flags and env vars are already text, so they don't use it.
	fileType reflect.Type

	// tag is the field's own, so the type a config file is decoded into reads the value as the field does.
	tag reflect.StructTag

	// entries set values inside a map's entries, such as chains.<key>.rpc; each is the one that sets a whole entry.
	entries []*entryLeaf
	each    *entryLeaf

	// required is a `required` rule on the field, which a profile's help leaves to the sources.
	required bool
}

type typedEntry[T any] struct {
	b      *Binder
	cmd    *cobra.Command
	dst    *T
	leaves []leafKey

	// ns, if set, is the first segment of every key.
	ns string

	// profiles run after the sources are decoded, given the leaves a source set.
	profiles []func(supplied map[string]sourceLayer) error

	// profileSelectors holds each registered profile's selector, so a new one overlapping its section fails Register:
	// each profile rebuilds its whole section, so two on one section would drop each other's values.
	profileSelectors []leafKey

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

var envVarReplacer = strings.NewReplacer(".", "_", "-", "_")

// EnvVarName is the env var a binder with prefix reads the config at path under,
// e.g. EnvVarName("CL", "cron", "fastestScheduleIntervalSeconds") is CL_CRON_FASTEST_SCHEDULE_INTERVAL_SECONDS.
func EnvVarName(prefix string, path ...string) string {
	segments := make([]string, len(path))
	for i, p := range path {
		segments[i] = kebab(p)
	}

	name := strings.ToUpper(envVarReplacer.Replace(strings.Join(segments, ".")))
	if prefix == "" {
		return name
	}

	return strings.TrimSuffix(strings.ToUpper(prefix), "_") + "_" + name
}

func (e *typedEntry[T]) envVars(k leafKey) []string {
	path := strings.Split(k.key, ".")
	if len(e.b.opts.Prefixes) == 0 {
		return []string{EnvVarName("", path...)}
	}

	names := make([]string, len(e.b.opts.Prefixes))
	for i, prefix := range e.b.opts.Prefixes {
		names[i] = EnvVarName(prefix, path...)
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

	if raw, ok := cc.keys.Lookup(cc.fileValues, k.fileKey); ok {
		val, err := decode.FromFile(commentparsing.DerefType(k.goType), raw)
		if err != nil {
			return nil, err
		}

		vals = append(vals, val)
	}

	return vals, nil
}

func (e *typedEntry[T]) validate() error {
	root := commentparsing.DerefType(reflect.TypeFor[T]())
	err := decode.NewValidator(e.b.opts.Markup).Struct(e.dst)
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
		errs[i] = decode.RuleError(e.b.opts.Markup, e.ns, root, fe)
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

// sourceLayer puts a map's source values over under, key by key, since maps merge by key at every source. A value that
// isn't a map has none: its source replaces it whole. Keys an env var names resolve against under, which can hold keys
// a profile added, so applying one can fail.
type sourceLayer func(under reflect.Value) (reflect.Value, error)

func (e *typedEntry[T]) decode(cc commandConfig) error {
	dst := reflect.ValueOf(e.dst).Elem()
	supplied := map[string]sourceLayer{}
	for _, k := range e.leaves {
		// A discarded value is parsed into any, then dropped.
		if decode.IsDiscard(k.goType) {
			continue
		}
		if len(k.entries) > 0 {
			layer, err := e.decodeEntries(k, cc, dst)
			if err != nil {
				return fmt.Errorf("%s: %w", k.key, err)
			}

			if layer != nil {
				supplied[k.key] = layer
			}

			continue
		}

		vals, err := e.sources(k, cc)
		if err != nil {
			return fmt.Errorf("%s: %w", k.key, err)
		}

		if len(vals) == 0 {
			continue
		}

		f := decode.AllocateField(dst, k.goPath)
		for f.Kind() == reflect.Pointer {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		if decode.MergesByKey(f.Type(), e.b.opts.Markup) {
			// The struct's own entries first, then sources lowest precedence first, so the highest wins a key.
			slices.Reverse(vals)
			t := f.Type()
			supplied[k.key] = func(under reflect.Value) (reflect.Value, error) {
				return decode.MergeMaps(t, append([]reflect.Value{under}, vals...)...), nil
			}
			v, err := supplied[k.key](f)
			if err != nil {
				return fmt.Errorf("%s: %w", k.key, err)
			}

			f.Set(v)
		} else {
			supplied[k.key] = nil
			f.Set(vals[0])
		}
	}

	for _, apply := range e.profiles {
		if err := apply(supplied); err != nil {
			return err
		}
	}

	return nil
}
