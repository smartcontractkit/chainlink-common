// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package basictarget

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)



func (cfg TargetConfig) New(w *workflows.WorkflowSpecFactory, input TargetInput) {
    
    def := workflows.StepDefinition{
       ID: "basic-test-target@1.0.0",
       Inputs: input.ToSteps(),
       Config: map[string]any{
           "name": cfg.Name,
           "number": cfg.Number,
       },
       CapabilityType: capabilities.CapabilityTypeTarget,
   }


    step := workflows.Step[struct{}]{Definition: def}
    step.AddTo(w)
}


type TargetInput struct {
    CoolInput workflows.CapDefinition[string]
}

func (input TargetInput) ToSteps() workflows.StepInputs {
    return workflows.StepInputs{
       Mapping: map[string]any{
        "cool_input,omitempty": input.CoolInput.Ref(),
       },
   }
}