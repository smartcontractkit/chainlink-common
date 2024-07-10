# Capabilities CLI tool

This tool was built for generating capability Go types from the JSON schema.

## Conventions

- Capability JSON schemas must end with `.capability.json`.
- Generated types are placed next to the JSON schema with the `.generated.go` extension.

## Running

_All commands run from the root of the repo._

Help:

```bash
go run ./pkg/capabilities/cli/...
```

Generate types:

```bash
go run ./pkg/capabilities/cli/... generate-types
```
