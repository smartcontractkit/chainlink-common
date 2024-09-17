package sdk

type Runner interface {
	Run(factory *WorkflowSpecFactory)
	Config() []byte
}
