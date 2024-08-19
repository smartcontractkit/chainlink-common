# Capabilities CLI tool

This tool was built for generating capability Go types from the JSON schema.

## Conventions

- Capability JSON schemas must be named using this pattern `[capability_name]_[capability_type].json`.
- Generated types are placed next to the JSON schema and named using this pattern `[capability_name]_[capability_type]_generated.go`.

## Running

_All commands run from the root of the repo._

Help:

```bash
go run ./pkg/capabilities/cli/cmd/generate --help
```

Generate types:

```bash
go run ./pkg/capabilities/cli/cmd/generate-types
```
