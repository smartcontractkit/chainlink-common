// Package nested is a config package whose tree reaches into a second package of this module.
package nested

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested/upstream"

// Config is the program's configuration.
type Config struct {
	upstream.Settings

	// Endpoint is the upstream URL to dial.
	Endpoint string `toml:"endpoint" validate:"required,url"`

	// Retry tunes how a failed request is repeated.
	Retry upstream.Retry `toml:"retry"`
}
