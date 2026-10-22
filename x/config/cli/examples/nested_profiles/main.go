package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested_profiles/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var chains = map[uint64]func(*settings.ChainConfig){
	1: func(c *settings.ChainConfig) {
		c.RPC = "https://ethereum.example.com"
		c.BlockTime = *config.MustNewDuration(12 * time.Second)
	},
	// No default RPC, so running chain 10 needs one from a source.
	10: func(c *settings.ChainConfig) {
		c.BlockTime = *config.MustNewDuration(2 * time.Second)
	},
	137: func(c *settings.ChainConfig) {
		c.RPC = "https://othrechain.example.com"
		c.BlockTime = *config.MustNewDuration(2 * time.Second)
		c.Finality = 128
	},
}

// Chains 1 and 10 keep this default finality; 137 replaces it.
var cfg = settings.Config{LogLevel: "info", Chain: settings.ChainConfig{Finality: 64}}

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
			func(c *settings.Config) *settings.ChainConfig { return &c.Chain },
			func(c *settings.ChainConfig) *uint64 { return &c.ID },
			chains))
	}

	if err != nil {
		log.Fatal(err)
	}

	if err = root.Execute(); err != nil {
		log.Fatal(err)
	}
}
