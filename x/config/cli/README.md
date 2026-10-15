# cli

Bind a Go config struct to a [cobra](https://github.com/spf13/cobra) command. Each field is
settable by flag, env var, config file, or the struct's own value, in that order of precedence.

```go
type Config struct {
	// Host is the host to dial.
	Host string `validate:"required"`
}

var cfg = Config{Host: "localhost"}

root := &cobra.Command{Use: "app", RunE: run}

// One Binder per command tree.
b, err := cli.New(cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})

// --host / APP_HOST / the config file key Host. Register also adds --config.
err = b.Register(root, &cfg)
```

When `RunE` runs, `cfg` is populated and its `validate` tags have passed. A failure names the key and how to set it:

```
invalid configuration: Host failed on the 'required' tag; set it with --host, APP_HOST, Host in a config file
```

Decoding is wired into the command's hooks when it runs, so a `PreRunE` or `PersistentPreRunE`
assigned after `Register` still sees the decoded struct.

## Config files

`--config` may be repeated; later files win key by key. Sections and maps merge. A list is one value, so a later
source's replaces an earlier one's. With the TOML markup from the example above:

```
app --config base.toml --config prod.toml
```

## Subcommands

A command runs with its own structs and every ancestor's. So subcommands can share the root's
config and keep their own apart:

```go
b.Register(root, &shared)    // every command
b.Register(embed, &embedCfg) // embed only
b.Register(run, &runCfg)     // run only
```

`embed`'s flags, env vars, and config file keys are those of `shared` and `embedCfg`; a `run` key in
its config file is unknown. A key, flag, or env var that a command and an ancestor both use fails
every command in the tree when it runs. Siblings may reuse keys.

## Namespaces

`Register` may be called more than once for a command. `RegisterInNamespace` puts a struct's
keys under a name, so structs from several packages can share field names:

```go
b.RegisterInNamespace(root, "Database", &dbCfg) // --database.url, APP_DATABASE_URL
b.RegisterInNamespace(root, "EVM", &evmCfg)     // --evm.url, APP_EVM_URL
```

In a config file each struct's keys sit in a section of that name. See
[`examples/namespaced`](examples/namespaced).

## Help text

A flag's help text is the field's doc comment, via the `DocComments` method
[`commentparsing`](../commentparsing) generates. The struct must live in its own package with a
`//go:generate` directive; see [`examples/simple`](examples/simple). Without it, flags bind with no
help text; `require.Empty(t, b.Undocumented())` in a test catches that.

Help also names each flag's env vars, and marks `required` fields `(required)` unless they
sit in an optional (pointer) section. A default is shown as its type's `MarshalText` renders it, so
hold a secret in `pkg/config.SecretString` or `SecretURL` to show it redacted. A type with
`UnmarshalText` but no `MarshalText` shows its default as `<cannot display>`.

## What gets a flag

| field                                                   | command line                              |
|---------------------------------------------------------|-------------------------------------------|
| scalar, `time.Duration`, `[]byte`, any `encoding.TextUnmarshaler` | `--timeout 5s`                 |
| nested struct                                           | `--chain.id 1`                            |
| pointer struct                                          | same; stays `nil` until one of its keys is set |
| list of those                                           | `--tags a,b` or `--tags a --tags b`       |
| map of those                                            | `--labels env=prod,region=us`             |
| map entry, like a struct's field                        | `--labels.env prod`, `--chains.mainnet.rpc x` |

Map keys parse like values, so `map[uint32]string` binds as `--chains 1=mainnet`. A config file's
keys are parsed as their flag would be, too; two texts that parse to one key, such as `16` and `0x10`,
are an error. Anything more structured is config file only.

## Map entries

A map's entries are set like a struct's fields, with the key in the flag or env var name. For

```go
type Config struct {
	Chains map[string]Chain
}

type Chain struct {
	RPC string
	WS  string
}
```

`--chains.mainnet.rpc x` and `APP_CHAINS_MAINNET_RPC=x` set `RPC` in entry `mainnet`, leaving the
config file's `WS` for it alone. A map inside an entry works the same way, one more key deep.
`--help` shows these as `--chains.<key>.rpc`.

A key here is one segment: no `.` in a flag name, no `_` in an env var name. So a name reads one
way or not at all; `--chains.main.net.rpc` is an unknown flag. Set such a key in a whole map,
`--labels main.net=x`, or a config file. An env var name is upper case, so its key is the existing
key, from the struct or a config file, that matches ignoring case, or else its lower case:
`APP_LABELS_MAINNET` sets `MainNet` if the map has it, and `mainnet` if not.

`--labels env=prod` and `--labels.env prod` both set entry `env`, so giving both is an error, as
are `APP_LABELS=env=prod` and `APP_LABELS_ENV=prod`. Across sources, precedence is as for any field.

Register sets the command's global normalization func, which pflag calls on each name it parses,
to add these flags as they appear. It wraps one set before `Register`; one set after replaces it,
and entry flags are then unknown.

Lists and maps are CSV; quote an element or whole entry containing a comma:

```
--tags '"a,b",c'               a,b and c
--labels '"env=a,b",region=us' env=a,b and region=us
```

A map key can't contain `=`, since an entry splits at its first `=`; set such keys in a config file.

## Types

pflag parses every flag, including each list element and map value, so a bad value fails as the
command line is parsed. An env var is parsed exactly as its flag is. A config file's values are
the markup's to decode. With the TOML markup in the example above, for instance, `Port = "8080"`
does not fill an `int`, `300` does not fit an `int8`, and neither a `time.Duration` nor a `[]byte`
takes a string.

## Where to look

- [`Options`](options.go): env var prefixes, base config, default config path.
- [`New` / `Binder.Register`](binder.go): registration, on the root or a subcommand. Flags are
  persistent, so subcommands take them too.
- [`Profile` / `ProfileWithSelector`](profile.go): build a struct's or a section's defaults with a
  function picked by one of its top-level fields; see [`examples/profiles`](examples/profiles) and
  [`examples/nested_profiles`](examples/nested_profiles).
- [`examples/`](examples): runnable programs.
