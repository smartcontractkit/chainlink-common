package ocr2key

import (
	"io"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// keccakEd25519Keyring signs OCR reports with ed25519 (like ed25519Keyring) over the canonical keccak256 report digest.
type keccakEd25519Keyring struct {
	*ed25519Keyring
}

func newKeccakEd25519Keyring(material io.Reader) (*keccakEd25519Keyring, error) {
	base, err := newEd25519Keyring(material)
	if err != nil {
		return nil, err
	}
	return &keccakEd25519Keyring{ed25519Keyring: base}, nil
}

func (kkr *keccakEd25519Keyring) Sign(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	return kkr.SignBlob(ReportToSigData(reportCtx, report))
}

func (kkr *keccakEd25519Keyring) Sign3(digest types.ConfigDigest, seqNr uint64, r ocrtypes.Report) (signature []byte, err error) {
	return kkr.SignBlob(ReportToSigData3(digest, seqNr, r))
}

func (kkr *keccakEd25519Keyring) Verify(publicKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report ocrtypes.Report, signature []byte) bool {
	return kkr.VerifyBlob(publicKey, ReportToSigData(reportCtx, report), signature)
}

func (kkr *keccakEd25519Keyring) Verify3(publicKey ocrtypes.OnchainPublicKey, cd ocrtypes.ConfigDigest, seqNr uint64, r ocrtypes.Report, signature []byte) bool {
	return kkr.VerifyBlob(publicKey, ReportToSigData3(cd, seqNr, r), signature)
}
