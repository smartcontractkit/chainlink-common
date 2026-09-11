package commentparsing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// GeneratedFileName is where [WriteFile] puts its output, and is skipped when parsing so a
// regeneration reads the same sources as the first run.
const GeneratedFileName = "doccomments_gen.go"

// Package is one parsed Go package. Types are sorted by name so [Generate] emits them in a stable
// order rather than the filesystem's.
type Package struct {
	// Name is the package clause, which the generated file needs because it is written beside
	// the source it was read from.
	Name string

	Types []Type
}

// Type is one struct declaration and the documentation of its fields.
type Type struct {
	Name string

	// Fields is keyed by Go field name rather than by config key, because the key depends on
	// which tag convention the caller follows and the Go name does not.
	Fields map[string]FieldDoc
}

// ParseDir reads the documentation of every struct type declared in the package rooted at dir.
//
// Test files are excluded, since a config struct declared in one is not part of the package a
// consumer imports.
func ParseDir(dir string) (*Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	pkg := &Package{}
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
			if fields, ok := documentedFields(structType); ok {
				types = append(types, Type{Name: typeSpec.Name.Name, Fields: fields})
			}
		}
	}
	return types
}

// documentedFields reports whether a struct takes part in configuration at all, and its fields if
// so. A struct with no comment and no tag anywhere is some internal type, and indexing it would
// have a caller's lookup succeed on types no config file ever names.
//
// An embedded field alone qualifies, even when the struct adds nothing of its own. Skipping it
// would leave it promoting the embedded type's DocComments and answering for fields that are not
// its own - see [DocCommenter].
func documentedFields(structType *ast.StructType) (map[string]FieldDoc, bool) {
	fields := make(map[string]FieldDoc)
	qualifies := false

	for _, field := range structType.Fields.List {
		tag := structTag(field)
		doc := FieldDoc{
			Comment:  commentText(field),
			Validate: tag.Get("validate"),
		}
		if len(field.Names) == 0 || doc.Comment != "" || tag != "" {
			qualifies = true
		}

		for _, name := range fieldNames(field) {
			fields[name] = doc
		}
	}

	if !qualifies {
		return nil, false
	}
	return fields, true
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

// structTag unquotes a field's tag literal. An unparseable literal yields no tag rather than an
// error, because the compiler will reject it far more clearly than this parser could.
func structTag(field *ast.Field) reflect.StructTag {
	if field.Tag == nil {
		return ""
	}
	unquoted, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}
	return reflect.StructTag(unquoted)
}
