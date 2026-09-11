# namespaced

Two config structs registered on the same command. Both declare a field called `URL`; the
namespace is what stops them colliding, and it does so identically across all three sources.

| struct           | namespace  | flag                     | env var                    | config key             |
|------------------|------------|--------------------------|----------------------------|------------------------|
| `DatabaseConfig` | `database` | `--database.url`         | `APP_DATABASE_URL`         | `[database] url`       |
|                  |            | `--database.max-open`    | `APP_DATABASE_MAX_OPEN`    | `[database] max-open`  |
| `EVMConfig`      | `evm`      | `--evm.url`              | `APP_EVM_URL`              | `[evm] url`            |
|                  |            | `--evm.chain-id`         | `APP_EVM_CHAIN_ID`         | `[evm] chain-id`       |

Each struct is decoded and validated on its own, so a mistake in one is reported without hiding
whatever is wrong with the other.

## Walking the fallback chain

```sh
# 1. compiled-in defaults
go run .

# 2. a config file fills both namespaces
go run . --config example.toml

# 3. env vars beat the file, one namespace at a time
APP_DATABASE_URL=postgres://env-host:5432/app go run . --config example.toml

# 4. flags beat env vars, and the two URLs stay independent
APP_DATABASE_URL=postgres://env-host:5432/app \
  go run . --config example.toml --database.url postgres://flag-host:5432/app --evm.url https://flag-host:8545
```

The point of step 4: `--database.url` and `--evm.url` are separate flags, so setting one leaves
the other at whatever the file or env supplied.

## Help

```sh
go run . --help
```
