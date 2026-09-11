package commentparsing

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GeneratedFileName is where the DocComments methods are written. It is skipped when parsing, so
// a regeneration reads the same sources as the first run.
const GeneratedFileName = "doccomments_gen.go"

// docCommentsMethod is the name generation claims on every documented type, so a type that
// declares a field or method of its own by that name cannot be documented.
const docCommentsMethod = "DocComments"

// Package is one package of a config tree, as a [Generator] receives it.
type Package struct {
	// ImportPath matches what reflection reports as a type's package. Empty for a package parsed
	// without one.
	ImportPath string

	// Name is the package clause.
	Name string

	// Dir is where the package's generated files belong. [ParseDir] reports the directory it
	// read; [Discover] rewrites it relative to the run directory.
	Dir string

	// Types are sorted by name, so regenerating an unchanged package produces an unchanged file.
	Types []Type

	// declaresDocComments holds the types whose own source declares the method generation
	// claims, keyed by type name. [GeneratedFileName] is not read, so a method from an earlier
	// run is not one of them.
	declaresDocComments map[string]bool
}

// Type is one struct declaration and the documentation of its fields.
type Type struct {
	Name   string
	Fields map[string]FieldDoc
}

// ParseDir reads the doc comments of every struct type declared in the package rooted at dir. Test
// files are excluded.
//
// Only the files the active build selects are read, as [go/build.Context.MatchFile] decides it
// from GOOS, GOARCH and the constraints in each file, so a package declaring one type across
// config_linux.go and config_windows.go yields the one the build in hand compiles rather than
// both. Since what generation writes is untagged, a config type that exists on one platform only
// is not supported: generating on another misses it, and generating on its own leaves a receiver
// the others cannot compile. A type behind a -tags build tag is not visible here either.
func ParseDir(dir string) (*Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	pkg := &Package{Dir: dir, declaresDocComments: map[string]bool{}}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || name == GeneratedFileName {
			continue
		}

		switch matches, err := build.Default.MatchFile(dir, name); {
		case err != nil:
			return nil, fmt.Errorf("%s: %w", filepath.Join(dir, name), err)
		case !matches:
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
		for _, receiver := range docCommentsReceivers(file) {
			pkg.declaresDocComments[receiver] = true
		}
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("no Go package in %s", dir)
	}
	sort.Slice(pkg.Types, func(i, j int) bool { return pkg.Types[i].Name < pkg.Types[j].Name })
	return pkg, nil
}

// Type finds a type by name. A miss means the type is not declared in the package, which differs
// from having no documentation.
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

			if typeSpec.TypeParams != nil {
				continue
			}
			types = append(types, Type{Name: typeSpec.Name.Name, Fields: fieldsIn(structType)})
		}
	}
	return types
}

// docCommentsReceivers names the types in the file that declare the method generation claims.
func docCommentsReceivers(file *ast.File) []string {
	var receivers []string
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Name.Name != docCommentsMethod {
			continue
		}
		for _, field := range funcDecl.Recv.List {
			// The receiver's own name, where it has one, is not wanted - the type it is
			// declared on is, which is what a field with no name reports.
			receivers = append(receivers, fieldNames(&ast.Field{Type: field.Type})...)
		}
	}
	return receivers
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

// fieldNames returns the Go names a field declares. For an embedded field that is the bare name
// of its type, which is what reflect reports.
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

// commentText prefers the comment above a field to the one trailing it, and keeps every line.
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
