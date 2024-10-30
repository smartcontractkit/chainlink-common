// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package notstreams

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func (cfg TriggerConfig) New(w *sdk.WorkflowSpecFactory) FeedCap {
	ref := "trigger"
	def := sdk.StepDefinition{
		ID: "notstreams@1.0.0", Ref: ref,
		Inputs: sdk.StepInputs{},
		Config: map[string]any{
			"maxFrequencyMs": cfg.MaxFrequencyMs,
		},
		CapabilityType: capabilities.CapabilityTypeTrigger,
	}

	step := sdk.Step[Feed]{Definition: def}
	raw := step.AddTo(w)
	return FeedWrapper(raw)
}

// FeedWrapper allows access to field from an sdk.CapDefinition[Feed]
func FeedWrapper(raw sdk.CapDefinition[Feed]) FeedCap {
	wrapped, ok := raw.(FeedCap)
	if ok {
		return wrapped
	}
	return &feedCap{CapDefinition: raw}
}

type FeedCap interface {
	sdk.CapDefinition[Feed]
	Metadata() SignerMetadataCap
	Payload() FeedReportCap
	Timestamp() sdk.CapDefinition[int64]
	private()
}

type feedCap struct {
	sdk.CapDefinition[Feed]
}

func (*feedCap) private() {}
func (c *feedCap) Metadata() SignerMetadataCap {
	return SignerMetadataWrapper(sdk.AccessField[Feed, SignerMetadata](c.CapDefinition, "Metadata"))
}
func (c *feedCap) Payload() FeedReportCap {
	return FeedReportWrapper(sdk.AccessField[Feed, FeedReport](c.CapDefinition, "Payload"))
}
func (c *feedCap) Timestamp() sdk.CapDefinition[int64] {
	return sdk.AccessField[Feed, int64](c.CapDefinition, "Timestamp")
}

func ConstantFeed(value Feed) FeedCap {
	return &feedCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewFeedFromFields(
	metadata SignerMetadataCap,
	payload FeedReportCap,
	timestamp sdk.CapDefinition[int64]) FeedCap {
	return &simpleFeed{
		CapDefinition: sdk.ComponentCapDefinition[Feed]{
			"Metadata":  metadata.Ref(),
			"Payload":   payload.Ref(),
			"Timestamp": timestamp.Ref(),
		},
		metadata:  metadata,
		payload:   payload,
		timestamp: timestamp,
	}
}

type simpleFeed struct {
	sdk.CapDefinition[Feed]
	metadata  SignerMetadataCap
	payload   FeedReportCap
	timestamp sdk.CapDefinition[int64]
}

func (c *simpleFeed) Metadata() SignerMetadataCap {
	return c.metadata
}
func (c *simpleFeed) Payload() FeedReportCap {
	return c.payload
}
func (c *simpleFeed) Timestamp() sdk.CapDefinition[int64] {
	return c.timestamp
}

func (c *simpleFeed) private() {}

// FeedReportWrapper allows access to field from an sdk.CapDefinition[FeedReport]
func FeedReportWrapper(raw sdk.CapDefinition[FeedReport]) FeedReportCap {
	wrapped, ok := raw.(FeedReportCap)
	if ok {
		return wrapped
	}
	return &feedReportCap{CapDefinition: raw}
}

type FeedReportCap interface {
	sdk.CapDefinition[FeedReport]
	BuyPrice() sdk.CapDefinition[[]uint8]
	FullReport() sdk.CapDefinition[[]uint8]
	ObservationTimestamp() sdk.CapDefinition[int64]
	ReportContext() sdk.CapDefinition[[]uint8]
	SellPrice() sdk.CapDefinition[[]uint8]
	Signature() sdk.CapDefinition[[]uint8]
	private()
}

type feedReportCap struct {
	sdk.CapDefinition[FeedReport]
}

func (*feedReportCap) private() {}
func (c *feedReportCap) BuyPrice() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[FeedReport, []uint8](c.CapDefinition, "BuyPrice")
}
func (c *feedReportCap) FullReport() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[FeedReport, []uint8](c.CapDefinition, "FullReport")
}
func (c *feedReportCap) ObservationTimestamp() sdk.CapDefinition[int64] {
	return sdk.AccessField[FeedReport, int64](c.CapDefinition, "ObservationTimestamp")
}
func (c *feedReportCap) ReportContext() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[FeedReport, []uint8](c.CapDefinition, "ReportContext")
}
func (c *feedReportCap) SellPrice() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[FeedReport, []uint8](c.CapDefinition, "SellPrice")
}
func (c *feedReportCap) Signature() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[FeedReport, []uint8](c.CapDefinition, "Signature")
}

func ConstantFeedReport(value FeedReport) FeedReportCap {
	return &feedReportCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewFeedReportFromFields(
	buyPrice sdk.CapDefinition[[]uint8],
	fullReport sdk.CapDefinition[[]uint8],
	observationTimestamp sdk.CapDefinition[int64],
	reportContext sdk.CapDefinition[[]uint8],
	sellPrice sdk.CapDefinition[[]uint8],
	signature sdk.CapDefinition[[]uint8]) FeedReportCap {
	return &simpleFeedReport{
		CapDefinition: sdk.ComponentCapDefinition[FeedReport]{
			"BuyPrice":             buyPrice.Ref(),
			"FullReport":           fullReport.Ref(),
			"ObservationTimestamp": observationTimestamp.Ref(),
			"ReportContext":        reportContext.Ref(),
			"SellPrice":            sellPrice.Ref(),
			"Signature":            signature.Ref(),
		},
		buyPrice:             buyPrice,
		fullReport:           fullReport,
		observationTimestamp: observationTimestamp,
		reportContext:        reportContext,
		sellPrice:            sellPrice,
		signature:            signature,
	}
}

type simpleFeedReport struct {
	sdk.CapDefinition[FeedReport]
	buyPrice             sdk.CapDefinition[[]uint8]
	fullReport           sdk.CapDefinition[[]uint8]
	observationTimestamp sdk.CapDefinition[int64]
	reportContext        sdk.CapDefinition[[]uint8]
	sellPrice            sdk.CapDefinition[[]uint8]
	signature            sdk.CapDefinition[[]uint8]
}

func (c *simpleFeedReport) BuyPrice() sdk.CapDefinition[[]uint8] {
	return c.buyPrice
}
func (c *simpleFeedReport) FullReport() sdk.CapDefinition[[]uint8] {
	return c.fullReport
}
func (c *simpleFeedReport) ObservationTimestamp() sdk.CapDefinition[int64] {
	return c.observationTimestamp
}
func (c *simpleFeedReport) ReportContext() sdk.CapDefinition[[]uint8] {
	return c.reportContext
}
func (c *simpleFeedReport) SellPrice() sdk.CapDefinition[[]uint8] {
	return c.sellPrice
}
func (c *simpleFeedReport) Signature() sdk.CapDefinition[[]uint8] {
	return c.signature
}

func (c *simpleFeedReport) private() {}

// SignerMetadataWrapper allows access to field from an sdk.CapDefinition[SignerMetadata]
func SignerMetadataWrapper(raw sdk.CapDefinition[SignerMetadata]) SignerMetadataCap {
	wrapped, ok := raw.(SignerMetadataCap)
	if ok {
		return wrapped
	}
	return &signerMetadataCap{CapDefinition: raw}
}

type SignerMetadataCap interface {
	sdk.CapDefinition[SignerMetadata]
	Signer() sdk.CapDefinition[string]
	private()
}

type signerMetadataCap struct {
	sdk.CapDefinition[SignerMetadata]
}

func (*signerMetadataCap) private() {}
func (c *signerMetadataCap) Signer() sdk.CapDefinition[string] {
	return sdk.AccessField[SignerMetadata, string](c.CapDefinition, "Signer")
}

func ConstantSignerMetadata(value SignerMetadata) SignerMetadataCap {
	return &signerMetadataCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewSignerMetadataFromFields(
	signer sdk.CapDefinition[string]) SignerMetadataCap {
	return &simpleSignerMetadata{
		CapDefinition: sdk.ComponentCapDefinition[SignerMetadata]{
			"Signer": signer.Ref(),
		},
		signer: signer,
	}
}

type simpleSignerMetadata struct {
	sdk.CapDefinition[SignerMetadata]
	signer sdk.CapDefinition[string]
}

func (c *simpleSignerMetadata) Signer() sdk.CapDefinition[string] {
	return c.signer
}

func (c *simpleSignerMetadata) private() {}
