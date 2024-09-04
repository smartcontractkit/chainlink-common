package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

var dir = flag.String("dir", "", fmt.Sprintf("Directory to search for %s files, if a file is provided, the directory it is in will be used", cmd.CapabilitySchemaFilePattern.String()))
var localPrefix = flag.String("local_prefix", "github.com/smartcontractkit", "The local prefix to use when formatting go files")
var extraUrls = flag.String("extra_urls", "", "Comma separated list of extra URLs to fetch schemas from")

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

	if localPrefix == nil {
		tmp := "github.com/smartcontractkit"
		localPrefix = &tmp
	}

	var extras []string
	if extraUrls != nil && *extraUrls != "" {
		extras = strings.Split(*extraUrls, ",")
	}

	return cmd.GenerateTypes(dir, *localPrefix, extras, []cmd.WorkflowHelperGenerator{
		&cmd.TemplateWorkflowGeneratorHelper{
			Templates: map[string]cmd.TemplateAndCondition{
				"{{.BaseName|ToSnake}}_builders_generated.go":              cmd.GoWorkflowTemplate(),
				"{{.Package}}test/{{.BaseName|ToSnake}}_mock_generated.go": cmd.GoTestTemplate(),
			},
		},
	})
}
