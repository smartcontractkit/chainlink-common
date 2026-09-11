module example.com/upstream

go 1.26.6

require github.com/smartcontractkit/chainlink-common/x/config v0.0.0

// This replace exists only because the example lives inside the library's own repository. A real
// dependency requires a released version and needs no directive.
replace github.com/smartcontractkit/chainlink-common/x/config => ../../../..
