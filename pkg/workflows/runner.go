package workflows

type Runner interface {
	Run(factory *WorkflowSpecFactory)
	Config() []byte
}
