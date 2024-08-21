package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unicode"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type GeneratedInfo struct {
	Package        string
	Config         Struct
	Input          *Struct
	Types          map[string]Struct
	CapabilityType capabilities.CapabilityType
	BaseName       string
	RootOutput     string
	RootNumSlice   int
	ExtraImports   []string
	ID             *string
	FullPackage    string
}

func (g GeneratedInfo) RootType() Struct {
	if r, ok := g.Types[g.RootOutput]; ok {
		return r
	}
	return Struct{
		Name: g.RootOutput,
		Ref:  &g.RootOutput,
	}
}

func generatedInfoFromSrc(src string, capID *string, typeInfo TypeInfo) (GeneratedInfo, error) {
	fset := token.NewFileSet()

	// Parse the source code string
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return GeneratedInfo{}, err
	}
	pkg := node.Name.Name

	generatedStructs := map[string]Struct{}
	var extraImports []string
	ast.Inspect(node, func(n ast.Node) bool {
		return inspectNode(n, fset, src, generatedStructs, &extraImports)
	})

	root := generatedStructs[typeInfo.RootType]
	input, config := extractInputAndConfig(generatedStructs, typeInfo, root)

	output := root.Outputs["Outputs"]

	pkgParts := strings.Split(typeInfo.SchemaID, "/")
	// skip http(s):// and drop the last part
	fullPkg := strings.Join(pkgParts[2:len(pkgParts)-1], "/")
	return GeneratedInfo{
		Package:        pkg,
		Config:         config,
		Types:          generatedStructs,
		RootOutput:     output.Type,
		RootNumSlice:   output.NumSlice,
		BaseName:       typeInfo.RootType,
		CapabilityType: typeInfo.CapabilityType,
		Input:          input,
		ExtraImports:   extraImports,
		ID:             capID,
		FullPackage:    fullPkg,
	}, nil
}

func extractInputAndConfig(generatedStructs map[string]Struct, typeInfo TypeInfo, root Struct) (*Struct, Struct) {
	delete(generatedStructs, typeInfo.RootType)
	inputField, ok := root.Outputs["Inputs"]
	var input *Struct
	if ok {
		inputType := inputField.Type
		inputS, ok2 := generatedStructs[inputType]
		if ok2 {
			input = &inputS
			delete(generatedStructs, inputType)
		} else {
			input = &Struct{
				Name: lastAfterDot(inputType),
				Ref:  &inputType,
			}
		}
	}

	configType := root.Outputs["Config"].Type
	config, ok := generatedStructs[configType]
	if !ok && typeInfo.CapabilityType.IsValid() == nil {
		config = Struct{
			Name: lastAfterDot(configType),
			Ref:  &configType,
		}
	}

	for k := range generatedStructs {
		if (ok && strings.HasPrefix(k, configType)) || (input != nil && strings.HasPrefix(k, input.Name)) {
			delete(generatedStructs, k)
		}
	}
	return input, config
}

func inspectNode(n ast.Node, fset *token.FileSet, src string, rawInfo map[string]Struct, extraImports *[]string) bool {
	if ts, ok := n.(*ast.TypeSpec); ok {
		s := Struct{
			Name:    strings.TrimSpace(ts.Name.Name),
			Outputs: map[string]Field{},
		}

		if structType, ok := ts.Type.(*ast.StructType); ok {
			for _, field := range structType.Fields.List {
				start := fset.Position(field.Type.Pos()).Offset
				end := fset.Position(field.Type.End()).Offset
				typeStr := src[start:end]
				if typeStr == "interface{}" {
					typeStr = "any"
				}
				f := Field{}

				if field.Tag != nil {
					// This is safe because the generator used to create the structs from jsonschema
					// will always have json tag if there's tags on the field, per configuration.
					// The substring removes the quotes around that tag.
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					jsonTag := tag.Get("json")
					if jsonTag != "" {
						f.ConfigName = jsonTag
					}
				}

				f.Type = typeStr
				if f.ConfigName == "" {
					f.ConfigName = field.Names[0].Name
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

				f.IsPrimitive = unicode.IsLower(rune(t[0]))
				s.Outputs[field.Names[0].Name] = f
			}
		}

		// artifact used for deserializing
		if s.Name != "Plain" {
			rawInfo[ts.Name.Name] = s
		}
	} else if imp, ok := n.(*ast.ImportSpec); ok {
		switch imp.Path.Value {
		case `"reflect"`, `"fmt"`, `"encoding/json"`, `"regexp"`:
		default:
			importStr := imp.Path.Value
			if imp.Name != nil {
				importStr = imp.Name.Name + " " + importStr
			}
			*extraImports = append(*extraImports, importStr)
		}
	}
	return true
}

func lastAfterDot(s string) string {
	parts := strings.Split(s, ".")
	return parts[len(parts)-1]
}
