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

`--config` may be repeated; later files win key by key. Sections and maps merge, and so does a map
from an env var or flag, over the file's and the default's. A list is one value, so a later
source's replaces an earlier one's. With the TOML markup from the example above:

```
app --config base.toml --config prod.toml
```

## Subcommands

A command runs with its own struct and every ancestor's, and each command takes at most one. So
subcommands can share the root's config and keep their own apart:

```go
b.Register(root, &shared)    // every command
b.Register(embed, &embedCfg) // embed only
b.Register(run, &runCfg)     // run only
```

`embed`'s flags, env vars, and config file keys are those of `shared` and `embedCfg`; a `run` key in
its config file is unknown. A key, flag, or env var that a command and an ancestor both use fails
every command in the tree when it runs. Siblings may reuse keys.

## Help text

A flag's help text is the field's doc comment, via the `DocComments` method
[`commentparsing`](../commentparsing) generates. The struct must live in its own package with a
`//go:generate` directive; see [`examples/simple`](examples/simple). Without it, flags bind with no
help text; `require.Empty(t, b.Undocumented())` in a test catches that.

Help also names each flag's env vars, and marks `required` fields `(required)` unless they
sit in an optional (pointer) section. A default is shown as its type's `MarshalText` renders it, so
hold a secret in `pkg/config.SecretString` or `SecretURL` to show it redacted.

## What gets a flag

| field                                                   | command line                              |
|---------------------------------------------------------|-------------------------------------------|
| scalar, `time.Duration`, `[]byte`, any `encoding.TextUnmarshaler` | `--timeout 5s`                 |
| nested struct                                           | `--chain.id 1`                            |
| pointer struct                                          | same; stays `nil` until one of its keys is set |
| list of those                                           | `--tags a,b` or `--tags a --tags b`       |
| map of those                                            | `--labels env=prod,region=us`             |

Map keys parse like values, so `map[uint32]string` binds as `--chains 1=mainnet`. In a config file,
where keys are text, a key type the markup reads whole is used as it is, and any other is parsed as
its flag would be; two texts that parse to one key, such as `16` and `0x10`, are an error. Anything more
structured is config file only.

Lists and maps are CSV; quote an element or whole entry containing a comma:

```
--tags '"a,b",c'               a,b and c
--labels '"env=a,b",region=us' env=a,b and region=us
```

## Types

pflag parses every flag, including each list element and map value, so a bad value fails as the
command line is parsed. An env var is parsed exactly as its flag is. A config file's values are
the markup's to decode. With the TOML markup in the example above, for instance, `Port = "8080"`
does not fill an `int`, `300` does not fit an `int8`, and neither a `time.Duration` nor a `[]byte`
takes a string.

## Where to look

- [`Options`](options.go): env var prefixes, base config, default config path.
- [`New` / `Binder.Register`](binder.go): registration, on the root or a subcommand. Flags are
  persistent, so subcommands take them too. To bind structs from several packages to one command,
  wrap them in one: a `Database` field gives `Database.URL`, `--database.url` and
  `APP_DATABASE_URL`.
- [`examples/`](examples): runnable programs.
