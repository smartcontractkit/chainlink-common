module example.com/upstream

go 1.26.6

require github.com/smartcontractkit/chainlink-common/x/config v0.0.0

require (
	github.com/smartcontractkit/chainlink-common v0.9.6-0.20260206011444-ed1fb0284e5d // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
)

// This replace exists only because the example lives inside the library's own repository. A real
// dependency requires a released version and needs no directive.
replace github.com/smartcontractkit/chainlink-common/x/config => ../../../..
