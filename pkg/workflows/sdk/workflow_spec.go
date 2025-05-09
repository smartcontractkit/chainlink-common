package sdk

import "github.com/smartcontractkit/chainlink-common/pkg/capabilities"

// MaxSpendDefinition represents the maximum amount of credits that can be spent
// on a workflow or step. Users can specify multiple limits for different credit types
// (e.g. "CRE_CAP", "CRE_GAS"). Each limit consists of a credit type and maximum value.
type MaxSpendDefinition struct {
	Credit string
	Value  int64
}

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
	MaxSpends     []MaxSpendDefinition // Optional max spend limits for this step
}

type WorkflowSpec struct {
	Name      string
	Owner     string
	Triggers  []StepDefinition
	Actions   []StepDefinition
	Consensus []StepDefinition
	Targets   []StepDefinition
	MaxSpends []MaxSpendDefinition // Optional max spend limits for the entire workflow
}

func (w *WorkflowSpec) Steps() []StepDefinition {
	s := []StepDefinition{}
	s = append(s, w.Actions...)
	s = append(s, w.Consensus...)
	s = append(s, w.Targets...)
	return s
}
