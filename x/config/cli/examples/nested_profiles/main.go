package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested_profiles/appconfig"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var chains = map[uint64]func(*appconfig.ChainConfig){
	1: func(c *appconfig.ChainConfig) {
		c.RPC = "https://ethereum.example.com"
		c.BlockTime = *config.MustNewDuration(12 * time.Second)
	},
	// No default RPC, so running chain 10 needs one from a source.
	10: func(c *appconfig.ChainConfig) {
		c.BlockTime = *config.MustNewDuration(2 * time.Second)
	},
	137: func(c *appconfig.ChainConfig) {
		c.RPC = "https://otherchain.example.com"
		c.BlockTime = *config.MustNewDuration(2 * time.Second)
		c.Finality = 128
	},
}

// Chains 1 and 10 keep this default finality; 137 replaces it.
var cfg = appconfig.Config{LogLevel: "info", Chain: appconfig.ChainConfig{Finality: 64}}

func main() {
	root := &cobra.Command{
		Use:   "nested_profiles",
		Short: "built-in defaults for a section, picked by one of its fields",
		RunE: func(*cobra.Command, []string) error {
			fmt.Printf("log-level        = %s\n", cfg.LogLevel)
			fmt.Printf("chain.id         = %d\n", cfg.Chain.ID)
			fmt.Printf("chain.rpc        = %s\n", cfg.Chain.RPC)
			fmt.Printf("chain.block-time = %s\n", cfg.Chain.BlockTime)
			fmt.Printf("chain.finality   = %d\n", cfg.Chain.Finality)
			return nil
		},
	}

	b, err := cli.New(cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err == nil {
		err = b.Register(root, &cfg, cli.ProfileWithSelector(
			func(c *appconfig.Config) *appconfig.ChainConfig { return &c.Chain },
			func(c *appconfig.ChainConfig) *uint64 { return &c.ID },
			chains))
	}

	if err != nil {
		log.Fatal(err)
	}

	if err = root.Execute(); err != nil {
		log.Fatal(err)
	}
}
