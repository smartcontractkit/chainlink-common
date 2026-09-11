package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// Options configures how a target struct is bound, decoded, and named.
//
// It is this package's extension point: further behaviour is added as a field here rather than
// as another Register* entry point, so existing callers keep compiling. The zero value works,
// falling back to mapstructure's own conventions; DefaultTOMLOptions is the standard
// `toml`-tagged setup.
type Options struct {
	// Namespace roots every key of the registered struct under it, so a dependency's settings
	// sit together (e.g. "database" gives the key database.url, the flag --database.url and the
	// env var PREFIX_DATABASE_URL). Empty leaves the struct at the top level. Independent
	// structs sharing one command should each take a namespace, or their same-named fields
	// collide in a flag set that neither of them names.
	Namespace string

	// Prefixes are the env var prefixes a key is bound under, tried in order - e.g. "CRE" and
	// "CL" bind chain.id to CRE_CHAIN_ID then CL_CHAIN_ID. Subcommands inherit the root
	// command's prefixes when they specify none.
	Prefixes []string

	// DecoderConfig controls how resolved values are decoded into the target. TagName,
	// SquashTagOption and Squash also determine the config key (and therefore the flag name) of
	// each field, so the flags, env vars and decoding all follow from this one setting and
	// cannot disagree. Result is filled in per target and must be left unset.
	//
	// WeaklyTypedInput and DecodeHook are always overridden by decoderConfigFor, whatever is set
	// here - env vars arrive as plain strings and pflag hands back several types (uint64,
	// duration, ...) as strings too, so every registration needs the same string->typed coercion
	// to decode at all; there's no valid registration that wants it off.
	DecoderConfig mapstructure.DecoderConfig
}

// DefaultTOMLOptions returns Options for structs tagged `toml:"key"`, with `,inline` marking a
// squashed (flattened) struct and embedded structs squashed automatically. It matches
// github.com/pelletier/go-toml/v2, which viper uses for TOML. String->typed coercion applies to
// every Options value, not just this one - see decoderConfigFor.
func DefaultTOMLOptions(prefixes ...string) Options {
	return Options{
		Prefixes: prefixes,
		DecoderConfig: mapstructure.DecoderConfig{
			TagName:         "toml",
			SquashTagOption: "inline",
			Squash:          true,
		},
	}
}

func (o Options) tagName() string {
	if o.DecoderConfig.TagName == "" {
		return "mapstructure"
	}
	return o.DecoderConfig.TagName
}

func (o Options) squashOption() string {
	if o.DecoderConfig.SquashTagOption == "" {
		return "squash"
	}
	return o.DecoderConfig.SquashTagOption
}

// checkEmbeddedIsUnnamed rejects a name on an embedded struct that the decoder will squash.
//
// With squashing on, `toml:"foo"` on an embedded struct is dead text: mapstructure flattens the
// fields into the parent and the name is never used, so the config key and the flag both ignore
// it. Other encoders do not agree - encoding/json would nest the same struct under "foo" - so
// the one struct would describe two different layouts depending on who read it. Drop the name to
// squash, or make the field named (non-embedded) to nest.
func (o Options) checkEmbeddedIsUnnamed(m fieldMeta) error {
	if !o.DecoderConfig.Squash || !m.field.Anonymous || m.field.Type.Kind() != reflect.Struct {
		return nil
	}
	if !o.hasExplicitName(m.field) {
		return nil
	}
	return fmt.Errorf("%s: embedded struct %s must not be named by its %q tag while DecoderConfig.Squash is set; it is squashed into the parent, so the name is silently ignored here but would nest the struct under other encoders",
		m.field.Name, m.elemType.Name(), o.tagName())
}

// checkSquashIsStructOnly rejects the squash tag option on a field that is not a struct.
//
// Squashing means "contribute your fields to the parent", which only a struct can do:
// mapstructure rejects the option on anything else, and the walk that names the flags has to drop
// the field's own key segment to stay in step with it - so several sibling scalars all marked
// squashed collapse onto one flag name. TOML's own `,inline` is a different instruction ("write
// this value on one line") that a config struct may reasonably carry on a scalar, and with
// SquashTagOption "inline" the two are spelled the same; saying so here beats a duplicate flag or
// a decode failure later.
func (o Options) checkSquashIsStructOnly(m fieldMeta) error {
	if isSquashableType(m.field.Type) {
		return nil
	}
	tag := m.field.Tag.Get(o.tagName())
	for _, p := range strings.Split(tag, ",")[1:] {
		if p != o.squashOption() {
			continue
		}
		return fmt.Errorf("%s: the %q tag option flattens a struct into its parent and cannot be used on a %s; remove it from %s",
			m.key(), o.squashOption(), m.field.Type, m.field.Name)
	}
	return nil
}

// hasExplicitName reports whether field's tag names it, as opposed to carrying only options
// (`toml:",inline"`) or no tag at all.
func (o Options) hasExplicitName(field reflect.StructField) bool {
	tag := field.Tag.Get(o.tagName())
	if tag == "" || tag == "-" {
		return false
	}
	return strings.Split(tag, ",")[0] != ""
}

// stringCoercionHooks decode the string-typed values that env vars always are (and that pflag
// hands back for several flag kinds too) into their target's actual type: "5s" into a
// time.Duration/config.Duration, "a,b" into a []string, and any TextUnmarshaler from its string
// form.
var stringCoercionHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	columnsToRowsHookFunc(),
	stringToTextSliceHookFunc(),
	stringToMapHookFunc(),
	mapstructure.TextUnmarshallerHookFunc(),
)

// stringToTextSliceHookFunc splits the text form of a list (see list.go) into its elements:
// "a,b" or the bracketed "[a,b]" a list flag writes, and, one level at a time, the "[a,b],[c]" of
// a list of lists. Both a flag and an env var deliver a list as one string - a list flag is
// deliberately not one of the flag types viper parses itself - so this is where every source that
// isn't a config file becomes a list.
//
// mapstructure's StringToSliceHookFunc is not used because it only splits into []string exactly,
// so a []int or a []config.Duration - including one inside a map, which is the whole point of
// "primary=a,b" - would arrive as a single element holding the unsplit string. The elements stay
// strings here; this hook fires again for each one that is itself a list, and the hooks after it
// plus WeaklyTypedInput convert the rest to the list's real element type.
func stringToTextSliceHookFunc() mapstructure.DecodeHookFuncType {
	return func(from, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if k := to.Kind(); k != reflect.Slice && k != reflect.Array {
			return data, nil
		}
		// A []byte is text rather than a list of numbers - weak typing converts a string straight
		// into one - and a slice that unmarshals itself from text owns its own form.
		if to.Elem().Kind() == reflect.Uint8 || implementsTextUnmarshaler(to) {
			return data, nil
		}
		if _, ok := listDepthOf(to.Elem()); !ok {
			return data, nil
		}

		raw := strings.TrimSpace(data.(string))
		inner, _ := trimListBrackets(raw)
		if strings.TrimSpace(inner) == "" {
			return []string{}, nil
		}

		parts := splitList(inner)
		// An element that is itself a list keeps its brackets, for the next pass; only the text at
		// the bottom is unquoted, which is what lets one hold a comma or a bracket.
		leaf := isTextLeafType(to.Elem())
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if leaf {
				p = unquoteListElem(p)
			}
			parts[i] = p
		}
		return parts, nil
	}
}

// stringToMapHookFunc parses the "k=v;k=v" text form of a map (see maps.go) into its entries.
// Both a flag and an env var deliver a map as one string - viper hands the flag's value straight
// through, since a map flag deliberately isn't one of the flag types viper parses itself - so
// this is where every source that isn't a config file becomes a map.
//
// The entries stay strings here; the hooks after this one and WeaklyTypedInput convert each to
// the map's actual key and value types. It runs before TextUnmarshallerHookFunc but defers to it
// for a map type that unmarshals itself from text, which owns its own string form.
func stringToMapHookFunc() mapstructure.DecodeHookFuncType {
	return func(from, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String || to.Kind() != reflect.Map || implementsTextUnmarshaler(to) {
			return data, nil
		}
		valKind := to.Elem().Kind()
		return parseTextMap(data.(string), valKind == reflect.Slice || valKind == reflect.Array)
	}
}

// decoderConfigFor returns a copy of o.DecoderConfig aimed at target, with weak-typed string
// coercion forced on regardless of what o.DecoderConfig set - see the field doc on
// Options.DecoderConfig for why this isn't a per-Options choice.
func (o Options) decoderConfigFor(target any) *mapstructure.DecoderConfig {
	dc := o.DecoderConfig
	dc.Result = target
	if dc.MatchName == nil {
		dc.MatchName = matchKeyToFieldName
	}

	dc.WeaklyTypedInput = true
	if dc.DecodeHook != nil {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(stringCoercionHooks, dc.DecodeHook)
	} else {
		dc.DecodeHook = stringCoercionHooks
	}

	return &dc
}

// matchKeyToFieldName matches a config key against a field name (or tag) ignoring case and word
// separators, so an untagged field is reached by the key it was bound under: tagKey names it
// "finality-tag-enabled", and mapstructure's own case-insensitive comparison would not see that
// as FinalityTagEnabled. Only consulted after mapstructure's exact lookup fails, so an explicit
// tag still matches itself.
func matchKeyToFieldName(mapKey, fieldName string) bool {
	return strings.EqualFold(stripSeparators(mapKey), stripSeparators(fieldName))
}

func stripSeparators(s string) string {
	return strings.NewReplacer("-", "", "_", "").Replace(s)
}
