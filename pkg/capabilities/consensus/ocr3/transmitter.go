package ocr3

import (
	"context"
	"errors"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// ErrDeprecated is returned by every method of ContractTransmitter.
var ErrDeprecated = errors.New("the OCR3 consensus capability has been removed from chainlink-common; this is a stub that no longer transmits")

var _ ocr3types.ContractTransmitter[[]byte] = (*ContractTransmitter)(nil)

// ContractTransmitter used to forward a report and its signatures back to the
// OCR3 consensus capability. That capability has been removed, so this type is
// retained only so that existing wiring keeps compiling: every method fails
// with [ErrDeprecated].
//
// Deprecated: remove the surrounding OCR3 capability provider instead of
// calling this.
type ContractTransmitter struct {
	lggr logger.Logger
}

// NewContractTransmitter returns a non-functional [ContractTransmitter].
//
// Deprecated: see [ContractTransmitter].
func NewContractTransmitter(lggr logger.Logger, _ core.CapabilitiesRegistry, _ string) *ContractTransmitter {
	return &ContractTransmitter{lggr: logger.Named(lggr, "DeprecatedOCR3ContractTransmitter")}
}

// Transmit always fails with [ErrDeprecated].
//
// Deprecated: see [ContractTransmitter].
func (c *ContractTransmitter) Transmit(_ context.Context, _ types.ConfigDigest, _ uint64, _ ocr3types.ReportWithInfo[[]byte], _ []types.AttributedOnchainSignature) error {
	c.lggr.Error(ErrDeprecated.Error())
	return ErrDeprecated
}

// FromAccount always fails with [ErrDeprecated].
//
// Deprecated: see [ContractTransmitter].
func (c *ContractTransmitter) FromAccount(_ context.Context) (types.Account, error) {
	return "", ErrDeprecated
}
