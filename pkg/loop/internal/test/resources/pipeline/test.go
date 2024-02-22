package pipeline_test

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

const spec = `
answer [type=sum values=<[ $(val), 2 ]>]
answer;
`

var (
	vars = types.Vars{
		Vars: map[string]interface{}{"foo": "baz"},
	}
	options = types.Options{
		MaxTaskDuration: 10 * time.Second,
	}
	taskResults = types.TaskResults([]types.TaskResult{
		{
			TaskValue: types.TaskValue{
				Value: "hello",
			},
			Index: 0,
		},
	})

	DefaultStaticPipelineRunnerConfig = StaticPipelineRunnerConfig{
		Spec:        spec,
		Vars:        vars,
		Options:     options,
		TaskResults: taskResults,
	}

	DefaultStaticPipelineRunner = StaticPipelineRunnerService{
		StaticPipelineRunnerConfig: DefaultStaticPipelineRunnerConfig,
	}
)
