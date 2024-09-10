package cmd

import (
	"embed"
	_ "embed"
)

//go:embed templates/workflow_builder.go.tmpl templates/wrapper.go.tmpl
var goWorkflowTemplate embed.FS

func GoWorkflowTemplate() TemplateAndCondition {
	return &BaseGenerate{
		FSValue:   goWorkflowTemplate,
		RootValue: "workflow_builder.go.tmpl",
	}
}

//go:embed templates/mock_capability_builder.go.tmpl
var goWorkflowTestTemplate embed.FS

func GoTestTemplate() TemplateAndCondition {
	return (&BaseGenerate{
		FSValue:   goWorkflowTestTemplate,
		RootValue: "mock_capability_builder.go.tmpl",
	}).ForTestOnly()
}

//go:embed templates/wrapper.go.tmpl
var goUserWrapper embed.FS

func GoUserWrapper() TemplateAndCondition {
	return &BaseGenerate{
		FSValue:   goUserWrapper,
		RootValue: "wrapper.go.tmpl",
	}
}

//go:embed templates/user_structs_header.go.tmpl
var goUserStructsHeader string

func GoUserStructsHeader() string {
	return goUserStructsHeader
}
