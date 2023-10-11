package types

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
)

type Vars struct {
	Vars map[string]interface{}
}
type TaskRunResults []TaskRunResult
type TaskRunResult struct {
	ID         uuid.UUID
	Task       Task
	TaskRun    TaskRun
	Result     Result
	Attempts   uint
	CreatedAt  time.Time
	FinishedAt null.Time
	RunInfo    RunInfo
}

type Task interface{}
type TaskRun struct {
	Type       string
	CreatedAt  time.Time
	FinishedAt time.Time
	Output     string
	Error      interface{}
	DotID      string
}

type Result struct {
	Value interface{}
	Error error
}
type RunInfo struct {
	IsRetryable bool
	IsPending   bool
}

type Spec struct {
	ID                int32
	DotDagSource      string
	CreatedAt         time.Time
	MaxTaskDuration   time.Duration
	GasLimit          *uint32
	ForwardingAllowed bool

	JobID   int32
	JobName string
	JobType string
}

type SpecData struct {
	ID string
}

type Run struct {
	ID               int64
	PipelineSpecID   int32
	PipelineSpec     Spec
	Meta             JSONSerializable
	AllErrors        RunErrors
	FatalErrors      RunErrors
	Inputs           JSONSerializable
	Outputs          JSONSerializable
	CreatedAt        time.Time
	FinishedAt       null.Time
	PipelineTaskRuns []TaskRun
	State            RunStatus

	Pending      bool
	FailSilently bool
}

type JSONSerializable struct {
	Val   interface{}
	Valid bool
}

type RunErrors []null.String
type RunStatus string
