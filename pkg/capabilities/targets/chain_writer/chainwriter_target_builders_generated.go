// Code generated by pkg/capabilities/cli, DO NOT EDIT.

package chainwriter

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func NewChainwriterTargetCapability(w *workflows.Workflow, ref string, input ChainwriterTargetCapabilityInput, cfg ChainwriterTargetConfig) (error) {
    def := workflows.StepDefinition{
       ID: ref,
       Ref: ref,
       Inputs: workflows.StepInputs{
           Mapping: map[string]any{
               "SignedReport": input.SignedReport,
           },
       },
       Config: map[string]any{
           "AggregationConfig": cfg.AggregationConfig,
           "AggregationMethod": cfg.AggregationMethod,
           "Encoder": cfg.Encoder,
           "EncoderConfig": cfg.EncoderConfig,
           "ReportId": cfg.ReportId,
       },
       CapabilityType: capabilities.CapabilityTypeTarget,
   }
    step := workflows.Step[ struct{}]{Ref: ref, Definition: def}
     _, err := workflows.AddStep(w, step)
    return err
}


type ChainwriterTargetOutputCapability interface {
    workflows.CapabilityDefinition[ChainwriterTargetOutput]
    Observations() ChainwriterTargetOutputObservationsCapability
    private()
}

type chainwriterTargetOutputCapability struct {
    workflows.CapabilityDefinition[ChainwriterTargetOutput]
}


func (*chainwriterTargetOutputCapability) private() {}
func (c *chainwriterTargetOutputCapability) Observations() ChainwriterTargetOutputObservationsCapability {
     return &chainwriterTargetOutputObservationsCapability{ CapabilityDefinition: workflows.AccessField[ChainwriterTargetOutput, ChainwriterTargetOutputObservations](c.CapabilityDefinition, "Observations")}
}

type ChainwriterTargetOutputObservationsCapability interface {
    workflows.CapabilityDefinition[ChainwriterTargetOutputObservations]
    Underlying() workflows.CapabilityDefinition[[]interface{}]
    private()
}

type chainwriterTargetOutputObservationsCapability struct {
    workflows.CapabilityDefinition[ChainwriterTargetOutputObservations]
}


func (*chainwriterTargetOutputObservationsCapability) private() {}
func (c *chainwriterTargetOutputObservationsCapability) Underlying() workflows.CapabilityDefinition[[]interface{}] {
    return workflows.AccessField[ChainwriterTargetOutputObservations, []interface{}](c.CapabilityDefinition, "Underlying")
}

type ChainwriterTargetCapabilityInput struct {
    SignedReport workflows.CapabilityDefinition[ChainwriterTargetInputsSignedReport]
}