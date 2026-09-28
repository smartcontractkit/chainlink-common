// Package settings holds the simple example's config struct.
//
// It is a package of its own so that the generator carrying its doc comments into the build can
// import it, which a package main cannot be. The comments here are what `--help` prints.
package settings

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

// Config is everything the program can be told.
type Config struct {
	// Host is the host to dial.
	Host string `toml:"host"`

	// Port is the port to dial.
	Port int `toml:"port"`

	// Timeout bounds a single request.
	// A negative value is rejected.
	Timeout config.Duration `toml:"timeout"`

	// Tags are labels to attach to every request.
	Tags []string `toml:"tags"`
}
