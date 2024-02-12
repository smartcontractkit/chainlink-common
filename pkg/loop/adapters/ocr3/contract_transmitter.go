package ocr3

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.ContractTransmitter[any] = (*ContractTransmitter)(nil)

type ContractTransmitter struct {
	c ocrtypes.ContractTransmitter
}

func NewContractTransmitter(c ocrtypes.ContractTransmitter) ContractTransmitter {
	return ContractTransmitter{c}
}

func (c ContractTransmitter) Transmit(ctx context.Context, digest ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[any], signatures []ocrtypes.AttributedOnchainSignature) error {
	return c.c.Transmit(ctx, ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: digest,
			Epoch:        uint32(seqNr),
			Round:        0,
		},
		ExtraHash: [32]byte(make([]byte, 32)),
	}, r.Report, signatures)
}

func (c ContractTransmitter) FromAccount() (ocrtypes.Account, error) {
	return c.c.FromAccount()
}
