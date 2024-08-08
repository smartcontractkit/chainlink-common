package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

var Dir string

// CapabilitySchemaFilePattern is used to extract the package name from the file path.
// This is used as the package name for the generated Go types.
var CapabilitySchemaFilePattern = regexp.MustCompile(`([^/]+)_(action|trigger|consensus|target)-schema\.json$`)

// reg := regexp.MustCompile(`([^/]+)_(trigger|action)\.json$`)

func init() {
	generateTypesCmd.Flags().StringVar(&Dir, "dir", ".", fmt.Sprintf("Directory to search for %s files", CapabilitySchemaFilePattern.String()))
	if err := generateTypesCmd.MarkFlagDirname("dir"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.AddCommand(generateTypesCmd)
}

// Finds all files that match CapabilitySchemaFilePattern in the provided directory and generates Go
// types for each.
var generateTypesCmd = &cobra.Command{
	Use:   "generate-types",
	Short: "Generate Go types from JSON schema capability definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := cmd.Flag("dir").Value.String()
		return GenerateTypes(dir, []WorkflowHelperGenerator{
			&TemplateWorkflowGeneratorHelper{
				Templates: map[string]string{"{{.Package|PkgToCapPkg}}/{{.BaseName|ToSnake}}_builders_generated.go": goWorkflowTemplate},
			},
		})
	},
}

func GenerateTypes(dir string, helpers []WorkflowHelperGenerator) error {
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
		return errors.New(fmt.Sprintf("error walking the directory %v: %v\n", dir, err))
	}

	cfg, err := ConfigFromSchemas(schemaPaths)
	if err != nil {
		return err
	}

	for _, schemaPath := range schemaPaths {
		allFiles := map[string]string{}
		file, content, rootType, capabilityTypeRaw, err := TypesFromJSONSchema(schemaPath, cfg)
		if err != nil {
			return err
		}

		var capabilityType capabilities.CapabilityType
		for ; capabilityType.IsValid() == nil && (capabilityType.String() != capabilityTypeRaw); capabilityType++ {
		}
		if err = capabilityType.IsValid(); err != nil {
			return fmt.Errorf("invalid capability type %v", capabilityTypeRaw)
		}

		allFiles[file] = content

		abs, err := filepath.Abs(schemaPath)
		if err != nil {
			return err
		}

		structs, err := StructsFromSrc(path.Dir(abs), content, rootType, capabilityType)
		if err != nil {
			return err
		}

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

		if err = printFiles(path.Dir(schemaPath), allFiles); err != nil {
			return err
		}

		fmt.Println("Generated types for", schemaPath)
	}
	return nil
}

func ConfigFromSchemas(schemaFilePaths []string) (generator.Config, error) {
	cfg := generator.Config{Warner: func(message string) { fmt.Printf("Warning: %s\n", message) }}
	for _, schemaFilePath := range schemaFilePaths {
		jsonSchema, err := schemas.FromJSONFile(schemaFilePath)
		if err != nil {
			return cfg, err
		}
		capabilityInfo := CapabilitySchemaFilePattern.FindStringSubmatch(schemaFilePath)
		if len(capabilityInfo) != 3 {
			return cfg, fmt.Errorf("invalid schema file path %v", schemaFilePath)
		}
		capabilityType := capabilityInfo[2]
		outputName := strings.Replace(schemaFilePath, capabilityType+"-schema.json", capabilityType+"_generated.go", 1)
		rootType := capitalize(capabilityType)

		cfg.SchemaMappings = append(cfg.SchemaMappings, generator.SchemaMapping{
			SchemaID:    jsonSchema.ID,
			PackageName: path.Dir(jsonSchema.ID[8:]),
			RootType:    rootType,
			OutputName:  outputName,
		})
	}
	return cfg, nil
}

// TypesFromJSONSchema generates Go types from a JSON schema file.
func TypesFromJSONSchema(schemaFilePath string, cfg generator.Config) (outputFilePath, outputContents, rootType, capabilityType string, err error) {
	capabilityInfo := CapabilitySchemaFilePattern.FindStringSubmatch(schemaFilePath)
	capabilityType = capabilityInfo[2]
	outputName := strings.Replace(schemaFilePath, capabilityType+"-schema.json", capabilityType+"_generated.go", 1)
	rootType = capitalize(capabilityType)

	gen, err := generator.New(cfg)
	if err != nil {
		return "", "", "", "", err
	}

	if err = gen.DoFile(schemaFilePath); err != nil {
		return "", "", "", "", err
	}

	generatedContents := gen.Sources()
	content := generatedContents[outputName]

	content = []byte(strings.Replace(string(content), "// Code generated by github.com/atombender/go-jsonschema", "// Code generated by pkg/capabilities/cli", 1))

	return outputName, string(content), rootType, capabilityType, nil
}

func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}
