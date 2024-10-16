package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unicode"
)

type GoStructReader struct {
	IncludeType func(name string) bool
}

func (g *GoStructReader) Read(src string) (map[string]Struct, string, []string, error) {
	fset := token.NewFileSet()

	// Parse the source code string
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return nil, "", nil, err
	}

	return g.gatherStructs(node, fset, src), node.Name.Name, g.gatherImports(node), nil
}

func (g *GoStructReader) gatherStructs(node *ast.File, fset *token.FileSet, src string) map[string]Struct {
	generatedStructs := map[string]Struct{}
	for _, decl := range node.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		for _, spec := range gd.Specs {
			if strct := g.getStructFromSpec(spec, fset, src); strct != nil {
				generatedStructs[strct.Name] = *strct
			}
		}
	}
	return generatedStructs
}

func (g *GoStructReader) getStructFromSpec(spec ast.Spec, fset *token.FileSet, src string) *Struct {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}

	name := ts.Name.Name
	if !g.IncludeType(name) {
		return nil
	}

	switch declType := ts.Type.(type) {
	case *ast.StructType:
		return g.structFromGoStruct(name, declType, fset, src)
	case *ast.MapType, *ast.Ident:
		return &Struct{Name: name}
	default:
		return nil
	}
}

func (g *GoStructReader) structFromGoStruct(name string, structType *ast.StructType, fset *token.FileSet, src string) *Struct {
	s := Struct{
		Name:    strings.TrimSpace(name),
		Outputs: map[string]Field{},
	}

	for _, field := range structType.Fields.List {
		start := fset.Position(field.Type.Pos()).Offset
		end := fset.Position(field.Type.End()).Offset
		typeStr := src[start:end]
		if typeStr == "interface{}" {
			typeStr = "any"
		}

		f := Field{
			Type:       typeStr,
			ConfigName: g.configName(field),
			SkipCap:    !g.IncludeType(typeStr),
		}

		for strings.HasPrefix(f.Type, "[]") {
			f.NumSlice++
			f.Type = f.Type[2:]
		}

		f.Type = strings.TrimPrefix(f.Type, "*")
		t := f.Type
		for t[0] == '*' {
			t = t[1:]
		}

		importLoc := strings.Index(t, ".")
		if importLoc != -1 {
			t = t[importLoc+1:]
		}
		f.IsPrimitive = unicode.IsLower(rune(t[0]))
		s.Outputs[field.Names[0].Name] = f
	}

	return &s
}

func (g *GoStructReader) configName(field *ast.Field) string {
	defaultName := field.Names[0].Name
	if field.Tag == nil {
		return defaultName
	}

	// TODO gather imports that are required from here instead
	// TODO tags
	// This is safe because the generator used to create the structs from jsonschema
	// will always have json tag if there's tags on the field, per configuration.
	// The substring removes the quotes around that tag.
	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
	jsonTag := tag.Get("json")
	if jsonTag != "" {
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName != "" {
			return jsonName
		}
	}

	return defaultName
}

func (g *GoStructReader) gatherImports(node *ast.File) []string {
	// TODO this should really look at what's used before the . on fields
	var imports []string
	for _, imp := range node.Imports {
		switch imp.Path.Value {
		case `"reflect"`, `"fmt"`, `"encoding/json"`, `"regexp"`:
		default:
			importStr := imp.Path.Value
			if imp.Name != nil {
				importStr = imp.Name.Name + " " + importStr
			}
			imports = append(imports, importStr)
		}
	}

	return imports
}
