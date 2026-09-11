// Command gen generates the DocComments for this example. The upstream package is never named
// here: Discover reaches it through Config's embedded and nested fields, and
// GenerateDocCommentFiles writes a file into both directories it found.
package main

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing/examples/nested"
)

func main() {
	docs, err := commentparsing.Discover(&nested.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := docs.GenerateDocCommentFiles(); err != nil {
		log.Fatal(err)
	}
	log.Println("documented", docs.Packages())
}
