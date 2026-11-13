package ocr2key

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"slices"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

// MultichainPublicKey is an oracle's onchain public keys, keyed by chain type, as an OCR config lists a signer: for
// each chain type, its type, the key's length, then the key.
type MultichainPublicKey map[string]ocrtypes.OnchainPublicKey

func (mpk MultichainPublicKey) MarshalBinary() ([]byte, error) {
	var pubKeys [][]byte
	for k, pubKey := range mpk {
		typ, err := corekeys.ChainType(k).Type()
		if err != nil {
			// skipping unknown key type
			continue
		}
		buf := new(bytes.Buffer)
		if err = binary.Write(buf, binary.LittleEndian, typ); err != nil {
			return nil, err
		}
		length := len(pubKey)
		if length < 0 || length > math.MaxUint16 {
			return nil, errors.New("pubKey doesn't fit into uint16")
		}
		if err = binary.Write(buf, binary.LittleEndian, uint16(length)); err != nil {
			return nil, err
		}
		_, _ = buf.Write(pubKey)
		pubKeys = append(pubKeys, buf.Bytes())
	}
	// sort keys based on encoded type to make encoding deterministic
	slices.SortFunc(pubKeys, func(a, b []byte) int { return cmp.Compare(a[0], b[0]) })
	return bytes.Join(pubKeys, nil), nil
}

func (mpk *MultichainPublicKey) UnmarshalBinary(d []byte) error {
	keys := MultichainPublicKey{}
	buf := bytes.NewReader(d)

	for {
		// type
		typ, err := buf.ReadByte()
		if err != nil {
			return err
		}
		// length
		var length uint16
		err = binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return err
		}
		// value
		pubKey := make([]byte, length)
		n, err := buf.Read(pubKey)
		if err != nil {
			return err
		}
		if n != int(length) {
			return io.EOF
		}

		k, err := corekeys.NewChainType(typ)
		if err != nil {
			// skipping unknown key type
			continue
		}
		keys[string(k)] = pubKey

		if buf.Len() == 0 {
			break
		}
	}

	*mpk = keys
	return nil
}

// VerifyOCR3 checks a peer's signature on an OCR3 report signed with its family key. publicKey is the peer's multichain
// entry as the configuration carries it. Verifying needs no key of its own, so none is made: making one needs cgo for
// some families.
func VerifyOCR3(family string, publicKey ocrtypes.OnchainPublicKey, digest ocrtypes.ConfigDigest, seqNr uint64,
	report ocrtypes.Report, signature []byte) bool {
	factory, ok := keyBundleFactories[corekeys.ChainType(family)]
	if !ok {
		return false
	}

	var keys MultichainPublicKey
	if err := keys.UnmarshalBinary(publicKey); err != nil {
		return false
	}
	key, ok := keys[family]
	if !ok {
		return false
	}

	return factory.empty().Verify(key, ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: digest,
			Epoch:        uint32(seqNr), //nolint:gosec // seqNr maps to the uint32 epoch field
			Round:        0,
		},
		ExtraHash: [32]byte(make([]byte, 32)),
	}, report, signature)
}
