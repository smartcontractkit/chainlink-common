package sdk

type WorkflowArgs[T any] struct {
	Handlers []Handler[T]
	MaxSpend *MaxSpendLimits // Optional max spend limits for the entire workflow
}

// WithMaxSpend sets the max spend limits for the workflow
func (w *WorkflowArgs[T]) WithMaxSpend(limits *MaxSpendLimits) *WorkflowArgs[T] {
	w.MaxSpend = limits
	return w
}
