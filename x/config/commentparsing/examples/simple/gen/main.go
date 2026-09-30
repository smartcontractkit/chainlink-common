// Command gen generates the DocComments for this example. It names one root type; Run discovers
// every package that type reaches.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&simple.Config{}},
		Markup:      tomlmarkup.New(),
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
