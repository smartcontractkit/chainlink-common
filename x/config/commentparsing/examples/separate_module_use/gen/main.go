// Command gen generates the DocComments for this module and reads back the dependency's.
//
// Discover writes a file only into the directory this module owns. The dependency's types are not
// skipped, they are resolved from the DocComments compiled into it - nothing here can reach that
// source, and could not write to the read-only module cache if it tried.
package main

import (
	"log"
	"reflect"

	"example.com/consumer"
	"example.com/upstream"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

func main() {
	docs, err := commentparsing.Discover(&consumer.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := docs.GenerateDocCommentFiles(); err != nil {
		log.Fatal(err)
	}
	log.Println("documented", docs.Packages())

	// Proof that it crossed the boundary: this comment was written in another module's source and
	// arrived here compiled.
	log.Println("from the dependency:", docs.Field(reflect.TypeFor[upstream.Retry](), "MaxAttempts").Comment)
}
