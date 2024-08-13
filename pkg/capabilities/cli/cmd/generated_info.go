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
}

func (g GeneratedInfo) RootType() Struct {
	return g.Types[g.RootOutput]
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

	return GeneratedInfo{
		Package:        pkg,
		Config:         config,
		Types:          generatedStructs,
		RootOutput:     output.Type,
		RootNumSlice:   output.NumSlice,
		BaseName:       typeInfo.RootType,
		CapabilityType: capabilityTypeFromString(typeInfo.CapabilityTypeRaw),
		Input:          input,
		ExtraImports:   extraImports,
		ID:             capID,
	}, nil
}

func extractInputAndConfig(generatedStructs map[string]Struct, typeInfo TypeInfo, root Struct) (*Struct, Struct) {
	delete(generatedStructs, typeInfo.RootType)
	configType := root.Outputs["Config"].Type
	inputField, ok := root.Outputs["Inputs"]
	var input *Struct
	if ok {
		inputType := inputField.Type
		inputS, ok := generatedStructs[inputType]
		if ok {
			input = &inputS
			delete(generatedStructs, inputType)
		}
	}
	config := generatedStructs[configType]
	for k := range generatedStructs {
		if strings.HasPrefix(k, configType) || (input != nil && strings.HasPrefix(k, input.Name)) {
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

				if strings.HasPrefix(f.Type, "*") {
					f.Type = f.Type[1:]
				}

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
		case `"reflect"`, `"fmt"`, `"encoding/json"`:
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
