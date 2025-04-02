package cmd

import (
	"fmt"
	"strings"

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

func generatedInfoFromSrc(
	src, fullPkg string, capID *string, typeInfo TypeInfo, includeType func(name string) bool) (GeneratedInfo, error) {
	reader := GoStructReader{IncludeType: includeType}

	generatedStructs, pkg, extraImports, err := reader.Read(src)
	if err != nil {
		return GeneratedInfo{}, err
	}

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
		CapabilityType: typeInfo.CapabilityType,
		Input:          input,
		ExtraImports:   extraImports,
		ID:             capID,
		FullPackage:    fullPkg,
	}, nil
}

func packageFromSchemaID(schemaID string) (string, error) {
	fullPkg := schemaID

	// drop protocol
	index := strings.Index(fullPkg, "//")
	if index != -1 {
		fullPkg = fullPkg[index+2:]
	}

	// drop the capability name and version
	index = strings.LastIndex(fullPkg, "/")
	if index == -1 {
		return "", fmt.Errorf("invalid schema ID: %s must end in /capability_name and optioanlly a version", schemaID)
	}

	fullPkg = fullPkg[:index]
	return fullPkg, nil
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

func lastAfterDot(s string) string {
	parts := strings.Split(s, ".")
	return parts[len(parts)-1]
}
