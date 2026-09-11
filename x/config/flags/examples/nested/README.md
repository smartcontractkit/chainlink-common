# nested

Nested sections, and the rule for when one has to be a pointer.

| section   | Go type          | why |
|-----------|------------------|-----|
| `Server`  | `ServerConfig`   | always present, so a value struct; its `Host` is `required` all the same |
| `TLS`     | `*TLSConfig`     | optional, so it must be able to be absent - it stays `nil` until something under it is set |
| `Metrics` | `*MetricsConfig` | mutually exclusive with `Tracing`, and a cross-field rule needs absence to be representable |
| `Tracing` | `*TracingConfig` | the other half of that pair |

Keys nest with the section: `--server.host`, `APP_SERVER_HOST`, `[server] host`.

## A required field does not need a pointer

`Server.Host` is `validate:"required"` on a plain value struct. Required describes the *leaf*,
not the section, so there is nothing for a pointer to express:

```sh
go run . --metrics.endpoint https://m.example/push --server.host ""
# invalid configuration: ... 'Host' failed on the 'required' tag
```

## A pointer section stays nil when nothing under it is set

```sh
go run . --metrics.endpoint https://m.example/push
# tls = <nil> (not configured)
```

Nothing under `tls` was supplied, so it is never allocated - and because it is never allocated,
its own `required` fields are not reported for a section nobody asked for.

Name any one of its keys and it is allocated, at which point those fields do apply:

```sh
go run . --metrics.endpoint https://m.example/push --tls.cert-file /etc/cert.pem
# invalid configuration: ... 'KeyFile' failed on the 'required' tag

go run . --metrics.endpoint https://m.example/push \
  --tls.cert-file /etc/cert.pem --tls.key-file /etc/key.pem
# tls = cert-file=/etc/cert.pem key-file=/etc/key.pem
```

## Exactly one of two sections

```sh
# neither: rejected
go run .

# one: accepted
go run . --metrics.endpoint https://m.example/push
go run . --tracing.collector https://t.example/v1/traces

# both: rejected
go run . --metrics.endpoint https://m.example/push --tracing.collector https://t.example/v1/traces
```

A value struct carrying `required_without`/`excluded_with` is rejected at registration rather
than mis-validated at run time: a zero struct is still a struct, so the rule could never fire.

## Walking the fallback chain

```sh
# 1. compiled-in defaults, plus the one section the rules demand
go run . --metrics.endpoint https://m.example/push

# 2. a config file fills server and metrics
go run . --config example.toml

# 3. an env var beats the file
APP_SERVER_PORT=7000 go run . --config example.toml

# 4. a flag beats the env var
APP_SERVER_PORT=7000 go run . --config example.toml --server.port 6000
```

## Help

```sh
go run . --help
```
