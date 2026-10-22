package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var cfg = settings.Config{
	Host:    "localhost",
	Port:    8080,
	Timeout: *config.MustNewDuration(5 * time.Second),
	Tags:    []string{"default"},
}

func main() {
	root := &cobra.Command{
		Use:   "simple",
		Short: "one config struct on one command",
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("host    = %s\n", cfg.Host)
			fmt.Printf("port    = %d\n", cfg.Port)
			fmt.Printf("timeout = %s\n", cfg.Timeout)
			fmt.Printf("tags    = %v\n", cfg.Tags)
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
