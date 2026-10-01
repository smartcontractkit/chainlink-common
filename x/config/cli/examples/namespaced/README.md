# Namespaced configurations

Two config structs registered on one command, both with a `URL` field. `RegisterInNamespace` puts each
under its own name, so they stay apart in flags, env vars, and the config file:

- `DatabaseConfig` under `Database`: `--database.url`, `APP_DATABASE_URL`, and `URL` in a `[Database]` section.
- `EVMConfig` under `EVM`: `--evm.url`, `APP_EVM_URL`, and `URL` in an `[EVM]` section.

Both share one `--config`.

```sh
# default values
go run .

# TOML values
go run . --config example.toml

# environment values
APP_DATABASE_URL=postgres://env.example.com:5432/app APP_EVM_URL=https://env.example.com:8545 go run .

# CLI values
go run . --database.url postgres://cli.example.com:5432/app --evm.url https://cli.example.com:8545
```
