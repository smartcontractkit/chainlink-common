// Command simple binds one config struct to one command, which is the whole of the API for most
// binaries. It prints what each setting resolved to, so the precedence chain (flag > env > config
// file > compiled-in default) is visible by setting one value several ways.
//
// The struct itself lives in the settings package, whose doc comments are the flags' help text -
// see settings/doccomments_gen.go for what the generator makes of them.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/cli"
	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

// cfg carries the compiled-in defaults, which is the bottom of the precedence chain: a field
// nobody configures is left exactly as constructed here.
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
	// "APP" is the env var prefix, so host is also settable as APP_HOST.
	b, err := cli.New(root, cli.Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP"}})
	if err == nil {
		err = b.Register(root, "", &cfg)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
