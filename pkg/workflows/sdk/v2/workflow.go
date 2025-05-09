package sdk

type WorkflowArgs[T any] struct {
	Handlers []Handler[T]
	Spend    *SpendLimits // Optional spend limits for the entire workflow
}

// WithMaxSpend sets the spend limits for the workflow
func (w *WorkflowArgs[T]) WithMaxSpend(limits *SpendLimits) *WorkflowArgs[T] {
	w.Spend = limits
	return w
}
