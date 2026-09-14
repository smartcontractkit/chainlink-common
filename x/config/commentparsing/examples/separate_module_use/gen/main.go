// Command gen generates the DocComments for this module, and shows a second generator reading the
// same discovered packages.
//
// Run writes a file only into the directory this module owns. The dependency's types are not
// skipped, they are resolved from the DocComments compiled into it - nothing here can reach that
// source, and could not write to the read-only module cache if it tried.
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"example.com/consumer"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
)

func main() {
	if err := commentparsing.Run(commentparsing.RunArgs{
		Roots:       []any{&consumer.Config{}},
		Tool:        "example.com/consumer/gen",
		LocalPrefix: "example.com",
	}, reference); err != nil {
		log.Fatal(err)
	}
}

// reference is a second generator, fed the same packages Run discovered. The dependency's comments
// arrive here having crossed a module boundary compiled, which is the whole point of the
// arrangement.
func reference(pkgs []commentparsing.Package) (map[string]string, error) {
	var out strings.Builder
	for _, pkg := range pkgs {
		fmt.Fprintf(&out, "# %s\n", pkg.ImportPath)
		for _, typ := range pkg.Types {
			fmt.Fprintf(&out, "## %s\n", typ.Name)

			// Field order is sorted rather than taken from the map, so regenerating an unchanged
			// package writes the same bytes and a CI diff means the comments actually changed.
			names := make([]string, 0, len(typ.Fields))
			for name := range typ.Fields {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				if comment := typ.Fields[name].Comment; comment != "" {
					fmt.Fprintf(&out, "- %s: %s\n", name, strings.ReplaceAll(comment, "\n", " "))
				}
			}
		}
	}
	return map[string]string{"REFERENCE.md": out.String()}, nil
}
