// Command nested shows how nested sections behave, and when a section has to be a pointer.
//
//   - Server is a plain value struct. Its Host is `validate:"required"`, which works fine on a
//     value struct: a *leaf* being required says nothing about whether its section is present.
//   - TLS is a *pointer* struct with no rule of its own. Nothing under it set means it stays nil,
//     which is how a program tells "TLS not configured" from "TLS configured to its zero value".
//     Setting any one of its keys allocates it, and only then does its own `required` apply.
//   - Metrics and Tracing are mutually exclusive pointer sections. Cross-field rules
//     (required_without, excluded_with, ...) need absence to be representable, so these have to
//     be pointers; registering a value struct with such a rule is rejected up front.
//
// See README.md for the commands that walk each of those cases.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/flags"
)

type ServerConfig struct {
	// Required, and not a pointer: the section always exists, this leaf inside it must be filled.
	Host string          `toml:"host" usage:"address to listen on" validate:"required"`
	Port int             `toml:"port" usage:"port to listen on"`
	Idle config.Duration `toml:"idle" usage:"idle connection timeout"`
}

type TLSConfig struct {
	CertFile string `toml:"cert-file" usage:"PEM certificate path" validate:"required"`
	KeyFile  string `toml:"key-file" usage:"PEM private key path" validate:"required"`
}

type MetricsConfig struct {
	Endpoint *config.URL `toml:"endpoint" usage:"where to push metrics" validate:"required"`
}

type TracingConfig struct {
	Collector *config.URL `toml:"collector" usage:"where to send spans" validate:"required"`
}

type Config struct {
	Server ServerConfig `toml:"server"`

	// Optional section: nil until something under it is supplied.
	TLS *TLSConfig `toml:"tls"`

	// Exactly one of these, which only a pointer can express.
	Metrics *MetricsConfig `toml:"metrics" validate:"required_without=Tracing,excluded_with=Tracing"`
	Tracing *TracingConfig `toml:"tracing" validate:"required_without=Metrics,excluded_with=Metrics"`
}

var cfg = Config{
	Server: ServerConfig{
		Host: "127.0.0.1",
		Port: 8080,
		Idle: *config.MustNewDuration(90 * time.Second),
	},
}

func main() {
	root := &cobra.Command{
		Use:   "nested",
		Short: "nested sections, pointer and value",
		// A rejected config is the point of several of the runs below; the usage dump would bury
		// the message that explains why.
		SilenceUsage: true,
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("server.host = %s\n", cfg.Server.Host)
			fmt.Printf("server.port = %d\n", cfg.Server.Port)
			fmt.Printf("server.idle = %s\n", cfg.Server.Idle)
			fmt.Printf("tls         = %s\n", describeTLS(cfg.TLS))
			fmt.Printf("metrics     = %s\n", describeMetrics(cfg.Metrics))
			fmt.Printf("tracing     = %s\n", describeTracing(cfg.Tracing))
			return nil
		},
	}
	root.PersistentFlags().String("config", "", "path to a config file")

	if err := flags.RegisterCommandFlags(root, &cfg, flags.DefaultTOMLOptions("APP")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func describeTLS(c *TLSConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return fmt.Sprintf("cert-file=%s key-file=%s", c.CertFile, c.KeyFile)
}

func describeMetrics(c *MetricsConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return "endpoint=" + c.Endpoint.String()
}

func describeTracing(c *TracingConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return "collector=" + c.Collector.String()
}
