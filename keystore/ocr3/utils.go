package ocr3

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// OnchainKeyring2Genericless is a genericless counterpart of ocr3types.OnchainKeyring2. If generic RI does not matter,
// then it is more convenient to implement this interface and then use OnchainKeyring2ToGenericAdapter to use it as
// ocr3types.OnchainKeyring2.
type OnchainKeyring2 interface {
	Sign(configDigest ocrtypes.ConfigDigest, seqNr uint64, report types.Report) (signature []byte, err error)
	Verify(publicKey ocrtypes.OnchainPublicKey, configDigest ocrtypes.ConfigDigest, seqNr uint64, report types.Report, signature []byte) bool
	Has(publicKey ocrtypes.OnchainPublicKey) bool
	MaxSignatureLength() int
	DebugIdentifier() string
}

var _ ocr3types.OnchainKeyring2[struct{}] = &onchainKeyring2[struct{}]{}

type onchainKeyring2[RI any] struct {
	k OnchainKeyring2
}

func (o *onchainKeyring2[RI]) Sign(c ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[RI]) (signature []byte, err error) {
	return o.k.Sign(c, seqNr, r.Report)
}

func (o *onchainKeyring2[RI]) Verify(pk ocrtypes.OnchainPublicKey, c ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[RI], signature []byte) bool {
	return o.k.Verify(pk, c, seqNr, r.Report, signature)
}

func (o *onchainKeyring2[RI]) Has(key ocrtypes.OnchainPublicKey) bool {
	return o.k.Has(key)
}

func (o *onchainKeyring2[RI]) MaxSignatureLength() int {
	return o.k.MaxSignatureLength()
}

func (o *onchainKeyring2[RI]) DebugIdentifier() string {
	return o.k.DebugIdentifier()
}

func AsOCR3OnchainKeyring2[RI any](k OnchainKeyring2) ocr3types.OnchainKeyring2[RI] {
	return &onchainKeyring2[RI]{k}
}
