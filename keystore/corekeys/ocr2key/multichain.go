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

// A multichain onchain public key is the identity of an oracle that may sign for
// more than one chain family: one length-prefixed entry per family, ordered by
// family, rather than the bare key of whichever chain it happens to sign for.
//
// It lives here rather than beside any one consumer because it is a wire format
// with several: whoever writes an OCR configuration encodes it, every oracle
// joining that configuration reports its own in the same form, and libocr
// compares the two byte for byte to decide membership. Two implementations of it
// are two chances for an oracle to be silently unrecognised.
//
// The entry layout is: family (1 byte, corekeys.ChainType.Type), length (uint16,
// little endian), key.

// MarshalMultichainKeyBundle encodes the public keys of the bundles an oracle
// signs with, keyed by chain family.
func MarshalMultichainKeyBundle(bundles map[string]KeyBundle) (ocrtypes.OnchainPublicKey, error) {
	pubKeys := make(map[string]ocrtypes.OnchainPublicKey, len(bundles))
	for family, bundle := range bundles {
		pubKeys[family] = ocrtypes.OnchainPublicKey(bundle.PublicKey())
	}
	return MarshalMultichainPublicKey(pubKeys)
}

// MarshalMultichainPublicKey encodes one entry per family.
//
// An unknown family is skipped rather than rejected, so a caller holding a key
// for something this build does not know about still produces the entries it
// does know - which is what the configurations already in use were written with.
func MarshalMultichainPublicKey(keys map[string]ocrtypes.OnchainPublicKey) (ocrtypes.OnchainPublicKey, error) {
	var entries [][]byte
	for family, pubKey := range keys {
		typ, err := corekeys.ChainType(family).Type()
		if err != nil {
			// skipping unknown key type
			continue
		}
		if len(pubKey) > math.MaxUint16 {
			return nil, errors.New("pubKey doesn't fit into uint16")
		}

		entry := make([]byte, 0, 3+len(pubKey))
		entry = append(entry, typ)
		entry = binary.LittleEndian.AppendUint16(entry, uint16(len(pubKey)))
		entry = append(entry, pubKey...)
		entries = append(entries, entry)
	}
	// sort keys based on encoded type to make encoding deterministic
	slices.SortFunc(entries, func(a, b []byte) int { return cmp.Compare(a[0], b[0]) })
	return bytes.Join(entries, nil), nil
}

// UnmarshalMultichainPublicKey reads the entries back, keyed by chain family.
//
// An unknown family is skipped for the same reason Marshal skips it: the rest of
// the key is still this oracle's identity, and a family this build cannot verify
// signatures for is one it has no use for anyway.
func UnmarshalMultichainPublicKey(d []byte) (map[string]ocrtypes.OnchainPublicKey, error) {
	m := map[string]ocrtypes.OnchainPublicKey{}
	buf := bytes.NewReader(d)

	for {
		// type
		typ, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		// length
		var length uint16
		err = binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return nil, err
		}
		// value
		pubKey := make([]byte, length)
		n, err := buf.Read(pubKey)
		if err != nil {
			return nil, err
		}
		if n != int(length) {
			return nil, io.EOF
		}

		k, err := corekeys.NewChainType(typ)
		if err != nil {
			// skipping unknown key type
			continue
		}
		m[string(k)] = pubKey

		if buf.Len() == 0 {
			break
		}
	}

	return m, nil
}
