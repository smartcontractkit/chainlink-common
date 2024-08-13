// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package notstreams

import (
    "github.com/smartcontractkit/chainlink-common/pkg/capabilities"
    "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func (cfg TriggerConfig) New(w *workflows.WorkflowSpecFactory,)TriggerCap { ref := "trigger"
    def := workflows.StepDefinition{
       ID: "notstreams@1.0.0",Ref: ref,
       Inputs: workflows.StepInputs{
       },
       Config: map[string]any{
           "maxFrequencyMs": cfg.MaxFrequencyMs,
       },
       CapabilityType: capabilities.CapabilityTypeTrigger,
   }
    step := workflows.Step[Feed]{Definition: def}
     raw := step.AddTo(w)
    return &trigger{CapDefinition: raw}
}


type TriggerCap interface {
    workflows.CapDefinition[Feed]
    FullReport() workflows.CapDefinition[string]
    Price() FeedPriceCap
    ReportContext() workflows.CapDefinition[string]
    Signatures() workflows.CapDefinition[[]string]
    Timestamp() workflows.CapDefinition[int]
    private()
}

type trigger struct {
    workflows.CapDefinition[Feed]
}


func (*trigger) private() {}
func (c *trigger) FullReport() workflows.CapDefinition[string] {
    return workflows.AccessField[Feed, string](c.CapDefinition, "FullReport")
}
func (c *trigger) Price() FeedPriceCap {
     return &feedPrice{ CapDefinition: workflows.AccessField[Feed, FeedPrice](c.CapDefinition, "Price")}
}
func (c *trigger) ReportContext() workflows.CapDefinition[string] {
    return workflows.AccessField[Feed, string](c.CapDefinition, "ReportContext")
}
func (c *trigger) Signatures() workflows.CapDefinition[[]string] {
    return workflows.AccessField[Feed, []string](c.CapDefinition, "Signatures")
}
func (c *trigger) Timestamp() workflows.CapDefinition[int] {
    return workflows.AccessField[Feed, int](c.CapDefinition, "Timestamp")
}

func NewTriggerFromFields(
                                                                        fullReport workflows.CapDefinition[string],
                                                                        price FeedPriceCap,
                                                                        reportContext workflows.CapDefinition[string],
                                                                        signatures workflows.CapDefinition[[]string],
                                                                        timestamp workflows.CapDefinition[int],) TriggerCap {
    return &simpleTrigger{
        CapDefinition: workflows.ComponentCapDefinition[Feed]{
        "fullReport": fullReport.Ref(),
        "price": price.Ref(),
        "reportContext": reportContext.Ref(),
        "signatures": signatures.Ref(),
        "timestamp": timestamp.Ref(),
        },
        fullReport: fullReport,
        price: price,
        reportContext: reportContext,
        signatures: signatures,
        timestamp: timestamp,
    }
}

type simpleTrigger struct {
    workflows.CapDefinition[Feed]
    fullReport workflows.CapDefinition[string]
    price FeedPriceCap
    reportContext workflows.CapDefinition[string]
    signatures workflows.CapDefinition[[]string]
    timestamp workflows.CapDefinition[int]
}
func (c *simpleTrigger) FullReport() workflows.CapDefinition[string] {
    return c.fullReport
}
func (c *simpleTrigger) Price() FeedPriceCap {
    return c.price
}
func (c *simpleTrigger) ReportContext() workflows.CapDefinition[string] {
    return c.reportContext
}
func (c *simpleTrigger) Signatures() workflows.CapDefinition[[]string] {
    return c.signatures
}
func (c *simpleTrigger) Timestamp() workflows.CapDefinition[int] {
    return c.timestamp
}

func (c *simpleTrigger) private() {}


type FeedPriceCap interface {
    workflows.CapDefinition[FeedPrice]
    PriceA() workflows.CapDefinition[string]
    PriceB() workflows.CapDefinition[string]
    private()
}

type feedPrice struct {
    workflows.CapDefinition[FeedPrice]
}


func (*feedPrice) private() {}
func (c *feedPrice) PriceA() workflows.CapDefinition[string] {
    return workflows.AccessField[FeedPrice, string](c.CapDefinition, "PriceA")
}
func (c *feedPrice) PriceB() workflows.CapDefinition[string] {
    return workflows.AccessField[FeedPrice, string](c.CapDefinition, "PriceB")
}

func NewFeedPriceFromFields(
                                                                        priceA workflows.CapDefinition[string],
                                                                        priceB workflows.CapDefinition[string],) FeedPriceCap {
    return &simpleFeedPrice{
        CapDefinition: workflows.ComponentCapDefinition[FeedPrice]{
        "priceA": priceA.Ref(),
        "priceB": priceB.Ref(),
        },
        priceA: priceA,
        priceB: priceB,
    }
}

type simpleFeedPrice struct {
    workflows.CapDefinition[FeedPrice]
    priceA workflows.CapDefinition[string]
    priceB workflows.CapDefinition[string]
}
func (c *simpleFeedPrice) PriceA() workflows.CapDefinition[string] {
    return c.priceA
}
func (c *simpleFeedPrice) PriceB() workflows.CapDefinition[string] {
    return c.priceB
}

func (c *simpleFeedPrice) private() {}

