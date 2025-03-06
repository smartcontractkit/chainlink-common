// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package ocr3cap

import (
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
)

// EncoderWrapper allows access to field from an sdk.CapDefinition[Encoder]
func EncoderWrapper(raw sdk.CapDefinition[Encoder]) EncoderCap {
	wrapped, ok := raw.(EncoderCap)
	if ok {
		return wrapped
	}
	return EncoderCap(raw)
}

type EncoderCap sdk.CapDefinition[Encoder]

// EncoderConfigWrapper allows access to field from an sdk.CapDefinition[EncoderConfig]
func EncoderConfigWrapper(raw sdk.CapDefinition[EncoderConfig]) EncoderConfigCap {
	wrapped, ok := raw.(EncoderConfigCap)
	if ok {
		return wrapped
	}
	return EncoderConfigCap(raw)
}

type EncoderConfigCap sdk.CapDefinition[EncoderConfig]

// KeyIdWrapper allows access to field from an sdk.CapDefinition[KeyId]
func KeyIdWrapper(raw sdk.CapDefinition[KeyId]) KeyIdCap {
	wrapped, ok := raw.(KeyIdCap)
	if ok {
		return wrapped
	}
	return KeyIdCap(raw)
}

type KeyIdCap sdk.CapDefinition[KeyId]

// ReportIdWrapper allows access to field from an sdk.CapDefinition[ReportId]
func ReportIdWrapper(raw sdk.CapDefinition[ReportId]) ReportIdCap {
	wrapped, ok := raw.(ReportIdCap)
	if ok {
		return wrapped
	}
	return ReportIdCap(raw)
}

type ReportIdCap sdk.CapDefinition[ReportId]

// SignedReportWrapper allows access to field from an sdk.CapDefinition[SignedReport]
func SignedReportWrapper(raw sdk.CapDefinition[SignedReport]) SignedReportCap {
	wrapped, ok := raw.(SignedReportCap)
	if ok {
		return wrapped
	}
	return &signedReportCap{CapDefinition: raw}
}

type SignedReportCap interface {
	sdk.CapDefinition[SignedReport]
	Context() sdk.CapDefinition[[]uint8]
	ID() sdk.CapDefinition[[]uint8]
	Report() sdk.CapDefinition[[]uint8]
	Signatures() sdk.CapDefinition[[][]uint8]
	private()
}

type signedReportCap struct {
	sdk.CapDefinition[SignedReport]
}

func (*signedReportCap) private() {}
func (c *signedReportCap) Context() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[SignedReport, []uint8](c.CapDefinition, "Context")
}
func (c *signedReportCap) ID() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[SignedReport, []uint8](c.CapDefinition, "ID")
}
func (c *signedReportCap) Report() sdk.CapDefinition[[]uint8] {
	return sdk.AccessField[SignedReport, []uint8](c.CapDefinition, "Report")
}
func (c *signedReportCap) Signatures() sdk.CapDefinition[[][]uint8] {
	return sdk.AccessField[SignedReport, [][]uint8](c.CapDefinition, "Signatures")
}

func ConstantSignedReport(value SignedReport) SignedReportCap {
	return &signedReportCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewSignedReportFromFields(
	context sdk.CapDefinition[[]uint8],
	iD sdk.CapDefinition[[]uint8],
	report sdk.CapDefinition[[]uint8],
	signatures sdk.CapDefinition[[][]uint8]) SignedReportCap {
	return &simpleSignedReport{
		CapDefinition: sdk.ComponentCapDefinition[SignedReport]{
			"Context":    context.Ref(),
			"ID":         iD.Ref(),
			"Report":     report.Ref(),
			"Signatures": signatures.Ref(),
		},
		context:    context,
		iD:         iD,
		report:     report,
		signatures: signatures,
	}
}

type simpleSignedReport struct {
	sdk.CapDefinition[SignedReport]
	context    sdk.CapDefinition[[]uint8]
	iD         sdk.CapDefinition[[]uint8]
	report     sdk.CapDefinition[[]uint8]
	signatures sdk.CapDefinition[[][]uint8]
}

func (c *simpleSignedReport) Context() sdk.CapDefinition[[]uint8] {
	return c.context
}
func (c *simpleSignedReport) ID() sdk.CapDefinition[[]uint8] {
	return c.iD
}
func (c *simpleSignedReport) Report() sdk.CapDefinition[[]uint8] {
	return c.report
}
func (c *simpleSignedReport) Signatures() sdk.CapDefinition[[][]uint8] {
	return c.signatures
}

func (c *simpleSignedReport) private() {}
