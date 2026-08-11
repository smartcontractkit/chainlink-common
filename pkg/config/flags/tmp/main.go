package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/config/flags"
	"github.com/spf13/cobra"
)

// --- CONFIG STRUCTS ---

type ChainConfig struct {
	ID       uint32 `toml:"id" json:"id" usage:"Target chain ID selector"`
	RPC      string `toml:"rpc" json:"rpc" usage:"RPC endpoint URL" validate:"required" example:"'https://mainnet.infura.io/v3/API_KEY'"`
	Contract string `toml:"contract" json:"contract" usage:"Core contract address" validate:"required" example:"'0xYourContractAddress'"`
}

type SystemConfig struct {
	Env      string `toml:"env" json:"env" usage:"Environment profile selector (dev, prod)"`
	LogLevel string `toml:"log_level" json:"log_level" usage:"Logging severity"`
	Metrics  bool   `toml:"metrics" json:"metrics" usage:"Enable metrics collection"`
}

type CommonConfig struct {
	Chain  ChainConfig  `toml:"chain" json:"chain"`
	System SystemConfig `toml:"system" json:"system"`
}

type FooSettings struct {
	Timeout *config.Duration `toml:"timeout" json:"timeout" usage:"Foo operation timeout"`
}

type BarSettings struct {
	Retries int `toml:"retries" json:"retries" usage:"Bar max retries"`
	// non-pointer config.Duration, to exercise the value form (UnmarshalText has a
	// pointer receiver, so this decodes differently from FooSettings.Timeout)
	Backoff config.Duration `toml:"backoff" json:"backoff" usage:"Bar retry backoff"`
}

// --- PROFILE DICTIONARIES ---

var ChainProfiles = map[uint32]CommonConfig{
	1: {
		Chain: ChainConfig{
			RPC:      "https://eth.blastapi.io",
			Contract: "0x1111111111111111111111111111111111111111",
		},
	},
	137: {
		Chain: ChainConfig{
			RPC:      "https://polygon.blastapi.io",
			Contract: "0x2222222222222222222222222222222222222222",
		},
	},
}

var SystemProfiles = map[string]CommonConfig{
	"dev": {
		System: SystemConfig{
			LogLevel: "debug",
			Metrics:  false,
		},
	},
	"prod": {
		System: SystemConfig{
			LogLevel: "info",
			Metrics:  true,
		},
	},
}

// --- STATE DECLARATIONS ---

var (
	configFile string

	AppConfig = CommonConfig{
		Chain:  ChainConfig{ID: 1},
		System: SystemConfig{Env: "dev"},
	}

	fooSettings = FooSettings{Timeout: config.MustNewDuration(5 * time.Second)}
	barSettings = BarSettings{Retries: 3, Backoff: *config.MustNewDuration(2 * time.Second)}
)

// --- COMMAND DECLARATIONS ---

var rootCmd = &cobra.Command{
	Use: "myapp",
}

var fooCmd = &cobra.Command{
	Use:   "foo",
	Short: "Foo-specific settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("FOO")
		return printConfig(AppConfig, fooSettings)
	},
}

var barCmd = &cobra.Command{
	Use:   "bar",
	Short: "Bar-specific settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("BAR")
		return printConfig(AppConfig, barSettings)
	},
}

func printConfig(common CommonConfig, settings any) error {
	out, err := json.MarshalIndent(struct {
		Common CommonConfig `json:"common"`
		Cmd    any          `json:"cmd"`
	}{common, settings}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file")

	// 1. Register base persistent flags on root with env prefixes ("CRE", "CL")
	if err := flags.RegisterCommandFlags(rootCmd, &AppConfig, flags.DefaultTOMLOptions("CRE", "CL")); err != nil {
		panic(err)
	}

	// 2. Attach profile maps to root command
	if err := flags.RegisterProfile(rootCmd, "Chain.ID", ChainProfiles, flags.DefaultTOMLOptions("CRE", "CL")); err != nil {
		panic(err)
	}
	if err := flags.RegisterProfile(rootCmd, "System.Env", SystemProfiles, flags.DefaultTOMLOptions("CRE", "CL")); err != nil {
		panic(err)
	}

	// 3. Register subcommand flags (inherits "CRE" / "CL" prefixes automatically)
	if err := flags.RegisterSubcommandFlags(fooCmd, "foo", &fooSettings, flags.DefaultTOMLOptions()); err != nil {
		panic(err)
	}
	if err := flags.RegisterSubcommandFlags(barCmd, "bar", &barSettings, flags.DefaultTOMLOptions()); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(fooCmd)
	rootCmd.AddCommand(barCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
