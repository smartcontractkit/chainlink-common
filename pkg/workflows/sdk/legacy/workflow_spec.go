package legacy

import (
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type StepInputs struct {
	OutputRef string
	Mapping   map[string]any
}

// StepDefinition is the parsed representation of a step in a workflow.
//
// Within the workflow spec, they are called "Capability Properties".
type StepDefinition struct {
	ID          string
	Ref         string
	Inputs      StepInputs
	Config      map[string]any
	ConfigProto *anypb.Any

	CapabilityType capabilities.CapabilityType
}

type WorkflowSpec struct {
	Name      string
	Owner     string
	Triggers  []StepDefinition
	Actions   []StepDefinition
	Consensus []StepDefinition
	Targets   []StepDefinition
}

func (w *WorkflowSpec) Steps() []StepDefinition {
	s := []StepDefinition{}
	s = append(s, w.Actions...)
	s = append(s, w.Consensus...)
	s = append(s, w.Targets...)
	return s
}
