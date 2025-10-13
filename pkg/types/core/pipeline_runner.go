package core

import (
	"context"
	"sort"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/jsonserializable"
)

type Vars struct {
	Vars map[string]any
}

type Options struct {
	MaxTaskDuration time.Duration
}

type TaskValue struct {
	Error      error
	Value      jsonserializable.JSONSerializable
	IsTerminal bool
}

type TaskResult struct {
	ID    string
	Type  string
	Index int

	TaskValue
}

type TaskResults []TaskResult

func (tr TaskResults) FinalResults() []TaskValue {
	sort.Slice(tr, func(i, j int) bool {
		return tr[i].Index < tr[j].Index
	})

	var found bool
	results := []TaskValue{}
	for _, t := range tr {
		if t.IsTerminal {
			results = append(results, t.TaskValue)
			found = true
		}
	}

	if !found {
		panic("expected at least one final task")
	}

	return results
}

type PipelineRunnerService interface {
	ExecuteRun(ctx context.Context, spec string, vars Vars, options Options) (TaskResults, error)
}
