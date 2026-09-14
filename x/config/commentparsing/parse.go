package commentparsing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GeneratedFileName is where the DocComments methods are written, and is skipped when parsing so a
// regeneration reads the same sources as the first run.
const GeneratedFileName = "doccomments_gen.go"

// Package is one package of a config tree, as a [Generator] receives it.
type Package struct {
	// ImportPath is what reflection reports as a type's package, so a generator keying on
	// reflect.Type can match against it. It is empty for a package parsed without one.
	ImportPath string

	// Name is the package clause, which a generator emitting Go source into Dir needs and cannot
	// derive from the path.
	Name string

	// Dir is where the package's generated files belong. [ParseDir] reports the directory it
	// read; [Discover] rewrites it relative to the run directory, which is what a generator's
	// paths are resolved against.
	Dir string

	// Types are sorted by name, so regenerating an unchanged package produces an unchanged file.
	Types []Type
}

// Type is one struct declaration and the documentation of its fields.
type Type struct {
	Name string

	// Fields is keyed by Go field name rather than by config key, because the key depends on
	// which tag convention the consumer follows and the Go name does not.
	Fields map[string]FieldDoc
}

// ParseDir reads the doc comments of every struct type declared in the package rooted at dir.
//
// Every struct is returned, with no judgement about which take part in configuration: the caller
// walking a real config value knows which types it reached, and a parser guessing from comments
// and tags would only disagree with it.
//
// Test files are excluded, since a config struct declared in one is not part of the package a
// consumer imports.
func ParseDir(dir string) (*Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	pkg := &Package{Dir: dir}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || name == GeneratedFileName {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil,
			parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		if pkg.Name == "" {
			pkg.Name = file.Name.Name
		}
		pkg.Types = append(pkg.Types, typesIn(file)...)
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("no Go package in %s", dir)
	}
	sort.Slice(pkg.Types, func(i, j int) bool { return pkg.Types[i].Name < pkg.Types[j].Name })
	return pkg, nil
}

// Type finds a type by name, because a name is all reflection reports about one - there is no
// index into the file it was declared in. A miss means the type is not part of the package a
// consumer imports, which is a different thing from having no documentation.
func (p *Package) Type(name string) (Type, bool) {
	for _, typ := range p.Types {
		if typ.Name == name {
			return typ, true
		}
	}
	return Type{}, false
}

func typesIn(file *ast.File) []Type {
	var types []Type
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			types = append(types, Type{Name: typeSpec.Name.Name, Fields: fieldsIn(structType)})
		}
	}
	return types
}

func fieldsIn(structType *ast.StructType) map[string]FieldDoc {
	fields := make(map[string]FieldDoc)
	for _, field := range structType.Fields.List {
		doc := FieldDoc{Comment: commentText(field)}
		for _, name := range fieldNames(field) {
			fields[name] = doc
		}
	}
	return fields
}

// fieldNames returns the Go names a field declares, which for an embedded field is the bare name
// of its type - the name reflect reports, so a caller can key on it without knowing whether the
// field was embedded.
func fieldNames(field *ast.Field) []string {
	if len(field.Names) > 0 {
		names := make([]string, 0, len(field.Names))
		for _, ident := range field.Names {
			names = append(names, ident.Name)
		}
		return names
	}

	expr := field.Type
	for {
		switch typed := expr.(type) {
		case *ast.StarExpr:
			expr = typed.X
		case *ast.SelectorExpr:
			return []string{typed.Sel.Name}
		case *ast.IndexExpr:
			expr = typed.X
		case *ast.IndexListExpr:
			expr = typed.X
		case *ast.Ident:
			return []string{typed.Name}
		default:
			return nil
		}
	}
}

// commentText prefers the comment above a field to the one trailing it, and keeps every line: a
// synopsis drops exactly the caveats a config reference is consulted for.
func commentText(field *ast.Field) string {
	group := field.Doc
	if group == nil {
		group = field.Comment
	}
	if group == nil {
		return ""
	}
	return strings.TrimSuffix(group.Text(), "\n")
}
