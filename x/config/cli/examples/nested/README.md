# nested

When a section must be a pointer.

- `Logging` is embedded, so its fields bind at the top level: `--log-level`, not `--logging.log-level`.
- `TLS` is `*TLSConfig`: it stays `nil` until one of its keys is set, so its `required` fields
  only apply once it is configured.
- `Metrics` and `Tracing` are mutually exclusive, so both are pointers. A value struct with a
  cross-field rule is rejected at registration.

```sh
go run . --metrics.endpoint https://m.example/push
# tls = <nil> (not configured)

go run . --metrics.endpoint https://m.example/push --tls.cert-file /etc/cert.pem
# tls.key-file failed on the 'required' tag; set it with --tls.key-file, APP_TLS_KEY_FILE, tls.key-file in a config file

go run . --metrics.endpoint https://m.example/push --tracing.collector https://t.example/v1/traces
# metrics failed on the 'excluded_with=tracing' tag
# tracing failed on the 'excluded_with=metrics' tag
```
