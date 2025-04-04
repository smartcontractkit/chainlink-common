package pkg

import (
	_ "embed"
	"errors"
	"fmt"
	"path"

	"google.golang.org/protobuf/compiler/protogen"
)

type serverArgs struct {
	*protogen.File
	CapabilityId string
}

type ServerLanguage string

const (
	ServerLangaugeGo ServerLanguage = "go"
)

func (s ServerLanguage) Validate() error {
	switch s {
	case ServerLangaugeGo:
		return nil
	default:
		return errors.New("unsupported server language")
	}
}

//go:embed templates/server.go.tmpl
var goServerTemplate string

var serverTemplates = map[ServerLanguage]templateGenerator{
	ServerLangaugeGo: {
		Name:             "go_server",
		Template:         goServerTemplate,
		FileNameTemplate: "{{.}}_server_gen.go",
	},
}

//go:embed templates/server_trigger.go.tmpl
var goTriggerServerTemplate string

var triggerServerTemplates = map[ServerLanguage]templateGenerator{
	ServerLangaugeGo: {
		Name:             "go_trigger_server",
		Template:         goTriggerServerTemplate,
		FileNameTemplate: "{{.}}_server_gen.go",
	},
}

func GenerateServer(plugin *protogen.Plugin, trigger bool, capabilityId string, serverLanguage ServerLanguage, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}

	templates := serverTemplates
	if trigger {
		templates = triggerServerTemplates
	}

	template, ok := templates[serverLanguage]
	if !ok {
		return fmt.Errorf("unsupported server language: %s", serverLanguage)
	}

	args := serverArgs{File: file, CapabilityId: capabilityId}
	fileName, content, err := template.Generate(file.GeneratedFilenamePrefix, args)
	if err != nil {
		return err
	}

	g := plugin.NewGeneratedFile(path.Join("server", path.Base(fileName)), "")
	g.P(content)

	return nil
}
