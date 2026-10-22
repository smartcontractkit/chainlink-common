// Package simple is a config package that documents itself.
package simple

//go:generate go run ./gen

import "time"

// Config is the program's configuration.
type Config struct {
	// Endpoint is the upstream URL to dial.
	// It must include a scheme, and a trailing path is preserved.
	Endpoint string `toml:"endpoint" validate:"required,url"`

	// Timeout bounds a single request. Zero means no timeout.
	Timeout time.Duration `toml:"timeout"`

	Retries int `toml:"retries" validate:"gte=0,lte=10"` // Retries is the attempt count after the first.
}
