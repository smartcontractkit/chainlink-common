// Code generated by pkg/capabilities/cli, DO NOT EDIT.

package chainwritercap

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"

    "github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
    ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
)

func NewTarget(w *workflows.Workflow,id string, input TargetInput, cfg chainwriter.TargetConfig) {
    def := workflows.StepDefinition{
       ID: id,
       Inputs: workflows.StepInputs{
           Mapping: map[string]any{
               "signed_report": input.SignedReport.Ref(),
           },
       },
       Config: map[string]any{
           "address": cfg.Address,
           "deltaStage": cfg.DeltaStage,
           "schedule": cfg.Schedule,
       },
       CapabilityType: capabilities.CapabilityTypeTarget,
   }
    step := workflows.Step[struct{}]{Definition: def}
     workflows.AddStep(w, step)
    return
}


type TargetInput struct {
    SignedReport workflows.CapDefinition[ocr3.SignedReport]
}