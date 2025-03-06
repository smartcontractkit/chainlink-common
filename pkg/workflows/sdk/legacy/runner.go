package legacy

type Runner interface {
	Run(factory *WorkflowSpecFactory)
	Config() []byte
}
