package ocr3

import (
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.OnchainKeyring[any] = (*OnchainKeyring)(nil)

type OnchainKeyring struct {
	o ocrtypes.OnchainKeyring
}

func NewOnchainKeyring(o ocrtypes.OnchainKeyring) OnchainKeyring {
	return OnchainKeyring{o}
}

func (k *OnchainKeyring) PublicKey() ocrtypes.OnchainPublicKey {
	return k.o.PublicKey()
}

func (k *OnchainKeyring) Sign(digest ocrtypes.ConfigDigest, seqNr uint64, r ocr3types.ReportWithInfo[any]) (signature []byte, err error) {
	return k.o.Sign(ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: digest,
			Epoch:        uint32(seqNr),
			Round:        0,
		},
		ExtraHash: [32]byte(make([]byte, 32)),
	}, r.Report)
}

func (k *OnchainKeyring) Verify(opk ocrtypes.OnchainPublicKey, digest ocrtypes.ConfigDigest, seqNr uint64, ri ocr3types.ReportWithInfo[any], signature []byte) bool {
	return k.o.Verify(opk, ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: digest,
			Epoch:        uint32(seqNr),
			Round:        0,
		},
		ExtraHash: [32]byte(make([]byte, 32)),
	}, ri.Report, signature)
}

func (k *OnchainKeyring) MaxSignatureLength() int {
	return k.o.MaxSignatureLength()
}
