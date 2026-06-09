package ocr3

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// OnchainKeyring2Genericless is a genericless counterpart of ocr3types.OnchainKeyring2. If generic RI does not matter,
// then it is more convenient to implement this interface and then use OnchainKeyring2ToGenericAdapter to use it as
// ocr3types.OnchainKeyring2.
type OnchainKeyring2Genericless interface {
	Sign(configDigest ocrtypes.ConfigDigest, seqNr uint64, report types.Report) (signature []byte, err error)
	Verify(publicKey ocrtypes.OnchainPublicKey, configDigest ocrtypes.ConfigDigest, seqNr uint64, report types.Report, signature []byte) bool
	Has(publicKey ocrtypes.OnchainPublicKey) bool
	MaxSignatureLength() int
	DebugIdentifier() string
}

var _ ocr3types.OnchainKeyring2[struct{}] = &OnchainKeyring2ToGenericAdapter[struct{}]{}

type OnchainKeyring2ToGenericAdapter[RI any] struct {
	K OnchainKeyring2Genericless
}

func (o *OnchainKeyring2ToGenericAdapter[RI]) Sign(c ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[RI]) (signature []byte, err error) {
	return o.K.Sign(c, seqNr, r.Report)
}

func (o *OnchainKeyring2ToGenericAdapter[RI]) Verify(pk ocrtypes.OnchainPublicKey, c ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[RI], signature []byte) bool {
	return o.K.Verify(pk, c, seqNr, r.Report, signature)
}

func (o *OnchainKeyring2ToGenericAdapter[RI]) Has(key ocrtypes.OnchainPublicKey) bool {
	return o.K.Has(key)
}

func (o *OnchainKeyring2ToGenericAdapter[RI]) MaxSignatureLength() int {
	return o.K.MaxSignatureLength()
}

func (o *OnchainKeyring2ToGenericAdapter[RI]) DebugIdentifier() string {
	return o.K.DebugIdentifier()
}
