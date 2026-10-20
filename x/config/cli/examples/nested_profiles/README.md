# Nested profiles

[`profiles`](../profiles) registers the chain on its own, so its profile builds the whole struct. Here the chain is one
section of a larger `Config`, registered whole, so a config file holds both `LogLevel` and `[Chain]`.
`cli.ProfileWithSelector` picks the `Chain` section, and `Chain.ID` (a top-level field of that section) picks a function
from `chains` in `main.go`. The profiles are the same as in `profiles`: they set the chain's `BlockTime`, and for chains 1
and 137 its `RPC`, and only chain 137 replaces the default `Finality` of 64.

A profile only builds its section: `LogLevel` comes from its default and the sources, whatever the chain.

```sh
# chain 1 and its profile; log-level keeps its default
go run . --chain.id 1

# a source outside the section, alongside a profile
go run . --chain.id 137 --log-level debug

# TOML values: log level and chain 137, with its own RPC
go run . --config example.toml

# each profile and its values
go run . --help
```
