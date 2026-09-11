// Command gen generates the DocComments for this example. It names one root type and nothing
// else: Discover finds every package that type reaches, and writes a file into each one it owns.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/simple"
)

func main() {
	docs, err := commentparsing.Discover(&simple.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := docs.GenerateDocCommentFiles(); err != nil {
		log.Fatal(err)
	}
	log.Println("documented", docs.Packages())
}
