package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type GeneratedInfo struct {
	Package        string
	Config         Struct
	Types          map[string]Struct
	CapabilityType capabilities.CapabilityType
	BaseName       string
	RootOutput     string
}

type Struct struct {
	Name    string
	Outputs map[string]Field
	Inputs  map[string]Field
}

type Field struct {
	Type        string
	NumSlice    int
	IsPrimitive bool
}

func StructsFromSrc(src, baseName string, tpe capabilities.CapabilityType) (GeneratedInfo, error) {
	fset := token.NewFileSet()

	// Parse the source code string
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return GeneratedInfo{}, err
	}

	rawInfo := map[string]Struct{}

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for TypeSpec nodes
		if ts, ok := n.(*ast.TypeSpec); ok {
			// Check if the TypeSpec's Type is a StructType
			if structType, ok := ts.Type.(*ast.StructType); ok {
				s := Struct{
					Name:    ts.Name.Name,
					Outputs: map[string]Field{},
					Inputs:  map[string]Field{},
				}

				for _, field := range structType.Fields.List {
					start := fset.Position(field.Type.Pos()).Offset
					end := fset.Position(field.Type.End()).Offset
					typeStr := src[start:end]
					if typeStr == "interface{}" {
						typeStr = "any"
					}
					f := Field{}

					originalTypeStr := typeStr
					if typeStr[0] != originalTypeStr[0] {
						f.IsPrimitive = true
					}
					f.Type = typeStr
					s.Outputs[field.Names[0].Name] = f
					rawInfo[ts.Name.Name] = s
				}
			}
		}
		return true
	})

	root := rawInfo[baseName]
	delete(rawInfo, baseName)
	configType := root.Outputs["Config"].Type
	config := rawInfo[configType]
	delete(rawInfo, configType)

	return GeneratedInfo{
		Package:/*TODO*/ "streams",
		Config:         config,
		Types:          rawInfo,
		RootOutput:     root.Outputs["Outputs"].Type,
		BaseName:       baseName,
		CapabilityType: tpe,
	}, nil
}
