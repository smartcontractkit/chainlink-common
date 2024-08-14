// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package streams

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func (cfg TriggerConfig) New(w *workflows.WorkflowSpecFactory,)workflows.CapDefinition[[]Feed] {
     ref := "trigger"
    def := workflows.StepDefinition{
       ID: "streams-trigger@1.0.0",Ref: ref,
       Inputs: workflows.StepInputs{
       },
       Config: map[string]any{
           "feedIds": cfg.FeedIds,
           "maxFrequencyMs": cfg.MaxFrequencyMs,
       },
       CapabilityType: capabilities.CapabilityTypeTrigger,
   }
    step := workflows.Step[[]Feed]{Definition: def}
     raw := step.AddTo(w)
     return raw
}


type TriggerCap interface {
    workflows.CapDefinition[Feed]
    BenchmarkPrice() workflows.CapDefinition[string]
    FeedId() FeedIdCap
    FullReport() workflows.CapDefinition[string]
    ObservationTimestamp() workflows.CapDefinition[int]
    ReportContext() workflows.CapDefinition[string]
    Signatures() workflows.CapDefinition[[]string]
    private()
}

type trigger struct {
    workflows.CapDefinition[Feed]
}

func (*trigger) private() {}
func (c *trigger) BenchmarkPrice() workflows.CapDefinition[string] {
    return workflows.AccessField[Feed, string](c.CapDefinition, "BenchmarkPrice")
}
func (c *trigger) FeedId() FeedIdCap {
     return FeedIdCap(workflows.AccessField[Feed, FeedId](c.CapDefinition, "FeedId"))
}
func (c *trigger) FullReport() workflows.CapDefinition[string] {
    return workflows.AccessField[Feed, string](c.CapDefinition, "FullReport")
}
func (c *trigger) ObservationTimestamp() workflows.CapDefinition[int] {
    return workflows.AccessField[Feed, int](c.CapDefinition, "ObservationTimestamp")
}
func (c *trigger) ReportContext() workflows.CapDefinition[string] {
    return workflows.AccessField[Feed, string](c.CapDefinition, "ReportContext")
}
func (c *trigger) Signatures() workflows.CapDefinition[[]string] {
    return workflows.AccessField[Feed, []string](c.CapDefinition, "Signatures")
}

func NewTriggerFromFields(
                                                                        benchmarkPrice workflows.CapDefinition[string],
                                                                        feedId FeedIdCap,
                                                                        fullReport workflows.CapDefinition[string],
                                                                        observationTimestamp workflows.CapDefinition[int],
                                                                        reportContext workflows.CapDefinition[string],
                                                                        signatures workflows.CapDefinition[[]string],) TriggerCap {
    return &simpleTrigger{
        CapDefinition: workflows.ComponentCapDefinition[Feed]{
        "benchmarkPrice": benchmarkPrice.Ref(),
        "feedId": feedId.Ref(),
        "fullReport": fullReport.Ref(),
        "observationTimestamp": observationTimestamp.Ref(),
        "reportContext": reportContext.Ref(),
        "signatures": signatures.Ref(),
        },
        benchmarkPrice: benchmarkPrice,
        feedId: feedId,
        fullReport: fullReport,
        observationTimestamp: observationTimestamp,
        reportContext: reportContext,
        signatures: signatures,
    }
}

type simpleTrigger struct {
    workflows.CapDefinition[Feed]
    benchmarkPrice workflows.CapDefinition[string]
    feedId FeedIdCap
    fullReport workflows.CapDefinition[string]
    observationTimestamp workflows.CapDefinition[int]
    reportContext workflows.CapDefinition[string]
    signatures workflows.CapDefinition[[]string]
}
func (c *simpleTrigger) BenchmarkPrice() workflows.CapDefinition[string] {
    return c.benchmarkPrice
}
func (c *simpleTrigger) FeedId() FeedIdCap {
    return c.feedId
}
func (c *simpleTrigger) FullReport() workflows.CapDefinition[string] {
    return c.fullReport
}
func (c *simpleTrigger) ObservationTimestamp() workflows.CapDefinition[int] {
    return c.observationTimestamp
}
func (c *simpleTrigger) ReportContext() workflows.CapDefinition[string] {
    return c.reportContext
}
func (c *simpleTrigger) Signatures() workflows.CapDefinition[[]string] {
    return c.signatures
}

func (c *simpleTrigger) private() {}


type FeedIdCap workflows.CapDefinition[FeedId]


type TriggerCap interface {
    workflows.CapDefinition[Trigger]
    Config() TriggerConfigCap
    Outputs() []FeedCap
    private()
}

type trigger struct {
    workflows.CapDefinition[Trigger]
}

func (*trigger) private() {}
func (c *trigger) Config() TriggerConfigCap {
     return TriggerConfigCap(workflows.AccessField[Trigger, TriggerConfig](c.CapDefinition, "Config"))
}
func (c *trigger) Outputs() FeedCap {
     return &feed{ CapDefinition: workflows.AccessField[Trigger, Feed](c.CapDefinition, "Outputs")}
}

func NewTriggerFromFields(
                                                                        config TriggerConfigCap,
                                                                        outputs []FeedCap,) TriggerCap {
    return &simpleTrigger{
        CapDefinition: workflows.ComponentCapDefinition[Trigger]{
        "config": config.Ref(),
        "outputs": outputs.Ref(),
        },
        config: config,
        outputs: outputs,
    }
}

type simpleTrigger struct {
    workflows.CapDefinition[Trigger]
    config TriggerConfigCap
    outputs []FeedCap
}
func (c *simpleTrigger) Config() TriggerConfigCap {
    return c.config
}
func (c *simpleTrigger) Outputs() FeedCap {
    return c.outputs
}

func (c *simpleTrigger) private() {}

