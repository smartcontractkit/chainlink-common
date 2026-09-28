// Package settings holds the namespaced example's two config structs, standing in for two
// packages that each own one.
//
// They are out of package main so that the generator carrying their doc comments into the build
// can import them; a real pair of structs owned by a database and a chain package is already in
// packages of its own. The comments here are what `--help` prints.
package settings

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

// DatabaseConfig stands in for a struct owned by a database package.
type DatabaseConfig struct {
	// URL is the connection string.
	URL string `toml:"url"`

	// MaxOpen is the maximum number of open connections.
	MaxOpen int `toml:"max-open"`

	// DialTimeout is how long to wait for a connection.
	DialTimeout config.Duration `toml:"dial-timeout"`
}

// EVMConfig stands in for a struct owned by a chain package. Its URL is a different setting from
// the database's, and only the namespace says so.
type EVMConfig struct {
	// URL is the RPC endpoint.
	URL string `toml:"url"`

	// ChainID is the chain selector.
	ChainID uint64 `toml:"chain-id"`
}
