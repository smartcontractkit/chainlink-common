package pipeline_test

import (
	"context"
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.PipelineRunnerService = (*StaticPipelineRunnerService)(nil)

type StaticPipelineRunnerConfig struct {
	Spec        string
	Vars        types.Vars
	Options     types.Options
	TaskResults types.TaskResults
}

type StaticPipelineRunnerService struct {
	StaticPipelineRunnerConfig
}

func (pr *StaticPipelineRunnerService) ExecuteRun(ctx context.Context, s string, v types.Vars, o types.Options) (types.TaskResults, error) {
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
