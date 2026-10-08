package cli

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

const defaultConfigName = "config"
const ConfigFlagName = "config"

// Binder attaches config structs to the commands of one cobra command tree.
type Binder struct {
	opts Options

	entries map[*cobra.Command]targetEntry

	// rootHooks keeps each root's own PersistentPreRunE, which preRun replaces and then runs.
	rootHooks map[*cobra.Command]func(*cobra.Command, []string) error

	// entryTemplates are the --help flags for values inside map entries, such as chains.<key>.rpc; entryFlags those
	// pflag parsed, such as chains.mainnet.rpc.
	entryTemplates  map[*pflag.Flag]*entryLeaf
	entryFlags      map[*pflag.Flag]entryFlag
	addingEntryFlag bool
	normalizing     map[*cobra.Command]bool
}

// New requires opts.Markup.
//
// New sets cobra.EnableTraverseRunHooks, which must stay set: decoding runs in the root's PersistentPreRunE, which
// cobra otherwise skips for a subcommand with its own.
func New(opts Options) (*Binder, error) {
	if opts.Markup == nil {
		return nil, markup.Err
	}

	cobra.EnableTraverseRunHooks = true
	b := &Binder{
		opts:      opts,
		entries:   map[*cobra.Command]targetEntry{},
		rootHooks: map[*cobra.Command]func(*cobra.Command, []string) error{},

		entryTemplates: map[*pflag.Flag]*entryLeaf{},
		entryFlags:     map[*pflag.Flag]entryFlag{},
		normalizing:    map[*cobra.Command]bool{},
	}
	cobra.OnInitialize(b.wire)
	return b, nil
}

// Register attaches target to cmd. Before cmd or a subcommand runs, target is filled from flags, env vars, and config
// files, then its `validate` tags are checked.
//
// Scalars, durations, []byte, [encoding.TextUnmarshaler] types, and lists and maps of those get a persistent flag and
// env vars; other fields are config file only. A list of pointers or interfaces, a map of them, and a map whose keys
// have no text form get no flag for the whole value; a map of pointers can still be set an entry at a time, below.
//
// A map's entries are also set like a struct's fields, with the key in the name: --chains.mainnet.rpc and
// APP_CHAINS_MAINNET_RPC for Chains map[string]struct{ RPC string }, or --labels.env for Labels map[string]string.
// --help shows these as --chains.<key>.rpc. A key there is one segment, holding no '.' in a flag name or '_' in an env
// var's; set others in a whole map or a config file. An env var's key is the existing key that matches it ignoring
// case, or else its lower case. Register sets cmd's global normalization func to add these flags as they are
// parsed, wrapping any set before; one set after replaces it.
//
// A command runs with its ancestors' structs too, so they must not share a key, flag, or env var; siblings may. A clash
// fails every command in the tree when it is executed.
//
// Register adds --config to cmd. Files layer in order over [Options].BaseConfig, later keys winning per key.
//
//	app --config base.toml --config prod.toml
func (b *Binder) Register[T any](cmd *cobra.Command, target *T, opts ...RegisterOption[T]) error {
	if cmd.Flags().Lookup(ConfigFlagName) != nil || cmd.PersistentFlags().Lookup(ConfigFlagName) != nil {
		return fmt.Errorf("flag --%s is already defined on %s", ConfigFlagName, cmd.Name())
	}

	// StringArray, not StringSlice, so a path may contain a comma.
	cmd.PersistentFlags().StringArray(ConfigFlagName, nil,
		"path to a config file; repeat to layer files, later ones winning")

	if !b.normalizes(cmd) {
		cmd.SetGlobalNormalizationFunc(b.addsEntryFlags(cmd.GlobalNormalizationFunc()))
		b.normalizing[cmd] = true
	}

	entry := &typedEntry[T]{
		b:   b,
		cmd: cmd,
		dst: target,
	}
	if err := bindStruct(entry); err != nil {
		return err
	}

	// Options are checked before any take effect, so an errored option leaves none applied.
	installs := make([]func(), 0, len(opts))
	for _, o := range opts {
		if o.setup == nil {
			continue
		}

		install, err := o.setup(entry)
		if err != nil {
			return err
		}

		installs = append(installs, install)
	}

	for _, install := range installs {
		install()
	}

	b.entries[cmd] = entry
	return nil
}

// cobra passes a command's normalization func on to its subcommands.
func (b *Binder) normalizes(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if b.normalizing[c] {
			return true
		}
	}

	return false
}

// commandConfig merges the keys of every struct c runs with into one tree, so one strict decode of a config file can
// reject any key none of them holds. A root struct and a subcommand's
//
//	struct{ LogLevel string; Server struct{ Host string } }
//	struct{ Server struct{ Port int } }
//
// merge into this tree for the subcommand:
//
//	{children: {
//		"LogLevel": {leaf: string, index: 0},
//		"Server": {leaf: nil, index: 1, children: {
//			"Host": {leaf: string, index: 0},
//			"Port": {leaf: int, index: 1},
//		}},
//	}}
//
// whose fileValuesType, with the TOML markup for example, is:
//
//	struct {
//		F0 *string `toml:"LogLevel"`
//		F1 *struct {
//			F0 *string `toml:"Host"`
//			F1 *int    `toml:"Port"`
//		} `toml:"Server"`
//	}
func (b *Binder) commandConfig(c *cobra.Command) (commandConfig, error) {
	cc := commandConfig{keys: &keyNode{children: map[string]*keyNode{}}}
	for ; c != nil; c = c.Parent() {
		if e := b.entries[c]; e != nil {
			cc.entries = append(cc.entries, e)
		}
	}

	slices.Reverse(cc.entries)

	lang := b.opts.Markup
	// Env vars fold '.', '-' and '_' to '_', so distinct keys can still share a name.
	// Env var names are stricter than flag names, so if they don't collide, flags don't either.
	claimed := map[string]string{}
	for _, e := range cc.entries {
		for _, k := range e.keys() {
			if err := cc.keys.add(k.fileKey, k.fileType, lang); err != nil {
				return commandConfig{}, fmt.Errorf("%s: %w", e.command().Name(), err)
			}

			if k.flag == nil {
				continue
			}

			for _, name := range e.envVars(k) {
				if other, dup := claimed[name]; dup {
					return commandConfig{}, fmt.Errorf("%s: %s and %s are both %s", e.command().Name(), other, k.key, name)
				}

				claimed[name] = k.key
			}
		}
	}

	if err := entryEnvVarsClash(cc.entries, claimed); err != nil {
		return commandConfig{}, err
	}

	return cc, nil
}

// A map entry's env vars are found by matching names, so none may match a name another key uses.
func entryEnvVarsClash(entries []targetEntry, claimed map[string]string) error {
	type template struct {
		cmd, owner, key, name string
	}

	var templates []template
	for _, e := range entries {
		for _, k := range e.keys() {
			for _, leaf := range k.entries {
				for _, name := range e.envVars(leafKey{key: leaf.key}) {
					templates = append(templates, template{e.command().Name(), k.key, leaf.key, name})
				}
			}
		}
	}

	names := slices.Sorted(maps.Keys(claimed))
	for _, t := range templates {
		for _, name := range names {
			if _, ok := matchKeys(t.name, envKeyPlaceholder, "_", name); ok {
				return fmt.Errorf("%s: %s and %s could both be %s", t.cmd, claimed[name], t.key, name)
			}
		}

		for _, other := range templates {
			if name, ok := overlap(t.name, other.name, envKeyPlaceholder, "_"); ok && other.owner != t.owner {
				return fmt.Errorf("%s: %s and %s could both be %s", t.cmd, other.key, t.key, name)
			}
		}
	}

	return nil
}

// Undocumented lists, sorted, the keys of flags without help text, usually because their struct's package has no
// generated DocComments. This can be used in a test, for example, as require.Empty(t, b.Undocumented()).
func (b *Binder) Undocumented() []string {
	var keys []string
	for _, e := range b.entries {
		keys = append(keys, e.undocumented()...)
	}

	slices.Sort(keys)
	return keys
}

// wire runs from cobra.OnInitialize, at Execute, so a hook the program assigns after Register runs after the decode
// rather than replacing it. Execute runs it every time, so a root already wired is left alone.
func (b *Binder) wire() {
	for cmd := range b.entries {
		root := cmd.Root()
		if _, wired := b.rootHooks[root]; wired {
			continue
		}

		b.rootHooks[root] = root.PersistentPreRunE
		root.PersistentPreRunE = b.preRun
	}
}

func (b *Binder) preRun(c *cobra.Command, args []string) error {
	if err := b.decode(c); err != nil {
		return err
	}

	root := c.Root()
	if hook := b.rootHooks[root]; hook != nil {
		return hook(c, args)
	}

	// cobra skips PersistentPreRun when PersistentPreRunE is set,
	// but we set PersistentPreRunE for the user so run their original PersistentPreRun for them.
	if root.PersistentPreRun != nil {
		root.PersistentPreRun(c, args)
	}

	return nil
}

func (b *Binder) decode(c *cobra.Command) error {
	// Help and completion must work without valid config.
	if IsBuiltinCommand(c) {
		return nil
	}

	// The whole tree, not just c's path: Register can't see commands attached after it, and a clash should fail every
	// command, not only the ones it affects.
	for cmd := range b.entries {
		if cmd.Root() == c.Root() {
			if _, err := b.commandConfig(cmd); err != nil {
				return err
			}
		}
	}

	cc, err := b.commandConfig(c)
	if err != nil || len(cc.entries) == 0 {
		return err
	}

	cc.flags = c.Flags()
	if err = b.shadowedEntries(cc.flags); err != nil {
		return fmt.Errorf("%s: %w", c.Name(), err)
	}

	if cc.fileValues, err = b.loadConfigFiles(c, cc.keys); err != nil {
		return err
	}

	var errs []error
	for _, entry := range cc.entries {
		if err = entry.decode(cc); err != nil {
			errs = append(errs, err)
			continue
		}

		if err = entry.validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid configuration: %w", err))
		}
	}

	return errors.Join(errs...)
}

// IsBuiltinCommand reports whether cmd is or is under cobra's help or completion commands. A [Binder] skips decoding
// for them, so help works without valid config; custom PreRunE checks should too.
func IsBuiltinCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return true
		}
	}

	return false
}

// The markup decodes strictly, so a key the command's structs don't hold is an error.
func (b *Binder) loadConfigFiles(cmd *cobra.Command, keys *keyNode) (reflect.Value, error) {
	lang := b.opts.Markup
	fileValuesType := keys.fileValuesType(lang)
	merged := reflect.New(fileValuesType).Elem()
	layer := func(data []byte, name string) error {
		v := reflect.New(fileValuesType)
		if err := lang.Unmarshal(data, v.Interface()); err != nil {
			return fmt.Errorf("invalid %s: %w", name, err)
		}

		keys.overlay(merged, v.Elem(), lang)
		return nil
	}

	if len(b.opts.BaseConfig) > 0 {
		if err := layer(b.opts.BaseConfig, "Options.BaseConfig"); err != nil {
			return reflect.Value{}, err
		}
	}

	// A subcommand's own --config shadows the one Register added, and holds no config files.
	paths, err := cmd.Flags().GetStringArray(ConfigFlagName)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("--%s on %s is not the flag Register added: %w",
			ConfigFlagName, cmd.Name(), err)
	}

	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
	}

	paths = slices.DeleteFunc(paths, func(p string) bool { return p == "" })
	named := len(paths) > 0
	if !named {
		defaultPath := b.opts.DefaultConfigPath
		if defaultPath == "" {
			defaultPath = defaultConfigName + "." + lang.Extension()
		}

		paths = []string{defaultPath}
	}

	for _, path := range paths {
		data, err := os.ReadFile(path) //nolint:gosec // G304: the path is the operator's own --config
		if err != nil {
			if !named && errors.Is(err, os.ErrNotExist) {
				continue
			}

			return reflect.Value{}, fmt.Errorf("failed to read config file %q: %w", path, err)
		}

		if err = layer(data, fmt.Sprintf("config file %q", path)); err != nil {
			return reflect.Value{}, err
		}
	}

	return merged, nil
}

type commandConfig struct {
	entries    []targetEntry
	keys       *keyNode
	fileValues reflect.Value
	flags      *pflag.FlagSet
}
