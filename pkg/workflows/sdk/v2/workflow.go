package sdk

type Workflow[C any] []ExecutionHandler[C, Runtime]
