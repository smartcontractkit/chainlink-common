package pkg

import (
	_ "embed"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"

	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:embed templates/client.go.tmpl
var goClientBaseTemplate string

//go:embed templates/client_trigger.go.tmpl
var goTriggerMethodTemplate string

//go:embed templates/client_action.go.tmpl
var goActionMethodTemplate string

//go:embed templates/mock.go.tmpl
var goMockTemplate string

var clientTemplates = []templateGenerator{
	{
		Name:             "go_client",
		Template:         goClientBaseTemplate,
		FileNameTemplate: "{{.}}_client_gen.go",
		Partials: map[string]string{
			"trigger_method": goTriggerMethodTemplate,
			"action_method":  goActionMethodTemplate,
		},
	},
	{
		Name:             "go_mock",
		Template:         goMockTemplate,
		FileNameTemplate: "{{ToLower .}}mock/{{.}}_mock_gen.go",
	},
}

type clientArgs struct {
	mode *wasmpb.Mode
	*protogen.File
	CapabilityId string
}

func (c clientArgs) Mode() string {
	return strcase.ToCamel(wasmpb.Mode_name[int32(*c.mode)])
}

func GenerateClient(plugin *protogen.Plugin, mode *wasmpb.Mode, capabilityId string, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}

	for _, template := range clientTemplates {
		args := clientArgs{mode: mode, File: file, CapabilityId: capabilityId}
		if err := template.GenerateFile(file, plugin, args); err != nil {
			return err
		}
	}

	return nil
}
