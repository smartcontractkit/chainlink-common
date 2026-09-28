// Command gen writes the DocComments methods of the namespaced example's config structs. Both
// are roots: a consumer reaches either one on its own, so neither is found by walking the other.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/namespaced/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

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
