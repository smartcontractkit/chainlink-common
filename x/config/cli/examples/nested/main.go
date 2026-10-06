package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/appconfig"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var cfg = appconfig.Config{
	// LogLevel is on an embedded struct
	LogLevel: "info",
	Server:   appconfig.ServerConfig{Host: "127.0.0.1"},
}

func main() {
	root := &cobra.Command{
		Use:   "nested",
		Short: "embedded, value, and pointer sections",
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("log-level   = %s\n", cfg.LogLevel)
			fmt.Printf("server.host = %s\n", cfg.Server.Host)
			if cfg.Metrics == nil {
				fmt.Println("metrics     = <nil>")
			} else {
				fmt.Printf("metrics     = endpoint=%s\n", cfg.Metrics.Endpoint)
			}

			return nil
		},
	}
	b, err := cli.New(cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err == nil {
		err = b.Register(root, &cfg)
	}

	if err != nil {
		log.Fatal(err)
	}

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
