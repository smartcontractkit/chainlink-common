package ocr2key

import (
	"errors"
	"fmt"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

// OCR3Keyring is one key bundle as an OCR3 onchain keyring.
//
// A key bundle is an OCR2 keyring: it signs a ReportContext and knows nothing
// about sequence numbers, and its public key is the bare key of its own chain.
// An OCR3 oracle needs the other side of both - the round expressed as a report
// context (see OCR3ReportContext), and the public key in the multichain form a
// configuration lists members by, even when the member signs for one family.
//
// It exists alongside a multi-bundle adapter rather than being replaced by one:
// choosing between bundles needs the report to say which family it is for, and a
// process holding a single bundle has no such report and no such choice. So this
// signs whatever it is handed.
//
// Only the onchain half. KeyBundle is already an ocrtypes.OffchainKeyring, so an
// oracle needing both passes the bundle itself for the offchain one.
type OCR3Keyring struct {
	bundle    KeyBundle
	family    string
	publicKey ocrtypes.OnchainPublicKey
}

var _ ocr3types.OnchainKeyring[[]byte] = (*OCR3Keyring)(nil)

// NewOCR3Keyring returns bundle as an OCR3 onchain keyring, announcing itself
// under family.
//
// family is taken rather than read off the bundle's ChainType so that a caller
// naming its bundles the way a configuration does - which is what the encoding is
// keyed by - is not overridden by what the key happens to be.
//
// The public key is encoded once here, so a bundle whose family cannot be encoded
// fails where the keyring is built rather than as an oracle no configuration
// recognises.
func NewOCR3Keyring(family string, bundle KeyBundle) (*OCR3Keyring, error) {
	if bundle == nil {
		return nil, errors.New("no key bundle to sign with")
	}
	publicKey, err := MarshalMultichainPublicKey(map[string]ocrtypes.OnchainPublicKey{family: bundle.PublicKey()})
	if err != nil {
		return nil, fmt.Errorf("failed to encode the %s onchain public key: %w", family, err)
	}
	if len(publicKey) == 0 {
		return nil, fmt.Errorf("%q is not a known signing family, so a keyring for it would announce nothing", family)
	}
	return &OCR3Keyring{bundle: bundle, family: family, publicKey: publicKey}, nil
}

// PublicKey is the multichain-encoded form, which is what a configuration lists
// this oracle as and therefore what libocr compares against.
func (k *OCR3Keyring) PublicKey() ocrtypes.OnchainPublicKey { return k.publicKey }

func (k *OCR3Keyring) Sign(digest ocrtypes.ConfigDigest, seqNr uint64, report ocr3types.ReportWithInfo[[]byte]) ([]byte, error) {
	return k.bundle.Sign(OCR3ReportContext(digest, seqNr), report.Report)
}

// Verify checks a peer's signature. publicKey is that peer's entry as the
// configuration carries it, so this family's key is taken out of it first.
func (k *OCR3Keyring) Verify(
	publicKey ocrtypes.OnchainPublicKey,
	digest ocrtypes.ConfigDigest,
	seqNr uint64,
	report ocr3types.ReportWithInfo[[]byte],
	signature []byte,
) bool {
	key, err := OnchainPublicKeyFor(k.family, publicKey)
	if err != nil {
		return false
	}
	return k.bundle.Verify(key, OCR3ReportContext(digest, seqNr), report.Report, signature)
}

func (k *OCR3Keyring) MaxSignatureLength() int { return k.bundle.MaxSignatureLength() }

// OnchainPublicKeyFor returns one family's entry from a multichain-encoded public
// key.
//
// A key with no entry for the family is an error rather than an empty result: it
// is a peer this keyring cannot check the signatures of, which is not the same as
// a peer whose signature is wrong.
func OnchainPublicKeyFor(family string, encoded ocrtypes.OnchainPublicKey) (ocrtypes.OnchainPublicKey, error) {
	if _, err := corekeys.ChainType(family).Type(); err != nil {
		return nil, fmt.Errorf("%q is not a known signing family: %w", family, err)
	}
	keys, err := UnmarshalMultichainPublicKey(encoded)
	if err != nil {
		return nil, err
	}
	key, ok := keys[family]
	if !ok {
		return nil, fmt.Errorf("onchain public key has no %s entry", family)
	}
	return key, nil
}
