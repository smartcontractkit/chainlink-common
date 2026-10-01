package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

const defaultConfigName = "config"
const ConfigFlagName = "config"

// Binder attaches config structs to the commands of one cobra command tree.
type Binder struct {
	opts Options

	entries map[*cobra.Command][]*targetEntry

	// rootHooks keeps each root's own PersistentPreRunE, which preRun replaces and then runs.
	rootHooks map[*cobra.Command]func(*cobra.Command, []string) error
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
		entries:   map[*cobra.Command][]*targetEntry{},
		rootHooks: map[*cobra.Command]func(*cobra.Command, []string) error{},
	}
	cobra.OnInitialize(b.wire)
	return b, nil
}

// Register attaches target to cmd. Before cmd or a subcommand runs, target is filled from flags, env vars, and config
// files, then its `validate` tags are checked.
//
// Scalars, durations, []byte, [encoding.TextUnmarshaler] types, and lists and maps of those get a persistent flag and
// env vars; other fields are config file only.
//
// A command runs with its ancestors' structs too, so they must not share a key, flag, or env var; siblings may. A clash
// fails every command in the tree when it is executed. Register may be called more than once for a command;
// [Binder.RegisterInNamespace] keeps its structs' keys apart.
//
// Register adds --config to cmd once. Files layer in order over [Options].BaseConfig, later keys winning per key.
//
//	app --config base.toml --config prod.toml
func (b *Binder) Register[T any](cmd *cobra.Command, target *T, opts ...RegisterOption[T]) error {
	return b.RegisterInNamespace(cmd, "", target, opts...)
}

// RegisterInNamespace is [Binder.Register] with every key of target under namespace, so structs registered on one
// command can share field names. With namespace "Database", URL is --database.url, APP_DATABASE_URL, and URL in a
// Database section of a config file. An empty namespace is the same as Register.
func (b *Binder) RegisterInNamespace[T any](
	cmd *cobra.Command, namespace string, target *T, opts ...RegisterOption[T],
) error {
	setups := make([]func(*targetEntry) (func(), error), 0, len(opts))
	for _, o := range opts {
		if o.setup == nil {
			continue
		}

		setups = append(setups, o.setup)
	}

	return b.register(cmd, target, namespace, setups...)
}

func (b *Binder) register(
	cmd *cobra.Command, target any, namespace string, setups ...func(*targetEntry) (func(), error),
) error {
	if len(b.entries[cmd]) == 0 {
		if cmd.Flags().Lookup(ConfigFlagName) != nil || cmd.PersistentFlags().Lookup(ConfigFlagName) != nil {
			return fmt.Errorf("flag --%s is already defined on %s", ConfigFlagName, cmd.Name())
		}

		// StringArray, not StringSlice, so a path may contain a comma.
		cmd.PersistentFlags().StringArray(ConfigFlagName, nil,
			"path to a config file; repeat to layer files, later ones winning")
	}

	entry := &targetEntry{
		b:         b,
		cmd:       cmd,
		target:    target,
		namespace: namespace,
	}
	if err := bindStruct(entry); err != nil {
		return err
	}

	// Options are checked before any take effect, so an errored option leaves none applied.
	installs := make([]func(), len(setups))
	for i, setup := range setups {
		install, err := setup(entry)
		if err != nil {
			return err
		}

		installs[i] = install
	}

	for _, install := range installs {
		install()
	}

	b.entries[cmd] = append(b.entries[cmd], entry)
	return nil
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
		cc.entries = append(slices.Clone(b.entries[c]), cc.entries...)
	}

	lang := b.opts.Markup
	// Env vars fold '.', '-' and '_' to '_', so distinct keys can still share a name.
	// Env var names are stricter than flag names, so if they don't collide, flags don't either.
	claimed := map[string]string{}
	for _, e := range cc.entries {
		for _, k := range e.keys {
			leaf := fileValueType(commentparsing.DerefType(k.goType), lang)
			if err := cc.keys.add(k.fileKey, leaf, lang); err != nil {
				return commandConfig{}, fmt.Errorf("%s: %w", e.cmd.Name(), err)
			}

			if k.flag == nil {
				continue
			}

			for _, name := range e.envVars(k) {
				if other, dup := claimed[name]; dup {
					return commandConfig{}, fmt.Errorf("%s: %s and %s are both %s", e.cmd.Name(), other, k.key, name)
				}

				claimed[name] = k.key
			}
		}
	}

	return cc, nil
}

// Undocumented lists, sorted, the keys of flags without help text, usually because their struct's package has no
// generated DocComments. This can be used in a test, for example, as require.Empty(t, b.Undocumented()).
func (b *Binder) Undocumented() []string {
	var keys []string
	for _, entries := range b.entries {
		for _, e := range entries {
			keys = append(keys, e.undocumented...)
		}
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

	if cc.fileValues, err = b.loadConfigFiles(c, cc.keys); err != nil {
		return err
	}

	var errs []error
	for _, entry := range cc.entries {
		if err = decodeEntry(entry, cc); err != nil {
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

		keys.overlay(merged, v.Elem())
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
	entries    []*targetEntry
	keys       *keyNode
	fileValues reflect.Value
}
