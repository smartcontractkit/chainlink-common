// Command gen generates the DocComments for the upstream package, so a consumer in another module
// can read this package's documentation with no access to this source.
package main

import (
	"log"

	"example.com/upstream"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

func main() {
	docs, err := commentparsing.Discover(&upstream.Settings{})
	if err != nil {
		log.Fatal(err)
	}
	if err := docs.GenerateDocCommentFiles(); err != nil {
		log.Fatal(err)
	}
}
