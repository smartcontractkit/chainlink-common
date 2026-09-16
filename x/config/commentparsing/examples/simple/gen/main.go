// Command gen generates the DocComments for this example. It names one root type; Run discovers
// every package that type reaches.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&simple.Config{}},
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
