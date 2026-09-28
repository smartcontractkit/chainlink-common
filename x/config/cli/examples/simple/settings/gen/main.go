// Command gen writes the DocComments method of the simple example's config struct, which is
// where its flags' help text comes from.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Markup:      tomlmarkup.New(),
		Roots:       []any{&settings.Config{}},
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
