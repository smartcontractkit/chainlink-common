package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/namespaced/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var (
	database = settings.DatabaseConfig{
		URL:         "postgres://localhost:5432/app",
		DialTimeout: *config.MustNewDuration(3 * time.Second),
	}
	evm = settings.EVMConfig{
		URL:     "http://localhost:8545",
		ChainID: 1,
	}
)

func main() {
	root := &cobra.Command{
		Use:   "namespaced",
		Short: "two config structs on one command",
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("database.url          = %s\n", database.URL)
			fmt.Printf("database.dial-timeout = %s\n", database.DialTimeout)
			fmt.Printf("evm.url               = %s\n", evm.URL)
			fmt.Printf("evm.chain-id          = %d\n", evm.ChainID)
			return nil
		},
	}

	b, err := cli.New(cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err == nil {
		err = b.RegisterInNamespace(root, "Database", &database)
	}

	if err == nil {
		err = b.RegisterInNamespace(root, "EVM", &evm)
	}

	if err != nil {
		log.Fatal(err)
	}

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
