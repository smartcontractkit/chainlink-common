package pkg

import (
	_ "embed"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/client.go.tmpl
var goClientBaseTemplate string

//go:embed templates/client_trigger.go.tmpl
var goTriggerMethodTemplate string

//go:embed templates/client_action.go.tmpl
var goActionMethodTemplate string

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
}

func GenerateClient(plugin *protogen.Plugin, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}

	for _, template := range clientTemplates {
		if err := template.GenerateFile(file, plugin, file); err != nil {
			return err
		}
	}

	return nil
}
