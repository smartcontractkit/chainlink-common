// Code generated by pkg/capabilities/cli, DO NOT EDIT.

package streams

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func NewStreamsTriggerCapability(w *workflows.Workflow, ref string, cfg StreamsTriggerConfig) (StreamsTriggerCapability, error) {
    def := workflows.StepDefinition{
       ID: ref,
       Ref: ref,
       Inputs: workflows.StepInputs{
           Mapping: map[string]any{
           },
       },
       Config: map[string]any{
           "FeedIds": cfg.FeedIds,
           "MaxFrequencyMs": cfg.MaxFrequencyMs,
       },
       CapabilityType: capabilities.CapabilityTypeTrigger,
   }
    step := workflows.Step[StreamsTriggerOutputsElem]{Ref: ref, Definition: def}
    raw, err := workflows.AddStep(w, step)
    return &streamsTriggerCapability{CapabilityDefinition: raw}, err
}


type FeedIdCapability interface {
    workflows.CapabilityDefinition[FeedId]
    private()
}

type feedIdCapability struct {
    workflows.CapabilityDefinition[FeedId]
}


func (*feedIdCapability) private() {}

type StreamsTriggerCapability interface {
    workflows.CapabilityDefinition[StreamsTriggerOutputsElem]
    BenchmarkPrice() workflows.CapabilityDefinition[string]
    FeedId() FeedIdCapability
    FullReport() workflows.CapabilityDefinition[string]
    ObservationTimestamp() workflows.CapabilityDefinition[int]
    ReportContext() workflows.CapabilityDefinition[string]
    Signatures() workflows.CapabilityDefinition[[]string]
    private()
}

type streamsTriggerCapability struct {
    workflows.CapabilityDefinition[StreamsTriggerOutputsElem]
}


func (*streamsTriggerCapability) private() {}
func (c *streamsTriggerCapability) BenchmarkPrice() workflows.CapabilityDefinition[string] {
    return workflows.AccessField[StreamsTriggerOutputsElem, string](c.CapabilityDefinition, "BenchmarkPrice")
}
func (c *streamsTriggerCapability) FeedId() FeedIdCapability {
     return &feedIdCapability{ CapabilityDefinition: workflows.AccessField[StreamsTriggerOutputsElem, FeedId](c.CapabilityDefinition, "FeedId")}
}
func (c *streamsTriggerCapability) FullReport() workflows.CapabilityDefinition[string] {
    return workflows.AccessField[StreamsTriggerOutputsElem, string](c.CapabilityDefinition, "FullReport")
}
func (c *streamsTriggerCapability) ObservationTimestamp() workflows.CapabilityDefinition[int] {
    return workflows.AccessField[StreamsTriggerOutputsElem, int](c.CapabilityDefinition, "ObservationTimestamp")
}
func (c *streamsTriggerCapability) ReportContext() workflows.CapabilityDefinition[string] {
    return workflows.AccessField[StreamsTriggerOutputsElem, string](c.CapabilityDefinition, "ReportContext")
}
func (c *streamsTriggerCapability) Signatures() workflows.CapabilityDefinition[[]string] {
    return workflows.AccessField[StreamsTriggerOutputsElem, string](c.CapabilityDefinition, "Signatures")
}
