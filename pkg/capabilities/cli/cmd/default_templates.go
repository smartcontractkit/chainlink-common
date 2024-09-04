package cmd

import (
	"embed"
	_ "embed"
)

//go:embed templates/workflow_builder.go.tmpl templates/wrapper.go.tmpl
var goWorkflowTemplate embed.FS

//go:embed templates/mock_capability_builder.go.tmpl
var goWorkflowTestTemplate embed.FS

func GoWorkflowTemplate() TemplateAndCondition {
	return &BaseGenerate{
		FSValue:   goWorkflowTemplate,
		RootValue: "workflow_builder.go.tmpl",
	}
}

func GoTestTemplate() TemplateAndCondition {
	return (&BaseGenerate{
		FSValue:   goWorkflowTestTemplate,
		RootValue: "mock_capability_builder.go.tmpl",
	}).ForTestOnly()
}
