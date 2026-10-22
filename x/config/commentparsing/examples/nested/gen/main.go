// Command gen generates the DocComments for this example. The upstream package is never named
// here: Run reaches it through Config's embedded and nested fields.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&nested.Config{}},
		Markup:      tomlmarkup.New(),
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
