package commentparsing

import (
	"sort"
	"strconv"
	"strings"
)

// selfImportPath is this package, imported by the generated code for the FieldDoc type.
const selfImportPath = "github.com/smartcontractkit/chainlink-common/x/config/commentparsing"

// docCommentsFile renders one package's DocComments methods. The package must be a local one:
// only those carry a package clause, which comes from parsed source.
func docCommentsFile(pkg Package) string {
	qualifier := "commentparsing."
	var buf strings.Builder
	buf.WriteString("package " + pkg.Name + "\n\n")
	if pkg.ImportPath == selfImportPath {
		qualifier = ""
	} else {
		// Written rather than left to goimports, which cannot resolve a package in another
		// module from the identifier alone.
		buf.WriteString("import " + strconv.Quote(selfImportPath) + "\n\n")
	}

	for _, typ := range pkg.Types {
		buf.WriteString("func (" + typ.Name + ") DocComments() map[string]" + qualifier + "FieldDoc {\n" +
			"\treturn map[string]" + qualifier + "FieldDoc{\n")
		for _, name := range sortedFieldNames(typ.Fields) {
			comment := typ.Fields[name].Comment
			if comment == "" {
				continue
			}
			buf.WriteString("\t\t" + strconv.Quote(name) + ": {Comment: " + strconv.Quote(comment) + "},\n")
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
