# simple

One config struct bound to one command. Every field gets a flag, an env var and a config file
key derived from the same `toml` tag, so all four sources describe the same setting.

| field     | flag        | env var       | config key |
|-----------|-------------|---------------|------------|
| `Host`    | `--host`    | `APP_HOST`    | `host`     |
| `Port`    | `--port`    | `APP_PORT`    | `port`     |
| `Timeout` | `--timeout` | `APP_TIMEOUT` | `timeout`  |
| `Tags`    | `--tags`    | `APP_TAGS`    | `tags`     |

## Walking the fallback chain

Run each of these from this directory and watch `port` change while everything else keeps the
value the previous step gave it.

```sh
# 1. compiled-in defaults only
go run .
#    host = localhost / port = 8080 / timeout = 5s / tags = [default]

# 2. a config file beats the defaults
go run . --config example.toml
#    host = file.example.com / port = 9090 / timeout = 30s / tags = [from-file toml]

# 3. an env var beats the config file
APP_PORT=9999 go run . --config example.toml
#    port = 9999, the rest still from the file

# 4. a flag beats the env var
APP_PORT=9999 go run . --config example.toml --port 1234
#    port = 1234
```

A file named `config.toml` in the working directory is picked up without `--config`:

```sh
cp example.toml config.toml && go run . ; rm config.toml
```

## Lists and durations

A list takes either form, and repeating the flag accumulates:

```sh
go run . --tags a,b
go run . --tags a --tags b
APP_TAGS=a,b go run .
```

`config.Duration` parses its usual text form and rejects a negative value:

```sh
go run . --timeout 90s
go run . --timeout -1s   # error: 'timeout' cannot make negative time duration: -1s
```

## Help

```sh
go run . --help
```
