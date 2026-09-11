# flags

Bind a Go config struct to a [cobra](https://github.com/spf13/cobra) command. One struct is the
single description of a program's configuration; a CLI flag, an environment variable, a config
file and the struct's own compiled-in values all follow from it and resolve in that order of
precedence.

```go
type Config struct {
	Host    string          `toml:"host" usage:"host to dial"`
	Timeout config.Duration `toml:"timeout" usage:"per-request timeout" validate:"required"`
}

var cfg = Config{Host: "localhost"}

root := &cobra.Command{Use: "app", RunE: run}
root.PersistentFlags().String("config", "", "path to a config file")

// --host / APP_HOST / host = "..." in the config file
err := flags.RegisterCommandFlags(root, &cfg, flags.DefaultTOMLOptions("APP"))
```

By the time `RunE` is called, `cfg` is populated and its `validate` tags have been checked.

## What is bound

| struct shape                  | command line |
|-------------------------------|--------------|
| scalar, `time.Duration`, any `encoding.TextUnmarshaler` | `--timeout 5s` |
| nested struct                 | `--chain.id 1` |
| pointer struct                | the same, and it stays `nil` until one of its keys is set |
| list                          | `--tags a,b` or `--tags a --tags b` |
| list of lists                 | `--urls '[a,b],[c]'` or one inner list per occurrence |
| list of structs               | one flag per field, lining up by position: `--nodes.name a,b --nodes.url ua,ub` |
| map of text values or lists   | `--labels 'env=prod;region=us'`, `--urls 'primary=a,b'` |
| map of structs, interfaces, … | no flag; config file only |

## Where to look

- [`Options`](options.go) - tag conventions, env var prefixes, namespacing, decoding. It is the
  extension point: new behaviour is added as a field there.
- [`RegisterCommandFlags`](register.go) / `RegisterSubcommandFlags` - registration and the rules
  for `validate` tags.
- [`RegisterProfile`](profile.go) - fill a section's unset fields from a map keyed by one of its
  own fields (a chain ID, an environment name).
- [`examples/`](examples) - runnable programs for the common shapes.
