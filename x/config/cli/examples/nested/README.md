# Nested configurations

Sections of a configuration can be embedded structs, value structs, or pointers to structs:

- `Logging` is embedded, so its fields are at the top level: `--log-level` and `APP_LOG_LEVEL`, not
  `--logging.log-level` and `APP_LOGGING_LOG_LEVEL`.
- `Server` is a value struct, so its field is prefixed: `--server.host` and `APP_SERVER_HOST`.
- `Metrics` is a pointer, so it stays `nil` until one of its fields is set. Its field is prefixed:
  `--metrics.endpoint` and `APP_METRICS_ENDPOINT`.

```sh
# default values
go run .

# TOML values
go run . --config example.toml

# environment values
APP_LOG_LEVEL=debug APP_SERVER_HOST=0.0.0.0 APP_METRICS_ENDPOINT=https://metrics.example.com/push go run .

# CLI values
go run . --log-level debug --server.host 0.0.0.0 --metrics.endpoint https://metrics.example.com/push
```
