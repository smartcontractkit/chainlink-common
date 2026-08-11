package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/config/configdoc"
)

// durationType lets bindLeafFlag recognize a time.Duration field (Kind() is Int64, same as
// any plain int64) and bind it as a pflag Duration ("5s" CLI syntax) instead of a raw integer.
var durationType = reflect.TypeOf(time.Duration(0))

// configDurationType is config.Duration, the non-negative duration used by config structs
// across chainlink. It's a TextMarshaler/TextUnmarshaler, so it would otherwise bind as an
// untyped string flag; special-casing it keeps the typed "duration" pflag (and pflag's own
// parse errors) while still decoding through its UnmarshalText.
var configDurationType = reflect.TypeOf(config.Duration{})

// validate runs `validate:"..."` struct tag checks (e.g. `required`) against a fully
// decoded target, after config file/flags/env/profile defaulting has all been applied.
var validate = sync.OnceValue(func() *validator.Validate { return validator.New() })

type profileApplier func(cmd *cobra.Command, target any) error

// targetEntry is one struct registered against a command via RegisterCommandFlags or
// RegisterSubcommandFlags. A single command can have multiple entries - e.g. several
// independent plugins/dependencies each registering their own config struct on a shared root
// command - decoded and validated independently of one another.
type targetEntry struct {
	// namespace roots this entry's config keys.
	namespace string
	// flagPrefix roots this entry's flag names. It matches namespace for a root-command
	// registration, where several structs share one flag set and would otherwise collide, but
	// is empty for a subcommand, whose own name already separates it from its siblings.
	flagPrefix string
	prefixes  []string
	target    any
	opts      Options
	profiles  []profileApplier

	// keys is every leaf bound for this target, recorded at registration so decoding can
	// resolve exactly this entry's keys (and no other entry's) through viper's normal
	// precedence.
	keys []leafKey
}

// leafKey ties a target field's position within its own struct (relPath) to the viper key it
// was bound under (viperKey, which carries the entry's namespace prefix, if any) and the CLI
// flag registered for it (flagName, which does not - it's derived from the key relative to the
// target, so it can't be recomputed from viperKey for a namespaced entry).
type leafKey struct {
	relPath  []string
	viperKey string
	flagName string
}

type commandMetaData struct {
	entries []*targetEntry

	// docFormat is the syntax documentation for this command's whole tree is written in,
	// taken from the Options of the registration that added the docs command (see
	// addDocsCommand).
	docFormat configdoc.Format

	// hookWired guards against chaining a redundant decode step onto PersistentPreRunE/PreRunE
	// every time RegisterCommandFlags/RegisterSubcommandFlags is called again for this command;
	// the single wired hook always decodes every entry (see decodeAndApplyProfiles).
	hookWired bool
}

var (
	registryMu  sync.RWMutex
	cmdRegistry = make(map[*cobra.Command]*commandMetaData)
)

func getOrCreateMeta(cmd *cobra.Command) *commandMetaData {
	registryMu.Lock()
	defer registryMu.Unlock()

	meta, exists := cmdRegistry[cmd]
	if !exists {
		meta = &commandMetaData{}
		cmdRegistry[cmd] = meta
	}
	return meta
}

func getMeta(cmd *cobra.Command) *commandMetaData {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return cmdRegistry[cmd]
}

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
// sees the value whatever supplied it (flag, env, config file, or profile). Mutually exclusive
// or conditionally required *sections* must be pointer fields, so that "not configured" is
// representable:
//
//	Local *LocalConfig `toml:"local" validate:"required_without=Proxy,excluded_with=Proxy"`
//	Proxy *ProxyConfig `toml:"proxy" validate:"required_without=Local,excluded_with=Local"`
//
// Registering a non-pointer struct with such a rule is an error, since a value struct is never
// absent and the rule could not work.
//
// opts controls the tag conventions, decoding, env prefixes, and doc generation; see
// DefaultTOMLOptions.
func RegisterCommandFlags(cmd *cobra.Command, target any, opts Options) error {
	meta := getOrCreateMeta(cmd)
	entry := &targetEntry{
		namespace:  opts.Namespace,
		flagPrefix: opts.Namespace,
		prefixes:   opts.Prefixes,
		target:     target,
		opts:       opts,
	}
	meta.entries = append(meta.entries, entry)

	if err := registerStructFlagsInternal(cmd, entry, false); err != nil {
		return err
	}

	wireDecodeHook(cmd, meta, true)
	addDocsCommand(cmd, opts)
	return nil
}

// RegisterSubcommandFlags registers local flags on a subcommand, inheriting the root command's
// env prefixes when opts sets none. It also wires an automatic decode step for target, running
// before cmd's own PreRunE (if any). See RegisterCommandFlags for the multiple-targets-per-command
// note.
//
// namespace roots the keys and env vars, as it does on a root command, but by default not the flag
// names: a subcommand's settings are usually namespaced by the subcommand itself, so "sub" gives
// the key sub.retries and the flag --retries, typed as `sub --retries`. Set opts.Namespace as well
// to prefix the flags too, for a config namespaced by whatever owns it rather than by the command
// it happens to hang off - two dependencies registered on one subcommand need that, or their
// same-named settings would collide in a flag set that neither of them names.
func RegisterSubcommandFlags(cmd *cobra.Command, namespace string, target any, opts Options) error {
	meta := getOrCreateMeta(cmd)
	// With no prefixes of its own, the entry inherits the root command's at decode time, since
	// cmd usually hasn't been attached to its parent yet (see effectivePrefixes).
	entry := &targetEntry{namespace: namespace, flagPrefix: opts.Namespace, prefixes: opts.Prefixes, target: target, opts: opts}
	meta.entries = append(meta.entries, entry)

	if err := registerStructFlagsInternal(cmd, entry, true); err != nil {
		return err
	}

	wireDecodeHook(cmd, meta, false)
	return nil
}

// effectivePrefixes returns the entry's own env-var prefixes, or - for a subcommand entry that
// didn't specify any - the union of those registered on cmd's root command.
//
// This is resolved when the command runs rather than when it's registered: a subcommand is
// typically registered with its config before rootCmd.AddCommand(sub) is called, so at
// registration time sub.Root() is still sub itself and the root's prefixes aren't reachable yet.
func (e *targetEntry) effectivePrefixes(cmd *cobra.Command) []string {
	if len(e.prefixes) > 0 {
		return e.prefixes
	}

	rootMeta := getMeta(cmd.Root())
	if rootMeta == nil {
		return nil
	}

	seen := make(map[string]bool)
	var prefixes []string
	for _, re := range rootMeta.entries {
		for _, p := range re.prefixes {
			if !seen[p] {
				seen[p] = true
				prefixes = append(prefixes, p)
			}
		}
	}
	return prefixes
}

// bindEnv binds each of the entry's keys to its PREFIX_UPPER_SNAKE env vars. Called at decode
// time for the same reason effectivePrefixes is resolved there.
func (e *targetEntry) bindEnv(prefixes []string) {
	if len(prefixes) == 0 {
		return
	}
	for _, k := range e.keys {
		envSuffix := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(k.viperKey, ".", "_"), "-", "_"))
		bindArgs := []string{k.viperKey}
		for _, prefix := range prefixes {
			bindArgs = append(bindArgs, strings.TrimSuffix(strings.ToUpper(prefix), "_")+"_"+envSuffix)
		}
		_ = viper.BindEnv(bindArgs...)
	}
}

// wireDecodeHook chains a decode step in front of whatever PreRunE/PersistentPreRunE the command
// already has, so the caller's own hook (if any) observes already-populated targets. Only wires
// once per command - later calls (from additional RegisterCommandFlags/RegisterSubcommandFlags
// calls on the same cmd) are no-ops here since decodeAndApplyProfiles always walks every
// registered entry for the command.
func wireDecodeHook(cmd *cobra.Command, meta *commandMetaData, persistent bool) {
	if meta.hookWired {
		return
	}
	meta.hookWired = true

	decode := func(c *cobra.Command, args []string) error {
		return decodeAndApplyProfiles(c, meta)
	}

	if persistent {
		prev := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
			if err := decode(c, args); err != nil {
				return err
			}
			if prev != nil {
				return prev(c, args)
			}
			return nil
		}
		return
	}

	prev := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		if err := decode(c, args); err != nil {
			return err
		}
		if prev != nil {
			return prev(c, args)
		}
		return nil
	}
}

// decodeAndApplyProfiles loads the optional config file, then for every entry registered
// against cmd: unmarshals Viper state into entry.target, applies profiles registered for that
// entry, and validates `validate:"required"` tags. Entries are independent - one entry's
// decode/validation failure doesn't stop the others from being decoded and validated too, so
// e.g. two dependencies sharing a command each get to report their own missing-required-field
// errors in the same run instead of one hiding the other.
func decodeAndApplyProfiles(cmd *cobra.Command, meta *commandMetaData) error {
	// Cobra's own commands describe the program rather than run it, so they must work on a
	// machine that has no valid configuration - asking someone to satisfy every required
	// setting before they can read the help that tells them what those settings are would be
	// backwards. They are generated when the command runs, too late to opt out at
	// registration the way the docs command does, so they're recognized here.
	if IsBuiltinCommand(cmd) {
		return nil
	}

	// Load into the (global) viper before reading any key. cmd here is the command actually
	// being executed, not the one this hook was registered on, so it can't be used to tell
	// "am I the root" - hence loading once per process rather than only on the root's pass.
	if err := loadConfigFileOnce(cmd); err != nil {
		return err
	}

	var errs []error
	for _, entry := range meta.entries {
		if entry.target == nil {
			continue
		}

		// Resolve prefixes and bind env vars now that the command tree is fully assembled.
		entry.prefixes = entry.effectivePrefixes(cmd)
		entry.bindEnv(entry.prefixes)

		if err := decodeEntry(cmd, entry); err != nil {
			errs = append(errs, err)
			continue
		}

		var applierErr error
		for _, applier := range entry.profiles {
			if err := applier(cmd, entry.target); err != nil {
				applierErr = err
				break
			}
		}
		if applierErr != nil {
			errs = append(errs, applierErr)
			continue
		}

		// Check `validate:"required"` (and any other validator tags) only now that config
		// file/flags/env/profile defaulting have all had a chance to fill fields in - a field
		// can be required yet still end up populated by a profile rather than the user directly.
		if err := validate().Struct(entry.target); err != nil {
			errs = append(errs, fmt.Errorf("invalid configuration: %w", err))
		}
	}

	return errors.Join(errs...)
}

// decodeEntry resolves each of entry's registered leaf keys through viper (so CLI flag / env
// var / config file / default precedence applies per key), assembles them into a nested map
// shaped like entry.target, and decodes that into the target.
//
// This is deliberately not viper.Unmarshal/UnmarshalKey: Unmarshal decodes the whole tree, so
// a namespaced entry's "foo.timeout" would never line up with its plain "timeout" field, and
// UnmarshalKey is decode(viper.Get(namespace)), which returns whichever single source holds
// that subtree - it does not merge per-key flag/env overrides underneath it. Both silently
// leave fields at their compiled-in defaults rather than erroring.
// Only keys the user actually supplied (changed flag, config file entry, or env var) are
// included. Fields nobody set keep whatever the caller's struct was constructed with, which is
// what makes an optional nested *struct stay nil when nothing under it was provided - decoding
// its defaults would allocate it and defeat `required_without`/`excluded_with` on that field.
func decodeEntry(cmd *cobra.Command, entry *targetEntry) error {
	settings := map[string]any{}
	for _, k := range entry.keys {
		if !isExplicitlySet(cmd, k.flagName, k.viperKey, entry.prefixes) {
			continue
		}
		val := viper.Get(k.viperKey)
		if val == nil {
			continue
		}
		setPath(settings, k.relPath, val)
	}

	decoder, err := mapstructure.NewDecoder(entry.opts.decoderConfigFor(entry.target))
	if err != nil {
		return err
	}
	return decoder.Decode(settings)
}

// setPath assigns val at the nested path within m, creating intermediate maps as needed.
func setPath(m map[string]any, path []string, val any) {
	for _, p := range path[:len(path)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[path[len(path)-1]] = val
}

var (
	configFileOnce = new(sync.Once)
	configFileErr  error
)

// loadConfigFileOnce reads the config file into viper the first time it's called, since viper
// is process-global and several commands' decode hooks can run in one execution.
func loadConfigFileOnce(cmd *cobra.Command) error {
	configFileOnce.Do(func() { configFileErr = loadConfigFile(cmd) })
	return configFileErr
}

// IsBuiltinCommand reports whether cmd is one of cobra's generated commands (help, completion,
// and the hidden completion callbacks), or lives under one.
//
// Decoding and validation are skipped for these automatically. Callers that chain their own
// PersistentPreRunE/PreRunE with extra checks - rules spanning several config structs, say -
// should return early on it too, or those checks will reject `help` on a machine that has no
// configuration yet.
func IsBuiltinCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return true
		}
	}
	return false
}

func loadConfigFile(cmd *cobra.Command) error {
	configFile, _ := cmd.Flags().GetString("config")

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read specified config file %q: %w", configFile, err)
		}
		return nil
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	return nil
}

// RegisterProfile attaches a profile map to a command using a selector field path (e.g. "Chain.ID" or "chain.id").
// T must match the type of a struct already registered for this command via
// RegisterCommandFlags/RegisterSubcommandFlags - if more than one entry of type T is registered
// on cmd, or none are, this returns an error (call RegisterCommandFlags/RegisterSubcommandFlags
// first, and don't register two same-typed structs on one command if you also need a profile).
func RegisterProfile[T any, K comparable](
	cmd *cobra.Command,
	selectorFieldName string,
	profiles map[K]T,
	opts Options,
) error {
	meta := getOrCreateMeta(cmd)

	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	entry, err := findEntryByTargetType(meta, targetType)
	if err != nil {
		return err
	}

	selectorPath, err := verifySelectorType(targetType, selectorFieldName, reflect.TypeOf((*K)(nil)).Elem(), opts)
	if err != nil {
		return err
	}

	applier := func(cmd *cobra.Command, target any) error {
		vTarget := reflect.ValueOf(target)
		if vTarget.Kind() == reflect.Pointer {
			vTarget = vTarget.Elem()
		}

		selectedKey := extractSelectorValue[K](vTarget, selectorPath)
		profile, exists := profiles[selectedKey]
		if !exists {
			// No profile for this selector value: nothing to default, not an error.
			return nil
		}

		// Scope the copy to the substruct owning the selector (e.g. "System" for
		// "System.Env") so this profile can't clobber unrelated branches (e.g. "Chain")
		// that just happen to be zero-valued in this profile's map entry.
		scopePath := selectorPath[:len(selectorPath)-1]
		leafName := selectorPath[len(selectorPath)-1]
		targetScope := navigateFields(vTarget, scopePath)
		profileScope := navigateFields(reflect.ValueOf(profile), scopePath)
		prefix := scopePrefix(targetType, scopePath, opts)

		// The selector's own field (e.g. Chain.ID) sits inside the scoped substruct too;
		// preserve the value that was actually used to pick this profile, since the
		// profile's map entry usually leaves it zero (the profile is keyed by that value,
		// not describing it).
		leafField := targetScope.FieldByName(leafName)
		selectedValue := reflect.ValueOf(leafField.Interface())

		applyProfileDefaults(cmd, entry, opts, prefix, targetScope, profileScope)

		leafField.Set(selectedValue)
		return nil
	}

	entry.profiles = append(entry.profiles, applier)
	enableProfileHelp(cmd, selectorPath, profiles)
	return nil
}

// findEntryByTargetType returns the single entry registered on meta whose target's type
// (dereferenced) matches targetType.
func findEntryByTargetType(meta *commandMetaData, targetType reflect.Type) (*targetEntry, error) {
	var match *targetEntry
	for _, e := range meta.entries {
		et := reflect.TypeOf(e.target)
		if et.Kind() == reflect.Pointer {
			et = et.Elem()
		}
		if et != targetType {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("multiple registered targets of type %s on this command; ambiguous profile registration", targetType)
		}
		match = e
	}
	if match == nil {
		return nil, fmt.Errorf("no registered target of type %s on this command; call RegisterCommandFlags/RegisterSubcommandFlags first", targetType)
	}
	return match, nil
}

// --- INTERNAL HELPERS ---

func verifySelectorType(tType reflect.Type, selectorFieldName string, kType reflect.Type, opts Options) ([]string, error) {
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target type T must be a struct")
	}

	path, fType, found := findFieldByTagOrName(tType, selectorFieldName, opts)
	if !found {
		return nil, fmt.Errorf("field or tag %q not found in struct %s", selectorFieldName, tType.Name())
	}

	if fType != kType {
		return nil, fmt.Errorf("type mismatch for %q in %s: field is %s, profile key K is %s", selectorFieldName, tType.Name(), fType, kType)
	}

	return path, nil
}

func findFieldByTagOrName(t reflect.Type, name string, opts Options) ([]string, reflect.Type, bool) {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		curr := t
		var fullPath []string

		for _, part := range parts {
			if curr.Kind() == reflect.Pointer {
				curr = curr.Elem()
			}
			if curr.Kind() != reflect.Struct {
				return nil, nil, false
			}

			subPath, fType, ok := findSingleField(curr, part, opts)
			if !ok {
				return nil, nil, false
			}
			fullPath = append(fullPath, subPath...)
			curr = fType
		}
		return fullPath, curr, true
	}

	return findSingleField(t, name, opts)
}

func findSingleField(t reflect.Type, name string, opts Options) ([]string, reflect.Type, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		keyName, _ := opts.tagKey(field)

		if strings.EqualFold(field.Name, name) || strings.EqualFold(keyName, name) {
			return []string{field.Name}, field.Type, true
		}

		elemType := field.Type
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
		}
		if elemType.Kind() == reflect.Struct {
			if subPath, fType, ok := findSingleField(elemType, name, opts); ok {
				return append([]string{field.Name}, subPath...), fType, true
			}
		}
	}
	return nil, nil, false
}

func extractSelectorValue[K comparable](v reflect.Value, path []string) K {
	curr := v
	for _, p := range path {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		curr = curr.FieldByName(p)
	}
	return curr.Interface().(K)
}

// navigateFields walks v down through the named fields in path (dereferencing pointers
// as needed), returning the resulting nested field value.
func navigateFields(v reflect.Value, path []string) reflect.Value {
	curr := v
	for _, p := range path {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		curr = curr.FieldByName(p)
	}
	return curr
}

// scopePrefix computes the dotted toml/mapstructure key prefix (e.g. "system") that
// corresponds to the given Go field-name path (e.g. ["System"]) starting from struct type t.
func scopePrefix(t reflect.Type, path []string, opts Options) string {
	var parts []string
	curr := t
	for _, name := range path {
		if curr.Kind() == reflect.Pointer {
			curr = curr.Elem()
		}
		field, ok := curr.FieldByName(name)
		if !ok {
			break
		}
		key, _ := opts.tagKey(field)
		parts = append(parts, key)
		curr = field.Type
	}
	return strings.Join(parts, ".")
}

func applyProfileDefaults(cmd *cobra.Command, entry *targetEntry, opts Options, prefix string, tVal, pVal reflect.Value) {
	if pVal.Kind() == reflect.Pointer {
		pVal = pVal.Elem()
	}
	copyDefaultsRecursive(cmd, entry, opts, prefix, tVal, pVal)
}

func copyDefaultsRecursive(cmd *cobra.Command, entry *targetEntry, opts Options, prefix string, tVal, pVal reflect.Value) {
	namespace, prefixes := entry.namespace, entry.prefixes
	tType := tVal.Type()
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		key, _ := opts.tagKey(field)

		relKey := key
		if prefix != "" {
			relKey = prefix + "." + key
		}

		viperKey := relKey
		if namespace != "" {
			viperKey = namespace + "." + relKey
		}

		targetField := tVal.Field(i)
		profileField := pVal.Field(i)

		if targetField.Kind() == reflect.Pointer {
			if targetField.IsNil() && !profileField.IsNil() {
				targetField.Set(reflect.New(targetField.Type().Elem()))
			}
			if !targetField.IsNil() && !profileField.IsNil() {
				targetField = targetField.Elem()
				profileField = profileField.Elem()
			} else {
				continue
			}
		}

		if targetField.Kind() == reflect.Struct {
			copyDefaultsRecursive(cmd, entry, opts, relKey, targetField, profileField)
			continue
		}

			if !isExplicitlySet(cmd, flagNameFromViperKey(relKey), viperKey, prefixes) {
			targetField.Set(profileField)
		}
	}
}

// isExplicitlySet reports whether viperKey's value came from something other than its
// registered code default: a changed CLI flag, a config file entry, or a bound env var.
// flagName is passed separately because a namespaced entry's flag ("timeout") does not match
// its viper key ("foo.timeout").
func isExplicitlySet(cmd *cobra.Command, flagName, viperKey string, prefixes []string) bool {
	if f := cmd.Flags().Lookup(flagName); f != nil && f.Changed {
		return true
	}

	if viper.InConfig(viperKey) {
		return true
	}

	envSuffix := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(viperKey, ".", "_"), "-", "_"))
	for _, prefix := range prefixes {
		cleanPrefix := strings.TrimSuffix(strings.ToUpper(prefix), "_")
		if os.Getenv(cleanPrefix+"_"+envSuffix) != "" {
			return true
		}
	}

	return false
}

func flagNameFromViperKey(viperKey string) string {
	parts := strings.Split(viperKey, ".")
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, "_", "-")
	}
	return strings.Join(parts, ".")
}

func enableProfileHelp[T any, K comparable](cmd *cobra.Command, selectorPath []string, profiles map[K]T) {
	selectorFlagName := strings.ReplaceAll(strings.ToLower(selectorPath[len(selectorPath)-1]), "_", "-")
	existingHelp := cmd.HelpFunc()

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if existingHelp != nil {
			existingHelp(c, args)
		} else {
			fmt.Printf("Usage of %s:\n\n", c.CommandPath())
			c.Flags().VisitAll(func(f *pflag.Flag) {
				defaultMsg := ""
				if f.DefValue != "" {
					defaultMsg = fmt.Sprintf(" (default %q)", f.DefValue)
				}
				fmt.Printf("  --%-35s %s%s\n", f.Name, f.Usage, defaultMsg)
			})
		}

		fmt.Printf("\nAVAILABLE PROFILES FOR --%s:\n", selectorFlagName)
		for k := range profiles {
			fmt.Printf("  - %v\n", k)
		}
	})
}

func registerStructFlagsInternal(cmd *cobra.Command, entry *targetEntry, isSubcommand bool) error {
	if entry.target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return walkStruct(entry.target, entry.opts, structVisitor{
		branch: func(m fieldMeta) (bool, error) {
			if err := checkExclusiveStructIsPointer(m); err != nil {
				return false, err
			}
			return false, entry.opts.checkEmbeddedIsUnnamed(m)
		},
		leaf: func(m fieldMeta) error {
			bindLeafFlag(cmd, entry, isSubcommand, m)
			return nil
		},
	})
}

// checkExclusiveStructIsPointer rejects a non-pointer nested struct carrying a cross-field
// rule. Such a field can never be absent - a zero struct is still a struct - so the rule can't
// distinguish "not configured" from "configured to zero", and the validator descends into it
// regardless and reports its inner `required` fields for a section the user never asked for.
// A pointer makes absence representable (nil), which is what these rules need.
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

func bindLeafFlag(cmd *cobra.Command, entry *targetEntry, isSubcommand bool, m fieldMeta) {
	namespace := entry.namespace

	relKey := m.key()
	viperKey := relKey
	if namespace != "" {
		viperKey = namespace + "." + relKey
	}

	flagKey := relKey
	if entry.flagPrefix != "" {
		flagKey = entry.flagPrefix + "." + relKey
	}
	flagName := flagNameFromViperKey(flagKey)
	entry.keys = append(entry.keys, leafKey{relPath: m.keyPath, viperKey: viperKey, flagName: flagName})
	defaultVal := m.elem.Interface()
	usageMsg := m.field.Tag.Get("usage")

	flags := cmd.Flags()
	if !isSubcommand {
		flags = cmd.PersistentFlags()
	}

	switch {
	case m.elemType == configDurationType:
		d := defaultVal.(config.Duration)
		// Default is stored as its text form so that every source (flag, env, config file,
		// default) reaches the decoder as a string and goes through UnmarshalText. Handing
		// mapstructure the config.Duration struct instead would decode struct-to-struct and
		// silently drop the value, since its only field is unexported.
		viper.SetDefault(viperKey, d.String())
		flags.Duration(flagName, d.Duration(), usageMsg)
	case m.isTextUnmarshaler:
		text := fmt.Sprintf("%v", defaultVal)
		viper.SetDefault(viperKey, text)
		flags.String(flagName, text, usageMsg)
	case m.elemType == durationType:
		viper.SetDefault(viperKey, defaultVal)
		flags.Duration(flagName, time.Duration(m.elem.Int()), usageMsg)
	case m.elemType.Kind() == reflect.Slice && m.elemType.Elem().Kind() == reflect.String:
		var def []string
		if !m.elem.IsNil() {
			def = m.elem.Interface().([]string)
		}
		viper.SetDefault(viperKey, def)
		flags.StringSlice(flagName, def, usageMsg)
	default:
		viper.SetDefault(viperKey, defaultVal)
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

	_ = viper.BindPFlag(viperKey, flags.Lookup(flagName))
	// Env vars are bound later, by entry.bindEnv at decode time - see effectivePrefixes.
}
