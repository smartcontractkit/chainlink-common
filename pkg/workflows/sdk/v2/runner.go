package sdk

type Runner[C any] interface {
	Run(initFn func(wcx *Environment[C]) (Workflow[C], error))
}
