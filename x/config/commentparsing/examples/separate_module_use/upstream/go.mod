module example.com/upstream

go 1.27.3

require github.com/smartcontractkit/chainlink-common/x/config v0.0.0

require (
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/smartcontractkit/chainlink-common v0.11.2-0.20261027153112-98a50810196d // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)

// This replace exists only because the example lives inside the library's own repository. A real
// dependency requires a released version and needs no directive.
replace github.com/smartcontractkit/chainlink-common/x/config => ../../../..
