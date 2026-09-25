module example.com/consumer

go 1.27.1

require (
	example.com/upstream v0.0.0
	github.com/smartcontractkit/chainlink-common/x/config v0.0.0
)

require (
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/smartcontractkit/chainlink-common v0.9.6-0.20260206011444-ed1fb0284e5d // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
)

// Both replaces exist only because this example lives inside the library's own repository. A real
// consumer requires released versions of each and needs neither directive.
replace (
	example.com/upstream => ./upstream
	github.com/smartcontractkit/chainlink-common/x/config => ../../..
)
