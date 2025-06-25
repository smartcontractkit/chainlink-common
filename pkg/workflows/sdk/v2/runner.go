package sdk

type Runner[C any] interface {
	Run(initFn func(env *Environment[C]) (Workflow[C], error))
}
