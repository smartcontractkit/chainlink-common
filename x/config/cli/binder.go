package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup"
)

// Binder binds config structs to the commands of one tree. Everything the tree shares, from the
// config language to the env prefixes and the --config files, is set once, in [New].
type Binder struct {
	root *cobra.Command
	opts Options

	// entries is every target registered anywhere in the tree, so unknown keys are judged against the whole binary
	entries []*targetEntry

	// keys is where a config file holds every field in entries.
	keys *keyNode

	// config is Options.BaseConfig with the files layered over it, of keys' fileType, read once
	// per execution.
	config         reflect.Value
	configFileOnce sync.Once
	configFileErr  error
}

// New returns root's Binder and adds a repeatable persistent --config flag to root, which must
// not already have a flag of that name. Later files win key by key; tables and maps merge. With
// tomlmarkup:
//
//	app --config base.toml --config prod.toml
//
// opts.Markup is required; x/config/markup/tomlmarkup is typical.
func New(root *cobra.Command, opts Options) (*Binder, error) {
	if opts.Markup == nil {
		return nil, markup.Err
	}

	if root.Flags().Lookup(ConfigFlagName) != nil || root.PersistentFlags().Lookup(ConfigFlagName) != nil {
		return nil, fmt.Errorf("flag --%s is already defined on %s; New adds it", ConfigFlagName, root.Name())
	}

	b := &Binder{root: root, opts: opts, keys: &keyNode{children: map[string]*keyNode{}}}
	cobra.OnInitialize(b.wire)

	// StringArray, not StringSlice, so a path may contain a comma.
	root.PersistentFlags().StringArray(ConfigFlagName, nil,
		"path to a config file; repeat to layer files, later ones winning")
	return b, nil
}

const (
	// defaultConfigName is the default file's base name when Options.DefaultConfigPath is unset.
	defaultConfigName = "config"

	// ConfigFlagName is the flag the config files are read from, which [New] adds.
	ConfigFlagName = "config"
)

// Register binds target's fields as flags and env vars on cmd, and decodes target (config file, env, flags)
// before cmd runs. It may be called repeatedly on one command; each target is decoded and validated independently.
//
// namespace roots target's config keys, env vars, and flags: "database" gives database.url,
// PREFIX_DATABASE_URL, and --database.url; "" roots them at the top level. It is independent of
// the command, so one command can take several structs that share field names. Flags are
// persistent, so cmd's subcommands take them and decode target too. Flags that collide are an error.
//
// `validate` tags run after decoding, with one added rule: `set` requires that a source supplied
// the field, even if it's the default value set.
//
// Only fields with a single text form get flags and env vars: scalars, durations, text
// unmarshalers, and lists or maps of those (see list.go, maps.go). Others are config file only.
//
// Help text is the field's doc comment, via the
// [github.com/smartcontractkit/chainlink-common/x/config/commentparsing] generator; without it,
// fields bind with no help.
//
// opts configure this registration alone.
func (b *Binder) Register[T any](cmd *cobra.Command, namespace string, target *T, opts ...RegisterOption[T]) error {
	setups := make([]func(*targetEntry) (func(), error), len(opts))
	for i, o := range opts {
		if o.setup == nil {
			return errors.New("zero cli.RegisterOption; build one with an option function")
		}
		setups[i] = o.setup
	}
	return b.register(cmd, namespace, target, setups...)
}

func (b *Binder) register(cmd *cobra.Command, namespace string, target any, setups ...func(*targetEntry) (func(), error)) error {
	entry := &targetEntry{
		b:         b,
		cmd:       cmd,
		namespace: namespace,
		target:    target,
		docs:      fieldDocs{},
		sections:  map[string]string{},
	}
	if err := bindStruct(entry); err != nil {
		return err
	}
	for _, k := range entry.keys {
		if err := b.keys.add(k.configPath, fileValueType(commentparsing.DerefType(k.goType), b.opts.Markup), b.opts.Markup); err != nil {
			return err
		}
	}

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

	b.entries = append(b.entries, entry)
	return nil
}

// Undocumented returns the config key of every flag bound with no help text, usually a struct
// whose DocComments were never generated. For a test: require.Empty(t, b.Undocumented()).
func (b *Binder) Undocumented() []string {
	var keys []string
	for _, e := range b.entries {
		keys = append(keys, e.undocumented...)
	}
	return keys
}

type (
	hookE = func(*cobra.Command, []string) error
	hook  = func(*cobra.Command, []string)
)

// wire hooks decoding into the tree as it stands when a command executes, so a hook assigned
// after Register is wrapped rather than replacing the decode. It runs from cobra.OnInitialize,
// before any hook. Decoding runs in the root's PersistentPreRunE.
func (b *Binder) wire() {
	b.wrap(&b.root.PersistentPreRunE, &b.root.PersistentPreRun)
	if cobra.EnableTraverseRunHooks {
		return
	}

	var visit func(*cobra.Command)
	visit = func(c *cobra.Command) {
		for _, child := range c.Commands() {
			// Cobra runs only the nearest persistent hook, so one below the root would skip the
			// root's decode; it decodes instead.
			if child.PersistentPreRunE != nil || child.PersistentPreRun != nil {
				b.wrap(&child.PersistentPreRunE, &child.PersistentPreRun)
			}
			visit(child)
		}
	}
	visit(b.root)
}

// wrap makes *slotE decode before running the hook it held, or the plain hook in *slot, which
// cobra would otherwise skip in favour of *slotE. A slot already wrapped is left.
func (b *Binder) wrap(slotE *hookE, slot *hook) {
	if *slotE != nil && reflect.ValueOf(*slotE).Pointer() == wrapperCode {
		return
	}
	*slotE = b.wrapper(*slotE, slot)
}

func (b *Binder) wrapper(prev hookE, slot *hook) hookE {
	return func(c *cobra.Command, args []string) error {
		if err := b.decode(c); err != nil {
			return err
		}
		if prev != nil {
			return prev(c, args)
		}
		// Read at run time, so a plain hook assigned after wrapping still runs.
		if *slot != nil {
			(*slot)(c, args)
		}
		return nil
	}
}

// wrapperCode identifies wrapper's closures, which is how wrap tells its own hook from a caller's.
var wrapperCode = reflect.ValueOf((&Binder{}).wrapper(nil, nil)).Pointer()

// decode decodes and validates every entry registered on c or an ancestor,
// whose flags c inherits. Errors are joined so one entry's failure doesn't hide another's.
func (b *Binder) decode(c *cobra.Command) error {
	// Help and completion must work without valid config.
	if IsBuiltinCommand(c) {
		return nil
	}

	onPath := map[*cobra.Command]bool{}
	for p := c; p != nil; p = p.Parent() {
		onPath[p] = true
	}
	entries := slices.DeleteFunc(slices.Clone(b.entries), func(e *targetEntry) bool { return !onPath[e.cmd] })
	if len(entries) == 0 {
		return nil
	}

	if err := b.loadConfigFilesOnce(c); err != nil {
		return err
	}

	var errs []error
	for _, entry := range entries {
		if err := decodeEntry(entry); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := entry.validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid configuration: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (b *Binder) loadConfigFilesOnce(cmd *cobra.Command) error {
	b.configFileOnce.Do(func() { b.config, b.configFileErr = b.loadConfigFiles(cmd) })
	return b.configFileErr
}

// loadConfigFiles layers the --config files, or the default path if none, over the base config,
// later files winning. Only named files must exist; every file read must decode.
func (b *Binder) loadConfigFiles(cmd *cobra.Command) (reflect.Value, error) {
	lang := b.opts.Markup
	fileType := b.keys.fileType(lang)
	merged := reflect.New(fileType).Elem()
	layer := func(data []byte, name string) error {
		v := reflect.New(fileType)
		if err := lang.Unmarshal(data, v.Interface()); err != nil {
			return fmt.Errorf("invalid %s: %w", name, err)
		}
		b.keys.overlay(merged, v.Elem())
		return nil
	}

	if len(b.opts.BaseConfig) > 0 {
		if err := layer(b.opts.BaseConfig, "Options.BaseConfig"); err != nil {
			return reflect.Value{}, err
		}
	}

	// New defined the flag, so it can't be missing or of another type.
	paths, _ := cmd.Flags().GetStringArray(ConfigFlagName)
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
