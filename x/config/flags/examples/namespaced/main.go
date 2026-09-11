// Command namespaced registers two independent config structs on one command. Both own a field
// called URL; a namespace is what keeps them apart, in the flag set, in the env vars and in the
// config file alike.
//
// A namespace also groups a dependency's settings, so a package can ship its own config struct
// and the binary decides where in the config tree it lands.
//
// See README.md for the commands that walk the fallback chain.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/flags"
)

// DatabaseConfig stands in for a struct owned by a database package.
type DatabaseConfig struct {
	URL         string          `toml:"url" usage:"connection string"`
	MaxOpen     int             `toml:"max-open" usage:"maximum open connections"`
	DialTimeout config.Duration `toml:"dial-timeout" usage:"how long to wait for a connection"`
}

// EVMConfig stands in for a struct owned by a chain package. Its URL is a different setting from
// the database's, and only the namespace says so.
type EVMConfig struct {
	URL     string `toml:"url" usage:"RPC endpoint"`
	ChainID uint64 `toml:"chain-id" usage:"chain selector"`
}

var (
	database = DatabaseConfig{
		URL:         "postgres://localhost:5432/app",
		MaxOpen:     10,
		DialTimeout: *config.MustNewDuration(3 * time.Second),
	}
	evm = EVMConfig{
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
	root.PersistentFlags().String("config", "", "path to a config file")

	dbOpts := flags.DefaultTOMLOptions("APP")
	dbOpts.Namespace = "database"
	if err := flags.RegisterCommandFlags(root, &database, dbOpts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	evmOpts := flags.DefaultTOMLOptions("APP")
	evmOpts.Namespace = "evm"
	if err := flags.RegisterCommandFlags(root, &evm, evmOpts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
