package flags

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// validate runs `validate:"..."` struct tag checks (e.g. `required`) against a fully decoded
// target, after config file/flags/env/profile defaulting has all been applied.
var validate = sync.OnceValue(func() *validator.Validate { return validator.New() })

type profileApplier func(cmd *cobra.Command) error

// targetEntry is one struct registered against a command via RegisterCommandFlags or
// RegisterSubcommandFlags. A single command can have multiple entries - e.g. several independent
// plugins each registering their own config struct on a shared root command - decoded and
// validated independently of one another.
type targetEntry struct {
	// namespace roots this entry's config keys.
	namespace string
	// flagPrefix roots this entry's flag names. It matches namespace for a root-command
	// registration, where several structs share one flag set and would otherwise collide, but is
	// empty for a subcommand, whose own name already separates it from its siblings.
	flagPrefix string
	prefixes   []string
	target     any
	opts       Options
	profiles   []profileApplier

	// v is the owning command's Viper instance (see commandMetaData.v), carried here so every
	// step that resolves one of this entry's keys reads the same one.
	v *viper.Viper

	// keys is every leaf bound for this target, recorded at registration so decoding can resolve
	// exactly this entry's keys (and no other entry's) through viper's normal precedence.
	keys []leafKey
}

// leafKey ties a target field's position within its own struct (relPath) to the viper key it was
// bound under (viperKey, which carries the entry's namespace prefix, if any) and the CLI flag
// registered for it (flagName, which does not - it's derived from the key relative to the target,
// so it can't be recomputed from viperKey for a namespaced entry).
type leafKey struct {
	relPath  []string
	viperKey string
	flagName string

	// listKey is the relative key path of the outermost list of structs this key sits inside,
	// empty for an ordinary key. Everything under it is assembled column-wise rather than written
	// straight into the settings map - see nested.go.
	listKey []string
	// aggregate marks the key of a list of structs itself. It carries no flag: it exists so a
	// config file's array of tables reaches the decoder, to be merged with the columns that flags
	// and env vars supply for the same field.
	aggregate bool
}

type commandMetaData struct {
	entries []*targetEntry

	// v is this command's own Viper instance. Each command (root or subcommand) gets one, so two
	// entries on different commands can reuse the same key/flag/env name (e.g. "foo" on both a
	// "foo" and a "bar" subcommand) without one's SetDefault/BindPFlag/BindEnv call clobbering
	// the other's - a single process-global viper.Viper cannot make that guarantee, since its
	// keyspace is flat and shared across every command in the tree.
	v *viper.Viper

	// hookWired guards against chaining a redundant decode step onto PersistentPreRunE/PreRunE
	// every time RegisterCommandFlags/RegisterSubcommandFlags is called again for this command;
	// the single wired hook always decodes every entry (see decodeAndApplyProfiles).
	hookWired bool

	// configFileOnce/configFileErr guard this command's one config-file load.
	configFileOnce sync.Once
	configFileErr  error
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
		meta = &commandMetaData{v: viper.New()}
		cmdRegistry[cmd] = meta
	}
	return meta
}

func getMeta(cmd *cobra.Command) *commandMetaData {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return cmdRegistry[cmd]
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
func (e *targetEntry) bindEnv() {
	for _, k := range e.keys {
		if k.aggregate {
			// A whole array of tables has no env var form; its columns each have one.
			continue
		}
		bindArgs := []string{k.viperKey}
		for _, prefix := range e.prefixes {
			bindArgs = append(bindArgs, envVarName(prefix, k.viperKey))
		}
		_ = e.v.BindEnv(bindArgs...)
	}
}

func envVarName(prefix, viperKey string) string {
	suffix := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(viperKey, ".", "_"), "-", "_"))
	return strings.TrimSuffix(strings.ToUpper(prefix), "_") + "_" + suffix
}

// wireDecodeHook chains a decode step in front of whatever PreRunE/PersistentPreRunE the command
// already has, so the caller's own hook (if any) observes already-populated targets. Only wires
// once per command - later calls on the same cmd are no-ops here since decodeAndApplyProfiles
// always walks every registered entry for the command.
func wireDecodeHook(cmd *cobra.Command, meta *commandMetaData, persistent bool) {
	if meta.hookWired {
		return
	}
	meta.hookWired = true

	if persistent {
		prev := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
			if err := decodeAndApplyProfiles(c, meta); err != nil {
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
		if err := decodeAndApplyProfiles(c, meta); err != nil {
			return err
		}
		if prev != nil {
			return prev(c, args)
		}
		return nil
	}
}

// decodeAndApplyProfiles loads the optional config file, then for every entry registered against
// cmd: unmarshals Viper state into entry.target, applies profiles registered for that entry, and
// validates `validate` tags. Entries are independent - one entry's decode/validation failure
// doesn't stop the others from being decoded and validated too, so e.g. two dependencies sharing
// a command each get to report their own missing-required-field errors in the same run instead of
// one hiding the other.
func decodeAndApplyProfiles(cmd *cobra.Command, meta *commandMetaData) error {
	// Cobra's own commands describe the program rather than run it, so they must work on a
	// machine that has no valid configuration - asking someone to satisfy every required setting
	// before they can read the help that tells them what those settings are would be backwards.
	if IsBuiltinCommand(cmd) {
		return nil
	}

	// Load into this command's own viper before reading any key - see commandMetaData.v.
	if err := meta.loadConfigFileOnce(cmd); err != nil {
		return err
	}

	var errs []error
	for _, entry := range meta.entries {
		if entry.target == nil {
			continue
		}

		// Resolve prefixes and bind env vars now that the command tree is fully assembled.
		entry.prefixes = entry.effectivePrefixes(cmd)
		entry.bindEnv()

		if err := decodeEntry(cmd, entry); err != nil {
			errs = append(errs, err)
			continue
		}

		if err := entry.applyProfiles(cmd); err != nil {
			errs = append(errs, err)
			continue
		}

		// Check `validate:"required"` (and any other validator tags) only now that config
		// file/flags/env/profile defaulting have all had a chance to fill fields in - a field can
		// be required yet still end up populated by a profile rather than the user directly.
		if err := validate().Struct(entry.target); err != nil {
			errs = append(errs, fmt.Errorf("invalid configuration: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (e *targetEntry) applyProfiles(cmd *cobra.Command) error {
	for _, applier := range e.profiles {
		if err := applier(cmd); err != nil {
			return err
		}
	}
	return nil
}

// decodeEntry resolves each of entry's registered leaf keys through viper (so CLI flag / env var
// / config file / default precedence applies per key), assembles them into a nested map shaped
// like entry.target, and decodes that into the target.
//
// This is deliberately not viper.Unmarshal/UnmarshalKey: Unmarshal decodes the whole tree, so a
// namespaced entry's "foo.timeout" would never line up with its plain "timeout" field, and
// UnmarshalKey is decode(viper.Get(namespace)), which returns whichever single source holds that
// subtree - it does not merge per-key flag/env overrides underneath it. Both silently leave
// fields at their compiled-in defaults rather than erroring.
//
// Only keys the user actually supplied (changed flag, config file entry, or env var) are
// included. Fields nobody set keep whatever the caller's struct was constructed with, which is
// what makes an optional nested *struct stay nil when nothing under it was provided - decoding
// its defaults would allocate it and defeat `required_without`/`excluded_with` on that field.
func decodeEntry(cmd *cobra.Command, entry *targetEntry) error {
	settings := map[string]any{}
	for _, k := range entry.keys {
		if !entry.isExplicitlySet(cmd, k.flagName, k.viperKey) {
			continue
		}
		val := entry.v.Get(k.viperKey)
		if val == nil {
			continue
		}
		if len(k.listKey) == 0 {
			setPath(settings, k.relPath, val)
			continue
		}
		// Everything under a list of structs is gathered per field rather than written straight
		// into the tree: the config file supplies whole rows while flags and env vars supply one
		// column each, and the two are only reconcilable once both have been collected. See
		// nested.go.
		c := columnsAt(settings, k.listKey)
		if k.aggregate {
			c.rows = val
			continue
		}
		setPath(c.cols, k.relPath[len(k.listKey):], val)
	}

	decoder, err := mapstructure.NewDecoder(entry.opts.decoderConfigFor(entry.target))
	if err != nil {
		return err
	}
	return decoder.Decode(settings)
}

// columnsAt returns the columns collected for the list of structs at the nested path within m,
// creating it (and any intermediate maps) on first use.
func columnsAt(m map[string]any, path []string) *columns {
	for _, p := range path[:len(path)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	last := path[len(path)-1]
	c, ok := m[last].(*columns)
	if !ok {
		c = newColumns()
		m[last] = c
	}
	return c
}

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

// loadConfigFileOnce reads the config file into meta's own viper the first time it's called for
// this command - each command's decode hook can run at most once per execution, but this guards
// against a command whose entries span more than one registration call.
func (meta *commandMetaData) loadConfigFileOnce(cmd *cobra.Command) error {
	meta.configFileOnce.Do(func() { meta.configFileErr = loadConfigFile(cmd, meta.v) })
	return meta.configFileErr
}

func loadConfigFile(cmd *cobra.Command, v *viper.Viper) error {
	configFile, _ := cmd.Flags().GetString("config")

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read specified config file %q: %w", configFile, err)
		}
		return nil
	}

	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	return nil
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

// isExplicitlySet reports whether viperKey's value came from something other than its registered
// code default: a changed CLI flag, a config file entry, or a bound env var. flagName is passed
// separately because a namespaced entry's flag ("timeout") does not match its viper key
// ("foo.timeout").
func (e *targetEntry) isExplicitlySet(cmd *cobra.Command, flagName, viperKey string) bool {
	if f := cmd.Flags().Lookup(flagName); f != nil && f.Changed {
		return true
	}

	if e.v.InConfig(viperKey) {
		return true
	}

	for _, prefix := range e.prefixes {
		if os.Getenv(envVarName(prefix, viperKey)) != "" {
			return true
		}
	}

	return false
}

// anyKeySetUnder reports whether any of the entry's keys at or below viperKey was explicitly set,
// which is what stands in for isExplicitlySet on a field bound as several keys.
func (e *targetEntry) anyKeySetUnder(cmd *cobra.Command, viperKey string) bool {
	for _, k := range e.keys {
		if k.viperKey != viperKey && !strings.HasPrefix(k.viperKey, viperKey+".") {
			continue
		}
		if e.isExplicitlySet(cmd, k.flagName, k.viperKey) {
			return true
		}
	}
	return false
}
