// Package settings holds the nested example's config structs, one per section.
//
// It is a package of its own so that the generator carrying its doc comments into the build can
// import it, which a package main cannot be. The comments here are what `--help` prints.
package settings

//go:generate go run ./gen

import "github.com/smartcontractkit/chainlink-common/pkg/config"

// ServerConfig is the always-present section: a value struct, not a pointer. Host being
// required is about the leaf, not the section - the section always exists, and this field inside
// it must be filled.
type ServerConfig struct {
	// Host is the address to listen on.
	Host string `toml:"host" validate:"required"`

	// Port is the port to listen on.
	Port int `toml:"port"`

	// Idle is the idle connection timeout.
	Idle config.Duration `toml:"idle"`
}

// TLSConfig is configured or absent, never partly either.
type TLSConfig struct {
	// CertFile is the path to a PEM certificate.
	CertFile string `toml:"cert-file" validate:"required"`

	// KeyFile is the path to the matching PEM private key.
	KeyFile string `toml:"key-file" validate:"required"`
}

// MetricsConfig is one half of a mutually exclusive pair.
type MetricsConfig struct {
	// Endpoint is where to push metrics.
	Endpoint *config.URL `toml:"endpoint" validate:"required"`
}

// TracingConfig is the other half.
type TracingConfig struct {
	// Collector is where to send spans.
	Collector *config.URL `toml:"collector" validate:"required"`
}

// Logging is embedded, so its fields sit at the top level rather than in a section.
type Logging struct {
	// LogLevel is the minimum level to log.
	LogLevel string `toml:"log-level"`
}

// Config is the whole configuration, section by section.
type Config struct {
	Logging

	// Server is a plain value struct, so the section is always there.
	Server ServerConfig `toml:"server"`

	// TLS is an optional section: nil until something under it is supplied.
	TLS *TLSConfig `toml:"tls"`

	// Metrics is where metrics go. Exactly one of Metrics and Tracing is configured, which only
	// a pointer can express.
	Metrics *MetricsConfig `toml:"metrics" validate:"required_without=Tracing,excluded_with=Tracing"`

	// Tracing is where spans go, and the alternative to Metrics.
	Tracing *TracingConfig `toml:"tracing" validate:"required_without=Metrics,excluded_with=Metrics"`
}
