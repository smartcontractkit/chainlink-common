// Command gen generates the DocComments for the upstream package, so a consumer in another module
// can read them without access to this source.
package main

import (
	"log"

	"example.com/upstream"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&upstream.Settings{}, &upstream.Retry{}},
		Markup:      tomlmarkup.New(),
		Tool:        "example.com/upstream/gen",
		LocalPrefix: "example.com",
	}); err != nil {
		log.Fatal(err)
	}
}
