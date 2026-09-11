package flags

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// durationType lets bindLeafFlag recognize a time.Duration field (Kind() is Int64, same as any
// plain int64) and bind it as a pflag Duration ("5s" CLI syntax) instead of a raw integer.
var durationType = reflect.TypeOf(time.Duration(0))

// configDurationType is config.Duration, the non-negative duration used by config structs across
// chainlink. It's a TextMarshaler/TextUnmarshaler, so it would otherwise bind as an untyped
// string flag; special-casing it keeps the typed "duration" pflag (and pflag's own parse errors)
// while still decoding through its UnmarshalText.
var configDurationType = reflect.TypeOf(config.Duration{})

// RegisterCommandFlags binds struct fields as CLI persistent flags, Viper defaults, and env vars.
// It also wires an automatic decode step (config file + flags + env + registered profiles) into
// target, running before cmd's own PersistentPreRunE (if any).
//
// It's safe to call this (and/or RegisterSubcommandFlags) more than once for the same cmd with
// different targets - e.g. several independent dependencies each registering their own config
// struct on a shared root command - each target is decoded, profile-defaulted, and validated
// independently.
//
// Fields are validated with go-playground/validator `validate` tags after decoding, so a rule
// sees the value whatever supplied it (flag, env, config file, or profile). Mutually exclusive or
// conditionally required *sections* must be pointer fields, so that "not configured" is
// representable:
//
//	Local *LocalConfig `toml:"local" validate:"required_without=Proxy,excluded_with=Proxy"`
//	Proxy *ProxyConfig `toml:"proxy" validate:"required_without=Local,excluded_with=Local"`
//
// Registering a non-pointer struct with such a rule is an error, since a value struct is never
// absent and the rule could not work.
//
// A list field is bound as a repeatable flag: --tags a,b and --tags a --tags b are the same
// value, and a list of lists takes one inner list per occurrence (--urls a,b --urls c) or the
// whole value at once (--urls '[a,b],[c]'), which is also how an env var writes it. A list of
// structs has no such form, so it is bound one flag per field of the element struct, each holding
// a list that lines up by position:
//
//	Nodes []Node   =>   --nodes.name a,b --nodes.url ua,ub
//
// and a config file still writes it row-wise as an array of tables, which merges with those flags
// per field. Nesting adds a list level rather than a new kind of key, so a list inside a list of
// structs is a list of lists. See nested.go and list.go.
//
// A field with no text form - a map of structs, an interface, a list of either - gets no flag and
// can only be set from a config file.
//
// opts controls the tag conventions, decoding and env prefixes; see DefaultTOMLOptions.
func RegisterCommandFlags(cmd *cobra.Command, target any, opts Options) error {
	meta := getOrCreateMeta(cmd)
	entry := &targetEntry{
		namespace:  opts.Namespace,
		flagPrefix: opts.Namespace,
		prefixes:   opts.Prefixes,
		target:     target,
		opts:       opts,
		v:          meta.v,
	}
	meta.entries = append(meta.entries, entry)

	if err := bindStruct(cmd, entry, false); err != nil {
		return err
	}

	wireDecodeHook(cmd, meta, true)
	return nil
}

// RegisterSubcommandFlags registers local flags on a subcommand, inheriting the root command's
// env prefixes when opts sets none. It also wires an automatic decode step for target, running
// before cmd's own PreRunE (if any). See RegisterCommandFlags for the
// multiple-targets-per-command note.
//
// namespace roots the keys and env vars, as it does on a root command, but by default not the
// flag names: a subcommand's settings are usually namespaced by the subcommand itself, so "sub"
// gives the key sub.retries and the flag --retries, typed as `sub --retries`. Set opts.Namespace
// as well to prefix the flags too, for a config namespaced by whatever owns it rather than by the
// command it happens to hang off - two dependencies registered on one subcommand need that, or
// their same-named settings would collide in a flag set that neither of them names.
func RegisterSubcommandFlags(cmd *cobra.Command, namespace string, target any, opts Options) error {
	meta := getOrCreateMeta(cmd)
	// With no prefixes of its own, the entry inherits the root command's at decode time, since
	// cmd usually hasn't been attached to its parent yet (see effectivePrefixes).
	entry := &targetEntry{
		namespace:  namespace,
		flagPrefix: opts.Namespace,
		prefixes:   opts.Prefixes,
		target:     target,
		opts:       opts,
		v:          meta.v,
	}
	meta.entries = append(meta.entries, entry)

	if err := bindStruct(cmd, entry, true); err != nil {
		return err
	}

	wireDecodeHook(cmd, meta, false)
	return nil
}

func bindStruct(cmd *cobra.Command, entry *targetEntry, isSubcommand bool) error {
	if entry.target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return walkStruct(entry.target, entry.opts, structVisitor{
		branch: func(m fieldMeta) (bool, error) {
			if err := entry.opts.checkSquashIsStructOnly(m); err != nil {
				return false, err
			}
			if m.isStructList {
				// Only the outermost list gets a key of its own. An inner one lives inside its
				// parent's rows, where no viper key reaches it - a config file's [[chains.nodes]]
				// is part of the value of "chains", not of a "chains.nodes".
				if m.listDepth == 0 {
					registerStructListKey(entry, m)
				}
				return false, nil
			}
			if err := checkExclusiveStructIsPointer(m); err != nil {
				return false, err
			}
			return false, entry.opts.checkEmbeddedIsUnnamed(m)
		},
		leaf: func(m fieldMeta) error {
			if err := entry.opts.checkSquashIsStructOnly(m); err != nil {
				return err
			}
			bindLeafFlag(cmd, entry, isSubcommand, m)
			return nil
		},
	})
}

// crossFieldRules are the validator rules whose outcome depends on a sibling field, and which
// therefore need a field to be distinguishably "absent" rather than merely zero.
var crossFieldRules = map[string]bool{
	"required_with": true, "required_with_all": true,
	"required_without": true, "required_without_all": true,
	"excluded_with": true, "excluded_with_all": true,
	"excluded_without": true, "excluded_without_all": true,
}

func crossFieldRuleNames(field reflect.StructField) []string {
	var names []string
	for _, rule := range strings.Split(field.Tag.Get("validate"), ",") {
		if name, _, hasArgs := strings.Cut(rule, "="); hasArgs && crossFieldRules[name] {
			names = append(names, name)
		}
	}
	return names
}

// checkExclusiveStructIsPointer rejects a non-pointer nested struct carrying a cross-field rule.
// Such a field can never be absent - a zero struct is still a struct - so the rule can't
// distinguish "not configured" from "configured to zero", and the validator descends into it
// regardless and reports its inner `required` fields for a section the user never asked for. A
// pointer makes absence representable (nil), which is what these rules need.
func checkExclusiveStructIsPointer(m fieldMeta) error {
	if m.field.Type.Kind() == reflect.Pointer || m.isTextUnmarshaler {
		return nil
	}
	rules := crossFieldRuleNames(m.field)
	if len(rules) == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s on a nested struct requires a pointer field (*%s); a value struct is never absent, so the rule cannot fire and %s's own required fields are reported even when the section is unused",
		m.key(), strings.Join(rules, "/"), m.elemType.Name(), m.elemType.Name())
}

// entryKeyNames returns the viper key and CLI flag name a field's config key binds under, which
// differ for a namespaced entry (see leafKey).
func entryKeyNames(entry *targetEntry, relKey string) (viperKey, flagName string) {
	viperKey = relKey
	if entry.namespace != "" {
		viperKey = entry.namespace + "." + relKey
	}
	flagKey := relKey
	if entry.flagPrefix != "" {
		flagKey = entry.flagPrefix + "." + relKey
	}
	return viperKey, flagNameFromViperKey(flagKey)
}

func flagNameFromViperKey(viperKey string) string {
	parts := strings.Split(viperKey, ".")
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, "_", "-")
	}
	return strings.Join(parts, ".")
}

// registerStructListKey records the key of a list of structs itself, so a config file's array of
// tables is read and merged with the columnar flags bound for the same field. It has no flag of
// its own - an array of tables has no command line form - and no env var either, for the same
// reason: the individual columns underneath it have both.
func registerStructListKey(entry *targetEntry, m fieldMeta) {
	viperKey, _ := entryKeyNames(entry, m.key())
	entry.keys = append(entry.keys, leafKey{
		relPath:   m.keyPath,
		viperKey:  viperKey,
		listKey:   m.keyPath,
		aggregate: true,
	})
}

func bindLeafFlag(cmd *cobra.Command, entry *targetEntry, isSubcommand bool, m fieldMeta) {
	v := entry.v
	relKey := m.key()
	viperKey, flagName := entryKeyNames(entry, relKey)
	entry.keys = append(entry.keys, leafKey{
		relPath:  m.keyPath,
		viperKey: viperKey,
		flagName: flagName,
		listKey:  m.listKeyPath,
	})
	defaultVal := m.elem.Interface()
	usageMsg := m.field.Tag.Get("usage")

	flags := cmd.Flags()
	if !isSubcommand {
		flags = cmd.PersistentFlags()
	}

	// A field inside a list of structs is bound as a list of its own values, one per element of
	// the list it lives in (and one level deeper again for every further list between) - see
	// nested.go. Nothing else about it changes, so a duration in a list is still written "5s".
	if m.listDepth > 0 {
		elemDepth, ok := listDepthOf(m.elemType)
		if !ok {
			// A map (or an interface, or a list of either) has no text form to repeat once per
			// element, so it stays config-file-only, the way an unsupported map does below.
			return
		}
		def := collectColumn(m.listRoot, m.listPath, elemDepth)
		v.SetDefault(viperKey, renderList(def))
		flags.Var(newTextListValue(def, m.listDepth+elemDepth), flagName, usageMsg)
		_ = v.BindPFlag(viperKey, flags.Lookup(flagName))
		return
	}

	switch {
	case m.elemType == configDurationType:
		d := defaultVal.(config.Duration)
		// The default is stored as its text form so that every source (flag, env, config file,
		// default) reaches the decoder as a string and goes through UnmarshalText. Handing
		// mapstructure the config.Duration struct instead would decode struct-to-struct and
		// silently drop the value, since its only field is unexported.
		v.SetDefault(viperKey, d.String())
		flags.Duration(flagName, d.Duration(), usageMsg)
	case m.isTextUnmarshaler:
		// Through MarshalText where the type has one, so the default is the text the type reads
		// back rather than its Go rendering - a config.URL printed with %v is its struct fields,
		// since its String/MarshalText are declared on the pointer receiver.
		text := textOf(m.elem)
		v.SetDefault(viperKey, text)
		flags.String(flagName, text, usageMsg)
	case m.elemType == durationType:
		v.SetDefault(viperKey, defaultVal)
		flags.Duration(flagName, time.Duration(m.elem.Int()), usageMsg)
	case m.elemType.Kind() == reflect.Slice && m.elemType.Elem().Kind() == reflect.Uint8:
		// A byte slice is text - a JSON blob, a PEM key - rather than a list of numbers, and weak
		// typing converts a string straight into one. Rendering it as a list ("[123 34]") would
		// give the flag a default that cannot be read back.
		text := string(m.elem.Bytes())
		v.SetDefault(viperKey, text)
		flags.String(flagName, text, usageMsg)
	case m.elemType.Kind() == reflect.Slice || m.elemType.Kind() == reflect.Array:
		depth, ok := listDepthOf(m.elemType)
		if !ok {
			// A list of structs is handled by the walk itself, so what is left here is a list with
			// no text form at all - of maps, or of interfaces. Same reasoning as the unsupported
			// map below: no flag is better than a broken one.
			return
		}
		v.SetDefault(viperKey, defaultVal)
		flags.Var(newTextListValue(treeOf(m.elem, depth), depth), flagName, usageMsg)
	case m.elemType.Kind() == reflect.Interface, m.elemType.Kind() == reflect.Func, m.elemType.Kind() == reflect.Chan:
		// No text form, so no flag - see the unsupported map below. The key stays registered, so a
		// config file can still supply one if the decoder knows how to read it.
		return
	case m.elemType.Kind() == reflect.Map:
		// Only a map whose keys are text values, and whose values are text values or lists of
		// them, has a flag form - "k=v;k=v" (see maps.go). Anything else, a map of structs or of
		// maps, can only be written in the config file, and binding a flag for it anyway would be
		// worse than binding none: the flag's default would be the string "map[]", which decoding
		// then rejects, and an env var set for that key would break the config file value too.
		valType := m.elemType.Elem()
		if !isTextValueType(m.elemType.Key()) || !isMapValueType(valType) {
			return
		}
		def := textMapOf(m.elem)
		v.SetDefault(viperKey, def)
		listValued := valType.Kind() == reflect.Slice || valType.Kind() == reflect.Array
		flags.Var(newTextMapValue(def, listValued), flagName, usageMsg)
	default:
		v.SetDefault(viperKey, defaultVal)
		switch m.elemType.Kind() {
		case reflect.String:
			flags.String(flagName, m.elem.String(), usageMsg)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			flags.Int64(flagName, m.elem.Int(), usageMsg)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			flags.Uint64(flagName, m.elem.Uint(), usageMsg)
		case reflect.Bool:
			flags.Bool(flagName, m.elem.Bool(), usageMsg)
		default:
			flags.String(flagName, fmt.Sprintf("%v", defaultVal), usageMsg)
		}
	}

	_ = v.BindPFlag(viperKey, flags.Lookup(flagName))
	// Env vars are bound later, by entry.bindEnv at decode time - see effectivePrefixes.
}
