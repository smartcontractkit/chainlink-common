package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/namespaced/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

// Generate the comments for our settings so that the CLI can use them.
func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Markup:      tomlmarkup.New(),
		Roots:       []any{&settings.DatabaseConfig{}, &settings.EVMConfig{}},
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/namespaced/settings/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
