package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

//go:embed go_workflow_builder.go.tmpl
var goWorkflowTemplate string

//go:embed go_mock_capability_builder.go.tmpl
var goWorkflowTestTemplate string

var dir = flag.String("dir", "", fmt.Sprintf("Directory to search for %s files, if a file is provided, the directory it is in will be used", cmd.CapabilitySchemaFilePattern.String()))

func main() {
	flag.Parse()
	if err := run(*dir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func run(dir string) error {
	// To allow go generate to work with $GO_FILE
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		dir = path.Dir(dir)
	}

	return cmd.GenerateTypes(dir, []cmd.WorkflowHelperGenerator{
		&cmd.TemplateWorkflowGeneratorHelper{
			Templates: map[string]cmd.TemplateAndCondition{
				"{{.BaseName|ToSnake}}_builders_generated.go":              cmd.BaseGenerate{TemplateValue: goWorkflowTemplate},
				"{{.Package}}test/{{.BaseName|ToSnake}}_mock_generated.go": cmd.TestHelperGenerate{TemplateValue: goWorkflowTestTemplate},
			},
		},
	})
}
