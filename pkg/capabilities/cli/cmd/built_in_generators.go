package cmd

import _ "embed"

//go:embed go_workflow_builder.go.tmpl
var goWorkflowTemplate string

//go:embed go_mock_capability_builder.go.tmpl
var goWorkflowTestTemplate string

func AddDefaultGoTemplates(to map[string]TemplateAndCondition, includeMocks bool) {
	to["{{if .BaseName}}{{.BaseName|ToSnake}}_builders{{ else }}wrappers{{ end }}_generated.go"] = BaseGenerate{TemplateValue: goWorkflowTemplate}
	if includeMocks {
		to["{{.Package}}test/{{.BaseName|ToSnake}}_mock_generated.go"] = TestHelperGenerate{TemplateValue: goWorkflowTestTemplate}
	}
}
