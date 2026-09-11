module example.com/consumer

go 1.26.6

require (
	example.com/upstream v0.0.0
	github.com/smartcontractkit/chainlink-common/x/config v0.0.0
)

// Both replaces exist only because this example lives inside the library's own repository. A real
// consumer requires released versions of each and needs neither directive.
replace (
	example.com/upstream => ./upstream
	github.com/smartcontractkit/chainlink-common/x/config => ../../..
)
