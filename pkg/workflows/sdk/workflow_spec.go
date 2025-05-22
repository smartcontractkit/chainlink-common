package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	v2 "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type StepInputs struct {
	OutputRef string
	Mapping   map[string]any
}

// StepDefinition is the parsed representation of a step in a workflow.
//
// Within the workflow spec, they are called "Capability Properties".
type StepDefinition struct {
	ID     string
	Ref    string
	Inputs StepInputs
	Config map[string]any

	CapabilityType capabilities.CapabilityType
	MaxSpends      []v2.SpendLimits // Optional max spend limits for this step
}

type WorkflowSpec struct {
	Name      string
	Owner     string
	Triggers  []StepDefinition
	Actions   []StepDefinition
	Consensus []StepDefinition
	Targets   []StepDefinition
	MaxSpends []v2.SpendLimits // Optional max spend limits for the entire workflow
}

func (w *WorkflowSpec) Steps() []StepDefinition {
	s := []StepDefinition{}
	s = append(s, w.Actions...)
	s = append(s, w.Consensus...)
	s = append(s, w.Targets...)
	return s
}
