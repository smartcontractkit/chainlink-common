package sdk

import (
	"io"
)

type RunnerBase interface {
	LogWriter() io.Writer
	Config() []byte
}

type DonRunner interface {
	RunnerBase

	Run(args *WorkflowArgs[DonRuntime])
}

type NodeRunner interface {
	RunnerBase

	Run(args *WorkflowArgs[NodeRuntime])
}
