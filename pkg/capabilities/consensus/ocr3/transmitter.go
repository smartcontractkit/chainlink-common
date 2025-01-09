package ocr3

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	"google.golang.org/protobuf/types/known/structpb"
)

var _ ocr3types.ContractTransmitter[[]byte] = (*ContractTransmitter)(nil)

// ContractTransmitter is a custom transmitter for the OCR3 capability.
// When called it will forward the report + its signatures back to the
// OCR3 capability by making a call to Execute with a special "method"
// parameter.
type ContractTransmitter struct {
	lggr        logger.Logger
	registry    core.CapabilitiesRegistry
	capability  capabilities.ExecutableCapability
	fromAccount string
	emitter     custmsg.MessageEmitter
}

func extractReportInfo(data []byte) (*pbtypes.ReportInfo, error) {
	info := &structpb.Struct{}
	err := proto.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}

	im := info.AsMap()
	ri, ok := im["reportInfo"]
	if !ok {
		return nil, errors.New("could not fetch reportInfo from structpb")
	}

	ris, ok := ri.(string)
	if !ok {
		return nil, errors.New("reportInfo is not bytes")
	}

	rib, err := base64.StdEncoding.DecodeString(ris)
	if err != nil {
		return nil, err
	}

	reportInfo := &pbtypes.ReportInfo{}
	err = proto.Unmarshal(rib, reportInfo)
	return reportInfo, err
}

func (c *ContractTransmitter) Transmit(ctx context.Context, configDigest types.ConfigDigest, seqNr uint64, rwi ocr3types.ReportWithInfo[[]byte], signatures []types.AttributedOnchainSignature) error {
	info, err := extractReportInfo(rwi.Info)
	if err != nil {
		c.lggr.Error("could not unmarshal info")
		return err
	}

	signedReport := &pbtypes.SignedReport{}
	if info.ShouldReport {
		signedReport.Report = rwi.Report

		// report context is the config digest + the sequence number padded with zeros
		// (see OCR3OnchainKeyringAdapter in core)
		seqToEpoch := make([]byte, 32)
		binary.BigEndian.PutUint32(seqToEpoch[32-5:32-1], uint32(seqNr))
		zeros := make([]byte, 32)
		repContext := append(append(configDigest[:], seqToEpoch[:]...), zeros...)
		signedReport.Context = repContext

		var sigs [][]byte
		for _, s := range signatures {
			sigs = append(sigs, s.Signature)
		}
		signedReport.Signatures = sigs
		reportIDBytes, err2 := hex.DecodeString(info.Id.ReportId)
		if err2 != nil {
			return fmt.Errorf("could not decode report id: %w", err2)
		}
		signedReport.ID = reportIDBytes
		c.lggr.Debugw("ContractTransmitter added signatures and context", "nSignatures", len(sigs), "contextLen", len(repContext))
	}

	resp := map[string]any{
		methodHeader:       methodSendResponse,
		transmissionHeader: signedReport,
		terminateHeader:    !info.ShouldReport,
	}
	inputs, err := values.Wrap(resp)
	if err != nil {
		c.lggr.Error("could not wrap report", "payload", resp)
		return err
	}

	c.lggr.Debugw("ContractTransmitter transmitting", "shouldReport", info.ShouldReport, "len", len(rwi.Report))
	if c.capability == nil {
		cp, innerErr := c.registry.Get(ctx, ocrCapabilityID)
		if innerErr != nil {
			return fmt.Errorf("failed to fetch ocr3 capability from registry: %w", innerErr)
		}

		c.capability = cp.(capabilities.ExecutableCapability)
	}

	msg := "report with id " + info.Id.ReportId + " should be reported: " + fmt.Sprint(info.ShouldReport)
	err = c.emitter.With(
		"workflowExecutionID", info.Id.WorkflowExecutionId,
		"workflowID", info.Id.WorkflowId,
		"workflowOwner", info.Id.WorkflowOwner,
		"workflowName", info.Id.WorkflowName,
		"reportId", info.Id.ReportId,
	).Emit(ctx, msg)
	if err != nil {
		c.lggr.Errorw(fmt.Sprintf("could not emit message: %s", msg), "error", err)
	}

	_, err = c.capability.Execute(ctx, capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowExecutionID: info.Id.WorkflowExecutionId,
			WorkflowID:          info.Id.WorkflowId,
			WorkflowDonID:       info.Id.WorkflowDonId,
		},
		Inputs: inputs.(*values.Map),
	})
	if err != nil {
		c.lggr.Errorw("could not transmit response", "error", err, "weid", info.Id.WorkflowExecutionId)
	}
	c.lggr.Debugw("ContractTransmitter transmitting done", "shouldReport", info.ShouldReport, "len", len(rwi.Report))
	return err
}

func (c *ContractTransmitter) FromAccount(_ context.Context) (types.Account, error) {
	return types.Account(c.fromAccount), nil
}

func NewContractTransmitter(lggr logger.Logger, registry core.CapabilitiesRegistry, fromAccount string) *ContractTransmitter {
	return &ContractTransmitter{lggr: lggr, registry: registry, fromAccount: fromAccount, emitter: custmsg.NewLabeler()}
}
