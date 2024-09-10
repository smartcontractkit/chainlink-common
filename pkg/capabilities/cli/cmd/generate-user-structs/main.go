package main

import (
	"flag"
)

var localPrefix = flag.String("local_prefix", "github.com/smartcontractkit", "The local prefix to use when formatting go files")
var structs = flag.String("structs", "", "Comma separated list of structs to generate capability wrappers for")
var outputFile = flag.String("output_file", "wrappers_generated.go", "File to write the generated code to")

func main() {

}
