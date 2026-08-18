package ocr2key

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

// This is a wire format shared with whoever writes an OCR configuration, so it is
// asserted as bytes rather than only round-tripped: a round trip passes just as
// well against an encoding no configuration uses.
func TestMarshalMultichainPublicKeyLayout(t *testing.T) {
	t.Parallel()

	// A 20 byte EVM address, which is what an EVM bundle's PublicKey is.
	evm := ocrtypes.OnchainPublicKey{
		0x1a, 0x0e, 0x37, 0x3a, 0x3b, 0xcb, 0x04, 0x96, 0xb0, 0x23,
		0x0a, 0xef, 0x64, 0x2f, 0x03, 0xcf, 0x52, 0x8b, 0x9f, 0x9c,
	}

	encoded, err := MarshalMultichainPublicKey(map[string]ocrtypes.OnchainPublicKey{string(corekeys.EVM): evm})
	require.NoError(t, err)

	// family 1 (EVM), then the length as a little-endian uint16, then the key.
	assert.Equal(t, append(ocrtypes.OnchainPublicKey{0x01, 0x14, 0x00}, evm...), encoded)
}

// The entries are ordered by family byte, so two processes holding the same keys
// produce the same bytes however their maps iterate.
func TestMarshalMultichainPublicKeyOrdersByFamily(t *testing.T) {
	t.Parallel()

	// Aptos is family 5 and EVM is 1, so EVM comes first.
	encoded, err := MarshalMultichainPublicKey(map[string]ocrtypes.OnchainPublicKey{
		string(corekeys.Aptos): {0xaa, 0xbb},
		string(corekeys.EVM):   {0xcc},
	})
	require.NoError(t, err)
	assert.Equal(t, ocrtypes.OnchainPublicKey{
		0x01, 0x01, 0x00, 0xcc,
		0x05, 0x02, 0x00, 0xaa, 0xbb,
	}, encoded)

	keys, err := UnmarshalMultichainPublicKey(encoded)
	require.NoError(t, err)
	assert.Equal(t, map[string]ocrtypes.OnchainPublicKey{
		string(corekeys.Aptos): {0xaa, 0xbb},
		string(corekeys.EVM):   {0xcc},
	}, keys)
}

// An unknown family is skipped rather than rejected: the rest of the key is still
// the oracle's identity, and a family this build cannot verify signatures for is
// one it has no use for.
func TestMultichainPublicKeySkipsUnknownFamilies(t *testing.T) {
	t.Parallel()

	encoded, err := MarshalMultichainPublicKey(map[string]ocrtypes.OnchainPublicKey{
		string(corekeys.EVM): {0xcc},
		"not-a-chain":        {0xdd},
	})
	require.NoError(t, err)
	assert.Equal(t, ocrtypes.OnchainPublicKey{0x01, 0x01, 0x00, 0xcc}, encoded)

	// Skipped mid-blob, with a known family after it.
	keys, err := UnmarshalMultichainPublicKey(ocrtypes.OnchainPublicKey{
		0xfe, 0x01, 0x00, 0xdd, // family 254 is not a known one
		0x01, 0x01, 0x00, 0xcc,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]ocrtypes.OnchainPublicKey{string(corekeys.EVM): {0xcc}}, keys)
}

// An unknown family in the LAST entry is reported as EOF rather than skipped:
// the skip path continues past the end-of-input check, so the next read fails.
//
// Asserted because it is the behaviour configurations in the wild were written and
// read with, not because it is wanted - the two unknown-family cases should agree,
// and today they do not.
func TestUnmarshalMultichainPublicKeyTrailingUnknownFamilyIsEOF(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalMultichainPublicKey(ocrtypes.OnchainPublicKey{
		0x01, 0x01, 0x00, 0xcc,
		0xfe, 0x01, 0x00, 0xdd, // family 254 is not a known one, and is last
	})
	require.ErrorIs(t, err, io.EOF)
}

func TestUnmarshalMultichainPublicKeyRejectsTruncated(t *testing.T) {
	t.Parallel()

	for name, encoded := range map[string]ocrtypes.OnchainPublicKey{
		"empty":       {},
		"header only": {0x01, 0x14},
		"short value": {0x01, 0x14, 0x00, 0xff},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := UnmarshalMultichainPublicKey(encoded)
			assert.Error(t, err)
		})
	}
}

// The bundle form is the key form applied to the bundles' public keys, so the two
// cannot disagree about what an oracle holding those bundles is called.
func TestMarshalMultichainKeyBundleMatchesPublicKeys(t *testing.T) {
	t.Parallel()

	bundle, err := New(corekeys.EVM)
	require.NoError(t, err)

	fromBundle, err := MarshalMultichainKeyBundle(map[string]KeyBundle{string(corekeys.EVM): bundle})
	require.NoError(t, err)

	fromKey, err := MarshalMultichainPublicKey(map[string]ocrtypes.OnchainPublicKey{
		string(corekeys.EVM): bundle.PublicKey(),
	})
	require.NoError(t, err)

	assert.Equal(t, fromKey, fromBundle)
}
