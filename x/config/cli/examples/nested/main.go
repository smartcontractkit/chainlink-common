// Command nested shows how nested sections behave, and when a section has to be a pointer.
//
//   - TLS is a *pointer* struct with no rule of its own. Nothing under it set means it stays nil,
//     which is how a program tells "TLS not configured" from "TLS configured to its zero value".
//     Setting any one of its keys allocates it, and only then does its own `required` apply.
//   - Metrics and Tracing are mutually exclusive pointer sections. Cross-field rules
//     (required_without, excluded_with, ...) need absence to be representable, so these have to
//     be pointers; registering a value struct with such a rule is rejected up front.
//
// The structs live in the settings package, whose doc comments are the flags' help text - see
// settings/doccomments_gen.go for what the generator makes of them.
//
// See README.md for commands that show each case.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var cfg = settings.Config{
	Logging: settings.Logging{LogLevel: "info"},
	Server: settings.ServerConfig{
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
			fmt.Printf("log-level   = %s\n", cfg.LogLevel)
			fmt.Printf("server.host = %s\n", cfg.Server.Host)
			fmt.Printf("server.port = %d\n", cfg.Server.Port)
			fmt.Printf("server.idle = %s\n", cfg.Server.Idle)
			fmt.Printf("tls         = %s\n", describeTLS(cfg.TLS))
			fmt.Printf("metrics     = %s\n", describeMetrics(cfg.Metrics))
			fmt.Printf("tracing     = %s\n", describeTracing(cfg.Tracing))
			return nil
		},
	}
	b, err := cli.New(root, cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err != nil {
		log.Fatal(err)
	}
	if err := b.Register(root, "", &cfg); err != nil {
		log.Fatal(err)
	}

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func describeTLS(c *settings.TLSConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return fmt.Sprintf("cert-file=%s key-file=%s", c.CertFile, c.KeyFile)
}

func describeMetrics(c *settings.MetricsConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return "endpoint=" + c.Endpoint.String()
}

func describeTracing(c *settings.TracingConfig) string {
	if c == nil {
		return "<nil> (not configured)"
	}
	return "collector=" + c.Collector.String()
}
