# Simple configuration

One config struct bound to one command. Each setting can come from, in order of precedence:

1. a CLI flag: `--host`
2. an environment variable, named with the `APP` prefix: `APP_HOST`
3. a config file: `Host` in `example.toml`
4. the default in `main.go`

```sh
# default values
go run .

# TOML values
go run . --config example.toml

# environment values
APP_HOST=env.example.com APP_PORT=7070 APP_TIMEOUT=4s go run .

# CLI values
go run . --host cli.example.com --port 6060 --timeout 11s

# all sources: host from the CLI, port from the environment, the rest from TOML
# APP_HOST is intentionally left in the call to demonstrate precedence.
APP_HOST=env.example.com APP_PORT=7070 go run . --config example.toml --host cli.example.com
```
