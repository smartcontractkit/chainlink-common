// Package consumer is a config package in its own module whose tree reaches into a dependency.
package consumer

//go:generate go -C upstream generate ./...
//go:generate go run ./gen

import "example.com/upstream"

// Config is the program's configuration, built partly from a dependency's types.
type Config struct {
	upstream.Settings

	// Endpoint is the upstream URL to dial.
	Endpoint string `toml:"endpoint" validate:"required,url"`

	// Retry tunes how a failed request is repeated.
	Retry upstream.Retry `toml:"retry"`
}
