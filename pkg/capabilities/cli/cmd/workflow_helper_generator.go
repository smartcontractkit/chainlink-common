package cmd

type WorkflowHelperGenerator interface {
	Generate(info GeneratedInfo) (map[string]string, error)
}
