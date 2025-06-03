package sdk

type Runner[C any] interface {
	Run(initFn func(wcx *WorkflowContext[C]) (Workflows[C], error))
}
