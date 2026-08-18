package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/config/configdoc"
)

// Options configures how a target struct is bound, decoded, and documented. Use
// DefaultTOMLOptions for the standard `toml`-tagged setup; the zero value works too, falling
// back to mapstructure's own defaults.
type Options struct {
	// Namespace roots every key of the registered struct under it, so a dependency's
	// settings sit together (e.g. Namespace "database" gives the key database.url, the flag
	// --database.url and the env var PREFIX_DATABASE_URL). Empty leaves the struct at the top
	// level. Independent structs sharing one command should each take a namespace, both to
	// group their settings and to keep same-named fields from colliding.
	Namespace string

	// Prefixes are the env var prefixes a key is bound under, tried in order - e.g. "CRE"
	// and "CL" bind chain.id to CRE_CHAIN_ID then CL_CHAIN_ID. Subcommands inherit the root
	// command's prefixes when they specify none.
	Prefixes []string

	// DecoderConfig controls how resolved values are decoded into the target. TagName,
	// SquashTagOption and Squash also determine the config key (and therefore the flag name)
	// of each field, so the flags, env vars, docs, and decoding all follow from this one
	// setting and cannot disagree. Result is filled in per target and must be left unset.
	//
	// WeaklyTypedInput and DecodeHook are always overridden by decoderConfigFor, whatever is
	// set here - env vars arrive as plain strings and pflag hands back several types (uint64,
	// duration, ...) as strings too, so every registration needs the same string->typed
	// coercion to decode at all; there's no valid registration that wants it off.
	DecoderConfig mapstructure.DecoderConfig

	// Format is the configuration file syntax the docs are written in: it renders the
	// comments, tables, fields and value literals that the generated document is assembled
	// from, and then renders the document itself. Defaults to configdoc.TOML.
	Format configdoc.Format
}

// DefaultTOMLOptions meant to be used with github.com/pelletier/go-toml/v2, used by default by viper for toml
// returns Options for structs tagged `toml:"key"`, with `,inline` marking a
// squashed (flattened) struct and embedded structs squashed automatically, and documented with
// configdoc.Generate. String->typed coercion (weak typing plus the duration/slice/TextUnmarshaler
// decode hooks) applies to every Options value, not just this one - see decoderConfigFor.
func DefaultTOMLOptions(prefixes ...string) Options {
	return Options{
		Prefixes: prefixes,
		DecoderConfig: mapstructure.DecoderConfig{
			TagName:         "toml",
			SquashTagOption: "inline",
			// Embedded structs are flattened into the parent rather than becoming a table
			// named after their type.
			Squash: true,
		},
		Format: configdoc.TOML{},
	}
}

// tagName is the struct tag holding config keys, defaulting to mapstructure's own.
func (o Options) tagName() string {
	if o.DecoderConfig.TagName == "" {
		return "mapstructure"
	}
	return o.DecoderConfig.TagName
}

// squashOption is the tag option marking a squashed struct, defaulting to mapstructure's own.
func (o Options) squashOption() string {
	if o.DecoderConfig.SquashTagOption == "" {
		return "squash"
	}
	return o.DecoderConfig.SquashTagOption
}

// checkEmbeddedIsUnnamed rejects a name on an embedded struct that the decoder will squash.
//
// With squashing on, `toml:"foo"` on an embedded struct is dead text: mapstructure flattens the
// fields into the parent and the name is never used, so the config key, the flag, and the docs
// all ignore it. Other encoders do not agree - encoding/json would nest the same struct under
// "foo" - so the one struct would describe two different layouts depending on who read it.
// Rejecting the name up front keeps the tag honest; drop it to squash, or make the field named
// (non-embedded) to nest.
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

// hasExplicitName reports whether field's tag names it, as opposed to carrying only options
// (`toml:",inline"`) or no tag at all.
func (o Options) hasExplicitName(field reflect.StructField) bool {
	tag := field.Tag.Get(o.tagName())
	if tag == "" || tag == "-" {
		return false
	}
	return strings.Split(tag, ",")[0] != ""
}

// format is Format, or configdoc.TOML if unset.
func (o Options) format() configdoc.Format {
	if o.Format == nil {
		return configdoc.TOML{}
	}
	return o.Format
}

// stringCoercionHooks decode the string-typed values that env vars always are (and that pflag
// hands back for several flag kinds too - duration, string slice) into their target's actual
// type: "5s" into a time.Duration/config.Duration, "a,b" into a []string the way pflag's own
// comma-separated StringSlice parsing would, and any TextUnmarshaler from its string form.
var stringCoercionHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	stringToTextSliceHookFunc(),
	stringToMapHookFunc(),
	mapstructure.TextUnmarshallerHookFunc(),
)

// stringToTextSliceHookFunc splits a comma-separated string into the elements of a list, the
// way pflag's own StringSlice parses one.
//
// mapstructure's StringToSliceHookFunc is not used because it only splits into []string
// exactly (`t != reflect.SliceOf(f)`), so a []int or a []config.Duration - including one
// inside a map, which is the whole point of "primary=a,b" - would arrive as a single element
// holding the unsplit string. The elements stay strings here; the hooks after this one and
// WeaklyTypedInput convert each to the list's real element type.
func stringToTextSliceHookFunc() mapstructure.DecodeHookFuncType {
	return func(from, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if k := to.Kind(); k != reflect.Slice && k != reflect.Array {
			return data, nil
		}
		// A []byte is text rather than a list of numbers - weak typing converts a string
		// straight into one - and a slice that unmarshals itself from text owns its own form.
		if to.Elem().Kind() == reflect.Uint8 || implementsTextUnmarshaler(to) || !isTextValueType(to.Elem()) {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}
		return strings.Split(raw, mapListSep), nil
	}
}

// stringToMapHookFunc parses the "k=v;k=v" text form of a map (see mapflag.go) into its
// entries. Both a flag and an env var deliver a map as one string - viper hands the flag's
// value straight through, since a map flag deliberately isn't one of the flag types viper
// parses itself - so this is where every source that isn't a config file becomes a map.
//
// The entries stay strings here; the hooks after this one and WeaklyTypedInput convert each to
// the map's actual key and value types, and StringToSliceHookFunc splits a list value on its
// commas, so map[string]int, map[string][]string and map[string]config.Duration all arrive
// filled in just as they would from a config file.
//
// Runs before TextUnmarshallerHookFunc but defers to it for a map type that unmarshals itself
// from text, which owns its own string form.
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
