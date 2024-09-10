package cmd

import (
	"go/types"
	"log"
	"maps"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// GenerateUserWrappers wraps user structs using the provided templates.
// header is called once with a single argument which is a slice of strings representing the imports for the package
// body is called once for each struct that is being wrapped.
func GenerateUserWrappers(
	structs map[string]bool,
	headerTemplate string,
	body TemplateAndCondition,
	outputFile string) error {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		log.Fatalf("Error loading package: %v\n", err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("Expected exactly one package, got %d\n", len(pkgs))
	}

	generatedInfos := FindStructsInPackage(pkgs[0], structs)
	headers := allHeaders(generatedInfos)

	t, err := template.New("headers").Parse(headerTemplate)
	if err != nil {
		return err
	}

	buf := &strings.Builder{}
	if err = t.Execute(buf, headers); err != nil {
		return err
	}

	outuptMap := map[string]string{outputFile: buf.String()}

	for _, info := range generatedInfos {
		th := &TemplateWorkflowGeneratorHelper{map[string]TemplateAndCondition{outputFile: body}}
		output, genErr := th.Generate(info)
		if genErr != nil {
			return genErr
		}

		for file, content := range output {
			outuptMap[file] += content
		}
	}

	// TODO I'm here
	return codgen.WriteFiles()
}

func allHeaders(generatedInfos []GeneratedInfo) []string {
	var headers []string
	var seenHeaders = map[string]bool{}
	for _, info := range generatedInfos {
		for _, header := range info.ExtraImports {
			if !seenHeaders[header] {
				headers = append(headers, header)
				seenHeaders[header] = true
			}
		}
	}

	return headers
}

func FindStructsInPackage(pkg *packages.Package, structs map[string]bool) []GeneratedInfo {
	genInfos := make([]GeneratedInfo, 0, len(structs))
	missingStructs := maps.Clone(structs)
	for _, def := range pkg.TypesInfo.Defs {
		if def == nil || !structs[def.Name()] {
			continue
		}

		missingStructs[def.Name()] = false

		if strct, ok := def.Type().Underlying().(*types.Struct); ok {
			genInfos = append(genInfos, genInfoFor(pkg, def, strct, structs))
		} else {
			log.Fatalf("Expected %s to be a struct, but it is a %T\n", def.Name(), def.Type)
		}
	}

	if len(missingStructs) > 0 {
		missing := make([]string, 0, len(missingStructs))
		for s := range missingStructs {
			missing = append(missing, s)
		}
		log.Fatalf("Could not find structs: %s\n", strings.Join(missing, ", "))
	}

	return genInfos
}

func genInfoFor(pkg *packages.Package, def types.Object, strct *types.Struct, structs map[string]bool) GeneratedInfo {
	tpe := Struct{
		Name:    def.Name(),
		Outputs: map[string]Field{},
	}

	fields, imports := getFieldsAndImports(strct, structs)
	tpe.Outputs = fields

	return GeneratedInfo{
		Package:      pkg.Name,
		Types:        map[string]Struct{def.Name(): tpe},
		RootOutput:   def.Name(),
		ExtraImports: imports,
		FullPackage:  pkg.PkgPath,
	}
}

func getFieldsAndImports(strct *types.Struct, structs map[string]bool) (map[string]Field, []string) {
	outputs := map[string]Field{}
	seenImports := map[string]bool{}
	var extraImports []string

	numFields := strct.NumFields()
	for i := 0; i < numFields; i++ {
		field := strct.Field(i)
		if !field.Exported() {
			continue
		}

		fieldType := field.Type().String()
		if structs[fieldType] {
			fieldType += "Cap"
		}

		outputs[field.Name()] = Field{
			Type: fieldType,
			// TODO more accurate way do this
			IsPrimitive: unicode.IsLower(rune(fieldType[0])),
		}

		if imprt := extractPackageName(field.Type()); imprt != "" && !seenImports[imprt] {
			extraImports = append(extraImports, imprt)
			seenImports[imprt] = true
		}
	}

	return outputs, extraImports
}

func extractPackageName(t types.Type) string {
	switch typ := t.(type) {
	case *types.Named:
		// If the type is named and has an external package, return the package path
		if typ.Obj().Pkg() != nil {
			return typ.Obj().Pkg().Path()
		}
	}
	return ""
}
