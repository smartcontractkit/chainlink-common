// Command gen generates the DocComments for this example. The upstream package is never named
// here: Run reaches it through Config's embedded and nested fields.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&nested.Config{}},
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
