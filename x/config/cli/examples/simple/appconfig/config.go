package appconfig

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

type Config struct {
	// Host is the host to dial.
	Host string

	// Port is the port to dial on.
	Port int

	// Timeout bounds a single request.
	Timeout config.Duration

	// Tags are labels to attach to every request.
	Tags []string
}
