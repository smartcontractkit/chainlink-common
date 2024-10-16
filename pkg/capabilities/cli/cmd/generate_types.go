package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
	"github.com/iancoleman/strcase"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
)

// CapabilitySchemaFilePattern is used to extract the package name from the file path.
// This is used as the package name for the generated Go types.
var CapabilitySchemaFilePattern = regexp.MustCompile(`([^/]+)_(action|trigger|consensus|target|common)-schema\.json$`)

const toolName = "github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli"

func GenerateTypes(dir, localPrefix string, extraUrls []string, helpers []WorkflowHelperGenerator) error {
	schemaPaths, err := schemaFilesFromDir(dir)
	if err != nil {
		return err
	}

	allPaths := make([]string, len(extraUrls)+len(schemaPaths))
	copy(allPaths, schemaPaths)
	copy(allPaths[len(schemaPaths):], extraUrls)
	cfgInfo, err := ConfigFromSchemas(allPaths)
	if err != nil {
		return err
	}

	for _, schemaPath := range schemaPaths {
		if err = generateFromSchema(schemaPath, localPrefix, cfgInfo, helpers); err != nil {
			return err
		}
	}
	return nil
}

func generateFromSchema(schemaPath, localPrefix string, cfgInfo ConfigInfo, helpers []WorkflowHelperGenerator) error {
	allFiles := map[string]string{}
	file, content, err := TypesFromJSONSchema(schemaPath, cfgInfo)
	if err != nil {
		return err
	}

	capabilityType := cfgInfo.SchemaToTypeInfo[schemaPath].CapabilityType
	if err = capabilityType.IsValid(); err != nil && string(capabilityType) != "common" {
		return fmt.Errorf("invalid capability type %v", capabilityType)
	}

	allFiles[file] = content
	typeInfo := cfgInfo.SchemaToTypeInfo[schemaPath]
	err = generateSchemaTypes(schemaPath, localPrefix, content, typeInfo, helpers, allFiles)
	if err != nil {
		return err
	}
	return nil
}

func generateSchemaTypes(schemaPath string, localPrefix string, content string, typeInfo TypeInfo, helpers []WorkflowHelperGenerator, allFiles map[string]string) error {
	fullPkg, err := packageFromSchemaID(typeInfo.SchemaID)
	if err != nil {
		return err
	}

	return generateFromGoSrc(path.Dir(schemaPath), localPrefix, content, typeInfo, helpers, allFiles, fullPkg, func(string) bool {
		return true
	})
}

func generateFromGoSrc(
	dir,
	localPrefix,
	content string,
	typeInfo TypeInfo,
	helpers []WorkflowHelperGenerator,
	allFiles map[string]string,
	fullPkg string,
	includeType func(name string) bool) error {
	structs, err := generatedInfoFromSrc(content, fullPkg, getCapID(typeInfo), typeInfo, includeType)
	if err != nil {
		return err
	}

	if err = generateHelpers(helpers, structs, allFiles); err != nil {
		return err
	}

	if err = codegen.WriteFiles(dir, localPrefix, toolName, allFiles); err != nil {
		return err
	}

	fmt.Println("Generated types for", dir)
	return nil
}

func getCapID(typeInfo TypeInfo) *string {
	id := typeInfo.SchemaID
	idParts := strings.Split(id, "/")
	id = idParts[len(idParts)-1]
	var capID *string
	if strings.Contains(id, "@") {
		capID = &id
	}
	return capID
}

func schemaFilesFromDir(dir string) ([]string, error) {
	var schemaPaths []string

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories and files that don't match the CapabilitySchemaFileExtension
		if info.IsDir() || !CapabilitySchemaFilePattern.MatchString(path) {
			return nil
		}

		schemaPaths = append(schemaPaths, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error walking the directory %v: %v", dir, err)
	}
	return schemaPaths, nil
}

func generateHelpers(helpers []WorkflowHelperGenerator, structs GeneratedInfo, allFiles map[string]string) error {
	for _, helper := range helpers {
		files, err := helper.Generate(structs)
		if err != nil {
			return err
		}

		for f, c := range files {
			if _, ok := allFiles[f]; ok {
				return fmt.Errorf("file %v is being created by more than one generator", f)
			}
			allFiles[f] = c
		}
	}
	return nil
}

func ConfigFromSchemas(schemaFilePaths []string) (ConfigInfo, error) {
	loader := schemas.NewDefaultCacheLoader([]string{}, []string{".yaml", ".yml"})
	configInfo := ConfigInfo{
		Config: generator.Config{
			Tags:           []string{"json", "yaml", "mapstructure"},
			Warner:         func(message string) { fmt.Printf("Warning: %s\n", message) },
			MinSizedInts:   true,
			YAMLExtensions: []string{".yaml", ".yml"},
			Loader:         loader,
		},
		SchemaToTypeInfo: map[string]TypeInfo{},
	}

	for _, schemaFilePath := range schemaFilePaths {
		capabilityInfo := CapabilitySchemaFilePattern.FindStringSubmatch(schemaFilePath)
		if len(capabilityInfo) != 3 {
			return configInfo, fmt.Errorf(
				"invalid schema file path %v, does not match pattern %s", schemaFilePath, CapabilitySchemaFilePattern)
		}

		jsonSchema, err := loader.Load(schemaFilePath, "")
		if err != nil {
			return configInfo, err
		}

		capabilityTypeRaw := capabilityInfo[2]
		fullName := strings.Join(append(strings.Split(capabilityInfo[1], "_")[1:], capabilityTypeRaw), "_")
		outputName := strings.Replace(schemaFilePath, fullName+"-schema.json", fullName+"_generated.go", 1)
		rootType := strcase.ToCamel(fullName)
		configInfo.SchemaToTypeInfo[schemaFilePath] = TypeInfo{
			CapabilityType:   capabilities.CapabilityType(capabilityTypeRaw),
			RootType:         rootType,
			SchemaID:         jsonSchema.ID,
			SchemaOutputFile: outputName,
		}

		configInfo.Config.SchemaMappings = append(configInfo.Config.SchemaMappings, generator.SchemaMapping{
			SchemaID:    jsonSchema.ID,
			PackageName: path.Dir(jsonSchema.ID[8:]),
			RootType:    rootType,
			OutputName:  outputName,
		})
	}
	return configInfo, nil
}

// TypesFromJSONSchema generates Go types from a JSON schema file.
func TypesFromJSONSchema(schemaFilePath string, cfgInfo ConfigInfo) (outputFilePath, outputContents string, err error) {
	typeInfo, ok := cfgInfo.SchemaToTypeInfo[schemaFilePath]
	if !ok {
		return "", "", fmt.Errorf("missing type info for %s", schemaFilePath)
	}

	gen, err := generator.New(cfgInfo.Config)
	if err != nil {
		return "", "", err
	}

	if err = gen.DoFile(schemaFilePath); err != nil {
		return "", "", err
	}

	generatedContents := gen.Sources()
	content := generatedContents[typeInfo.SchemaOutputFile]

	// codegen.WriteToFiles will add the header, so remove it here.
	content = []byte(strings.Replace(string(content), "// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.", "", 1))

	return typeInfo.SchemaOutputFile, string(content), nil
}
