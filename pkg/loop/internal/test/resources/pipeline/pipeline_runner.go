package pipeline_test

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

const pipleine_spec = `
answer [type=sum values=<[ $(val), 2 ]>]
answer;
`

var PipelineRunnerImpl = staticPipelineRunnerService{
	staticPipelineRunnerConfig: staticPipelineRunnerConfig{
		Spec: pipleine_spec,
		Vars: types.Vars{
			Vars: map[string]interface{}{"foo": "baz"},
		},
		Options: types.Options{
			MaxTaskDuration: 10 * time.Second,
		},
		TaskResults: types.TaskResults([]types.TaskResult{
			{
				TaskValue: types.TaskValue{
					Value: "hello",
				},
				Index: 0,
			},
		}),
	},
}

type PipelineRunnerEvaluator interface {
	types.PipelineRunnerService

	// Evaluate runs the pipeline and returns the results
	Evaluate(ctx context.Context, other types.PipelineRunnerService) error
}

var _ types.PipelineRunnerService = (*staticPipelineRunnerService)(nil)

type staticPipelineRunnerConfig struct {
	Spec        string
	Vars        types.Vars
	Options     types.Options
	TaskResults types.TaskResults
}

type staticPipelineRunnerService struct {
	staticPipelineRunnerConfig
}

func (pr *staticPipelineRunnerService) ExecuteRun(ctx context.Context, s string, v types.Vars, o types.Options) (types.TaskResults, error) {
	if s != pr.Spec {
		return nil, fmt.Errorf("expected %s but got %s", pr.Spec, s)
	}
	if !reflect.DeepEqual(v, pr.Vars) {
		return nil, fmt.Errorf("expected %+v but got %+v", pr.Vars, v)
	}
	if !reflect.DeepEqual(o, pr.Options) {
		return nil, fmt.Errorf("expected %+v but got %+v", pr.Options, o)
	}
	return pr.TaskResults, nil
}

func (pr *staticPipelineRunnerService) Evaluate(ctx context.Context, other types.PipelineRunnerService) error {
	tr, err := pr.ExecuteRun(ctx, pr.Spec, pr.Vars, pr.Options)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	if !reflect.DeepEqual(tr, pr.TaskResults) {
		return fmt.Errorf("expected TaskResults %+v but got %+v", pr.TaskResults, tr)
	}
	return nil
}
