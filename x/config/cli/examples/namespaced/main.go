// Command namespaced registers two independent config structs on one command. Both own a field
// called URL; a namespace is what keeps them apart, in the flag set, in the env vars and in the
// config file alike.
//
// A namespace also groups a dependency's settings, so a package can ship its own config struct
// and the binary decides where in the config tree it lands.
//
// The structs live in the settings package, whose doc comments are the flags' help text - see
// settings/doccomments_gen.go for what the generator makes of them.
package main

import (
	"fmt"
	"os"
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
		MaxOpen:     10,
		DialTimeout: *config.MustNewDuration(3 * time.Second),
	}
	evm = settings.EVMConfig{
		URL:     "https://localhost:8545",
		ChainID: 1,
	}
)

func main() {
	root := &cobra.Command{
		Use:   "namespaced",
		Short: "two independent config structs on one command",
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("database.url          = %s\n", database.URL)
			fmt.Printf("database.max-open     = %d\n", database.MaxOpen)
			fmt.Printf("database.dial-timeout = %s\n", database.DialTimeout)
			fmt.Printf("evm.url               = %s\n", evm.URL)
			fmt.Printf("evm.chain-id          = %d\n", evm.ChainID)
			return nil
		},
	}
	b, err := cli.New(root, cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err == nil {
		err = b.Register(root, "database", &database)
	}
	if err == nil {
		err = b.Register(root, "evm", &evm)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
