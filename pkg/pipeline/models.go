package pipeline

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

type Spec struct {
	ID                int32
	DotDagSource      string         `json:"dotDagSource"`
	CreatedAt         time.Time      `json:"-"`
	MaxTaskDuration   utils.Interval `json:"-"`
	GasLimit          *uint32        `json:"-"`
	ForwardingAllowed bool           `json:"-"`

	JobID   int32  `json:"-"`
	JobName string `json:"-"`
	JobType string `json:"-"`
}

// RunStatus represents the status of a run
type RunStatus string

const (
	// RunStatusUnknown is the when the run status cannot be determined.
	RunStatusUnknown RunStatus = "unknown"
	// RunStatusRunning is used for when a run is actively being executed.
	RunStatusRunning RunStatus = "running"
	// RunStatusSuspended is used when a run is paused and awaiting further results.
	RunStatusSuspended RunStatus = "suspended"
	// RunStatusErrored is used for when a run has errored and will not complete.
	RunStatusErrored RunStatus = "errored"
	// RunStatusCompleted is used for when a run has successfully completed execution.
	RunStatusCompleted RunStatus = "completed"
)

// Completed returns true if the status is RunStatusCompleted.
func (s RunStatus) Completed() bool {
	return s == RunStatusCompleted
}

// Errored returns true if the status is RunStatusErrored.
func (s RunStatus) Errored() bool {
	return s == RunStatusErrored
}

// Finished returns true if the status is final and can't be changed.
func (s RunStatus) Finished() bool {
	return s.Completed() || s.Errored()
}
