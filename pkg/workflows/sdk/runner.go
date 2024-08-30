package sdk

type Runner interface {
	Run(factory *WorkflowSpecFactory) error
	Config() []byte
}
