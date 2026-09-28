// Command gen writes the DocComments methods of the nested example's config structs. One root is
// enough: Run follows Config's fields into every section, pointers included.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Markup:      tomlmarkup.New(),
		Roots:       []any{&settings.Config{}},
		Tool:        "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/settings/gen",
		LocalPrefix: "github.com/smartcontractkit",
	}); err != nil {
		log.Fatal(err)
	}
}
