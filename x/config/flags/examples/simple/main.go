// Command simple binds one config struct to one command, which is the whole of the API for most
// binaries. It prints what each setting resolved to and where that value came from, so the
// precedence chain (flag > env > config file > compiled-in default) is visible in the output.
//
// See README.md for the commands that walk that chain.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/flags"
)

type Config struct {
	Host    string          `toml:"host" usage:"host to dial"`
	Port    int             `toml:"port" usage:"port to dial"`
	Timeout config.Duration `toml:"timeout" usage:"per-request timeout"`
	Tags    []string        `toml:"tags" usage:"labels to attach to every request"`
}

// cfg carries the compiled-in defaults, which is the bottom of the precedence chain: a field
// nobody configures is left exactly as constructed here.
var cfg = Config{
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
	root.PersistentFlags().String("config", "", "path to a config file")

	// "APP" is the env var prefix, so host is also settable as APP_HOST.
	if err := flags.RegisterCommandFlags(root, &cfg, flags.DefaultTOMLOptions("APP")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
