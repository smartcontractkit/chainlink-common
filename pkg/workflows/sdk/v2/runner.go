package sdk

import (
	"io"
	"log/slog"
)

type RunnerBase interface {
	LogWriter() io.Writer
	Logger() *slog.Logger
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
