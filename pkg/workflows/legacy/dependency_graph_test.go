package legacy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	workflows "github.com/smartcontractkit/chainlink-common/pkg/workflows/legacy"
)

func TestParseDependencyGraph(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		yaml   string
		graph  map[string]map[string]struct{}
		errMsg string
	}{
		{
			name: "basic example",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567      
triggers:
  - id: "a-trigger@1.0.0"
    config: {}

actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      trigger_output: $(trigger.outputs)

consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      trigger_output: $(trigger.outputs)
      an-action_output: $(an-action.outputs)

targets:
  - id: "a-target@1.0.0"
    config: {}
    ref: "a-target"
    inputs: 
      consensus_output: $(a-consensus.outputs)
`,
			graph: map[string]map[string]struct{}{
				workflows.KeywordTrigger: {
					"an-action":   struct{}{},
					"a-consensus": struct{}{},
				},
				"an-action": {
					"a-consensus": struct{}{},
				},
				"a-consensus": {
					"a-target": struct{}{},
				},
				"a-target": {},
			},
		},
		{
			name: "circular relationship",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}

actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      trigger_output: $(trigger.outputs)
      output: $(a-second-action.outputs)
  - id: "a-second-action@1.0.0"
    config: {}
    ref: "a-second-action"
    inputs:
      output: $(an-action.outputs)

consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      trigger_output: $(trigger.outputs)
      an-action_output: $(an-action.outputs)

targets:
  - id: "a-target@1.0.0"
    config: {}
    ref: "a-target"
    inputs: 
      consensus_output: $(a-consensus.outputs)
`,
			errMsg: "edge would create a cycle",
		},
		{
			name: "indirect circular relationship",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}

actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      trigger_output: $(trigger.outputs)
      action_output: $(a-third-action.outputs)
  - id: "a-second-action@1.0.0"
    config: {}
    ref: "a-second-action"
    inputs:
      output: $(an-action.outputs)
  - id: "a-third-action@1.0.0"
    config: {}
    ref: "a-third-action"
    inputs:
      output: $(a-second-action.outputs)

consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      trigger_output: $(trigger.outputs)
      an-action_output: $(an-action.outputs)

targets:
  - id: "a-target@1.0.0"
    config: {}
    ref: "a-target"
    inputs: 
      consensus_output: $(a-consensus.outputs)
`,
			errMsg: "edge would create a cycle",
		},
		{
			name: "relationship doesn't exist",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}

actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      trigger_output: $(trigger.outputs)
      action_output: $(missing-action.outputs)

consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      an-action_output: $(an-action.outputs)

targets:
  - id: "a-target@1.0.0"
    config: {}
    ref: "a-target"
    inputs: 
      consensus_output: $(a-consensus.outputs)
`,
			errMsg: "source vertex missing-action: vertex not found",
		},
		{
			name: "two trigger nodes",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}
  - id: "a-second-trigger@1.0.0"
    config: {}

actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      trigger_output: $(trigger.outputs)

consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      an-action_output: $(an-action.outputs)

targets:
  - id: "a-target@1.0.0"
    ref: "a-target"
    config: {}
    inputs:
      consensus_output: $(a-consensus.outputs)
`,
			graph: map[string]map[string]struct{}{
				workflows.KeywordTrigger: {
					"an-action": struct{}{},
				},
				"an-action": {
					"a-consensus": struct{}{},
				},
				"a-consensus": {
					"a-target": struct{}{},
				},
				"a-target": {},
			},
		},
		{
			name: "non-trigger step with no dependent refs",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}
  - id: "a-second-trigger@1.0.0"
    config: {}
actions:
  - id: "an-action@1.0.0"
    config: {}
    ref: "an-action"
    inputs:
      hello: "world"
consensus:
  - id: "a-consensus@1.0.0"
    ref: "a-consensus"
    config: {}
    inputs:
      trigger_output: $(trigger.outputs)
      action_output: $(an-action.outputs)
targets:
  - id: "a-target@1.0.0"
    config: {}
    ref: "a-target"
    inputs:
      consensus_output: $(a-consensus.outputs)
`,
			errMsg: "invalid refs",
		},
		{
			name: "duplicate refs in a step",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567
triggers:
  - id: "a-trigger@1.0.0"
    config: {}
  - id: "a-second-trigger@1.0.0"
    config: {}
actions:
  - id: "an-action@1.0.0"
    ref: "an-action"
    config: {}
    inputs:
      trigger_output: $(trigger.outputs)
consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      action_output: $(an-action.outputs)
targets:
  - id: "a-target@1.0.0"
    ref: "a-target"
    config: {}
    inputs:
      consensus_output: $(a-consensus.outputs)
      consensus_output: $(a-consensus.outputs)
`,
			graph: map[string]map[string]struct{}{
				workflows.KeywordTrigger: {
					"an-action": struct{}{},
				},
				"an-action": {
					"a-consensus": struct{}{},
				},
				"a-consensus": {
					"a-target": struct{}{},
				},
				"a-target": {},
			},
		},
		{
			name: "passthrough ref interpolation",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567  
triggers:
  - id: "a-trigger@1.0.0"
    config: {}
  - id: "a-second-trigger@1.0.0"
    config: {}
actions:
  - id: "an-action@1.0.0"
    ref: "an-action"
    config: {}
    inputs: $(trigger.outputs)
consensus:
  - id: "a-consensus@1.0.0"
    config: {}
    ref: "a-consensus"
    inputs:
      action_output: $(an-action.outputs)
targets:
  - id: "a-target@1.0.0"
    ref: "a-target"
    config: {}
    inputs:
      consensus_output: $(a-consensus.outputs)
`,
			graph: map[string]map[string]struct{}{
				workflows.KeywordTrigger: {
					"an-action": struct{}{},
				},
				"an-action": {
					"a-consensus": struct{}{},
				},
				"a-consensus": {
					"a-target": struct{}{},
				},
				"a-target": {},
			},
		},
		{
			name: "workflow without consensus",
			yaml: `
name: length_ten # exactly 10 characters
owner: 0x0123456789abcdef0123456789abcdef01234567
triggers:
  - id: "a-trigger@1.0.0"
    config: {}
targets:
  - id: "a-target@1.0.0"
    ref: "a-target"
    config: {}
    inputs: $(trigger.outputs)
`,
			graph: map[string]map[string]struct{}{
				workflows.KeywordTrigger: {
					"a-target": struct{}{},
				},
				"a-target": {},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			//wf, err := workflows.Parse(tc.yaml)
			wf, err := parseDependencyGraph(tc.yaml)
			if tc.errMsg != "" {
				assert.ErrorContains(st, err, tc.errMsg)
			} else {
				require.NoError(st, err)

				adjacencies, err := wf.AdjacencyMap()
				require.NoError(t, err)

				got := map[string]map[string]struct{}{}
				for k, v := range adjacencies {
					if _, ok := got[k]; !ok {
						got[k] = map[string]struct{}{}
					}
					for adj := range v {
						got[k][adj] = struct{}{}
					}
				}

				assert.Equal(st, tc.graph, got, adjacencies)
			}
		})
	}
}

func parseDependencyGraph(yamlWorkflow string) (*workflows.DependencyGraph, error) {
	spec, err := workflows.ParseWorkflowSpecYaml(yamlWorkflow)
	if err != nil {
		return nil, err
	}

	return workflows.BuildDependencyGraph(spec)
}
