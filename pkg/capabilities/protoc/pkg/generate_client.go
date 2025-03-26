package pkg

import (
	_ "embed"
	"path"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"

	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:embed templates/client.go.tmpl
var goClientTemplate string

var clientTemplates = []templateGenerator{
	{
		Name:             "go_client",
		Template:         goClientTemplate,
		FileNameTemplate: "{{.}}_client.go",
	},
}

//go:embed templates/client_trigger.go.tmpl
var goTriggerClientTemplate string

var triggerClientTemplates = []templateGenerator{
	{
		Name:             "go_trigger_client",
		Template:         goTriggerClientTemplate,
		FileNameTemplate: "{{.}}_client.go",
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

func GenerateClient(plugin *protogen.Plugin, mode *wasmpb.Mode, trigger bool, capabilityId string, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}

	templates := clientTemplates
	if trigger {
		templates = triggerClientTemplates
	}

	for _, template := range templates {
		args := clientArgs{mode: mode, File: file, CapabilityId: capabilityId}
		fileName, content, err := template.Generate(path.Base(file.GeneratedFilenamePrefix), args)
		if err != nil {
			return err
		}

		g := plugin.NewGeneratedFile(fileName, "")
		g.P(content)
	}

	return nil
}
