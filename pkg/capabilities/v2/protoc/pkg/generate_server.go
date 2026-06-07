package pkg

import (
	_ "embed"
	"errors"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type serverArgs struct {
	*protogen.File
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

//go:embed templates/server_with_monitoring.go.tmpl
var goServerWithMonitoringTemplate string

var serverTemplates = map[ServerLanguage]TemplateGenerator{
	ServerLangaugeGo: {
		Name:               "go_server",
		Template:           goServerTemplate,
		FileNameTemplate:   "server/{{.}}_server_gen.go",
		StringLblValue:     StringLblValue(true),
		PbLabelTLangLabels: PbLabelToGoLabels,
	},
}

var serverWithMonitoringTemplates = map[ServerLanguage]TemplateGenerator{
	ServerLangaugeGo: {
		Name:               "go_server_with_monitoring",
		Template:           goServerWithMonitoringTemplate,
		FileNameTemplate:   "server/{{.}}_server_gen.go",
		StringLblValue:     StringLblValue(true),
		PbLabelTLangLabels: PbLabelToGoLabels,
	},
}

func GenerateServer(
	plugin *protogen.Plugin,
	file *protogen.File,
	serverLanguage ServerLanguage,
	toolName,
	localPrefix string,
	withMonitoring bool) error {
	if len(file.Services) == 0 {
		return nil
	}

	templates := serverTemplates
	if withMonitoring {
		templates = serverWithMonitoringTemplates
	}

	tmpl, ok := templates[serverLanguage]
	if !ok {
		return fmt.Errorf("unsupported server language: %s", serverLanguage)
	}

	args := serverArgs{File: file}
	return tmpl.GenerateFile(file, plugin, args, toolName, localPrefix)
}
