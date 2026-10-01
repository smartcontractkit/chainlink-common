package settings

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

// DatabaseConfig stands in for a struct owned by a database package.
type DatabaseConfig struct {
	// URL is the connection string.
	URL string

	// DialTimeout is how long to wait for a connection.
	DialTimeout config.Duration
}

// EVMConfig stands in for a struct owned by a chain package.
type EVMConfig struct {
	// URL is the RPC endpoint.
	URL string

	// ChainID is the chain to connect to.
	ChainID uint64
}
