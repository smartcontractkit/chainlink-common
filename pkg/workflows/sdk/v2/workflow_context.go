package sdk

import (
	"io"
	"log/slog"
)

type WorkflowContext[C any] struct {
	Config    C
	LogWriter io.Writer
	Logger    *slog.Logger
}
