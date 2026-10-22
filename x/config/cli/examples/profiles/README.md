# Profiles

A profile is a function that builds a registered struct from its defaults, picked by one of its top-level fields. Here,
`ChainConfig` is registered in the `Chain` namespace, and `Chain.ID` picks a function from `chains` in `main.go`, which
sets the chain's `BlockTime`, and for chains 1 and 137 its `RPC`. Chains 1 and 10 leave `Finality` alone, so both share
the default of 64 from `chain`; chain 137 replaces it with 128. A profile can set any value, zero included. To profile
a section of a larger struct instead, see [`nested_profiles`](../nested_profiles).

`RPC` is required, and chain 10's profile doesn't provide one, so chain 10 needs an RPC from a flag, env var, or config
file.

- Any source still wins: a flag, env var, or config file value for a field replaces the profile's.
- An ID with no profile changes nothing, so its fields come from the sources alone.
- `Chain.ID` is tagged `validate:"set"`, so a source must pick the chain; there is no default.
- `--help` lists each profile with the values it gives, defaults included.

```sh
# no chain picked: fails, naming every way to set it
go run .

# chain 1 and its profile
go run . --chain.id 1

# chain 10 without an RPC: fails, since its profile doesn't provide one
go run . --chain.id 10

# chain 10 with an RPC; it shares the default finality
go run . --chain.id 10 --chain.rpc https://differentchain.example.com

# a chain whose profile replaces the default finality
go run . --chain.id 137

# a profile with one field overridden
go run . --chain.id 137 --chain.finality 10

# a chain with no profile
APP_CHAIN_ID=999 go run . --chain.rpc https://local.example.com

# TOML values: chain 137, with its own RPC
go run . --config example.toml

# each profile and its values
go run . --help
```
