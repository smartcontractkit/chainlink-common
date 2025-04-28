package sdk

type WorkflowArgs[T any] struct {
	Handlers []Handler[T]
}
