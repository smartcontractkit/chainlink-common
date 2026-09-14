package commentparsing

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// selfImportPath is this package, imported by the code generated for the FieldDoc type. A package
// documenting itself would import itself, so that one case is emitted unqualified.
const selfImportPath = "github.com/smartcontractkit/chainlink-common/x/config/commentparsing"

// docCommentsFile renders one package's DocComments methods.
//
// The package must be a local one: only those carry a package clause, since it comes from parsed
// source and not from the import path a dependency is known by. [DocComments] skips the rest.
//
// No generated-by header and no gofmt pass here: [Run] writes through
// chainlink-common/pkg/utils/codegen, which adds the header and formats every file the same way,
// so a second one emitted here would be duplicated and a second formatting could disagree.
func docCommentsFile(pkg Package) string {
	qualifier := "commentparsing."
	var buf strings.Builder
	fmt.Fprintf(&buf, "package %s\n\n", pkg.Name)
	if pkg.ImportPath == selfImportPath {
		qualifier = ""
	} else {
		// The import is written rather than left to goimports, which cannot be relied on to
		// resolve a package in some other module from the identifier alone.
		fmt.Fprintf(&buf, "import %q\n\n", selfImportPath)
	}

	for _, typ := range pkg.Types {
		// A value receiver puts the method on both T and *T, so a consumer holding either can
		// reach it, and reflect.New is enough to call it from the type alone.
		fmt.Fprintf(&buf, "func (%s) DocComments() (string, map[string]%sFieldDoc) {\n", typ.Name, qualifier)
		fmt.Fprintf(&buf, "\treturn %q, map[string]%sFieldDoc{\n", typ.Name, qualifier)
		for _, name := range sortedFieldNames(typ.Fields) {
			comment := typ.Fields[name].Comment
			if comment == "" {
				continue // nothing compilation discarded, so nothing to carry
			}
			fmt.Fprintf(&buf, "\t\t%q: {Comment: %s},\n", name, strconv.Quote(comment))
		}
		buf.WriteString("\t}\n}\n\n")
	}
	return buf.String()
}

func sortedFieldNames(fields map[string]FieldDoc) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
