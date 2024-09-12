package workflows

type Runner interface {
	Run(factory *WorkflowSpecFactory) error
	Config() []byte
}
