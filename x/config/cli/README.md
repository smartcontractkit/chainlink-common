# cli

Bind a Go config struct to a [cobra](https://github.com/spf13/cobra) command. Each field is
settable by flag, env var, config file, or the struct's own value, in that order of precedence.

```go
type Config struct {
	// Host is the host to dial.
	Host string `toml:"host" validate:"required"`
}

var cfg = Config{Host: "localhost"}

root := &cobra.Command{Use: "app", RunE: run}

// One Binder per command tree; it also adds --config.
b, err := cli.New(root, cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})

// --host / APP_HOST / host = "..." in the config file
err = b.Register(root, "", &cfg)
```

When `RunE` runs, `cfg` is populated and its `validate` tags have passed. The extra `set` rule
requires that a source supplied the field, whatever its value. A failure names the key and how to
set it:

```
invalid configuration: host failed on the 'required' tag; set it with --host, APP_HOST, host in a config file
```

Decoding is wired into the command's hooks when it runs, so a `PreRunE` or `PersistentPreRunE`
assigned after `Register` still sees the decoded struct.

## Config files

`--config` may be repeated; later files win key by key. Tables and maps merge, and so does a map
from an env var or flag, over the file's and the default's. A list is one value, so a later
source's replaces an earlier one's:

```
app --config base.toml --config prod.toml
```

Without `--config`, `config.<extension>` in the working directory, such as tomlmarkup's
`config.toml`, is read if present. Each file is decoded by the markup itself, so keys, types, and
which types read themselves whole are exactly the markup's (for tomlmarkup, go-toml's); a key no
field reads is an error. A field the markup ignores
(tagged `-`) is read from no source: no config key, flag, or env var.

## Help text

A flag's help text is the field's doc comment, via the `DocComments` method
[`commentparsing`](../commentparsing) generates. The struct must live in its own package with a
`//go:generate` directive; see [`examples/simple`](examples/simple). Without it, flags bind with no
help text; `require.Empty(t, b.Undocumented())` in a test catches that.

Help also names each flag's env vars, and marks `required` / `set` fields `(required)` unless they
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
its flag would be; one key spelled two ways, such as `16` and `0x10`, is an error. Anything more
structured is config file only.

Lists and maps are CSV; quote an element or whole entry containing a comma:

```
--tags '"a,b",c'               a,b and c
--labels '"env=a,b",region=us' env=a,b and region=us
```

## Types

pflag parses every flag, including each list element and map value, so a bad value fails as the
command line is parsed. An env var is parsed exactly as its flag is. A config file's values are
the markup's to decode: with tomlmarkup, `port = "8080"` does not fill an `int`, `300` does not fit
an `int8`, and neither a `time.Duration` nor a `[]byte` takes a string.

## Where to look

- [`Options`](options.go): env var prefixes, base config, default config path.
- [`New` / `Binder.Register`](register.go): registration, and namespacing: `b.Register(root, "database", &db)`
  gives `database.url`, `--database.url` and `APP_DATABASE_URL`, on the root or a subcommand; `""`
  gives `url`, `--url` and `APP_URL`. Flags are persistent, so subcommands take them too.
- [`examples/`](examples): runnable programs.
