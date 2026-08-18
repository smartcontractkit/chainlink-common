package ocr2key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

// A keyring's own signature must pass its own Verify, and the public key it
// presents must be the one Verify is handed - which is the multichain form, since
// that is what a configuration carries.
func TestOCR3KeyringSignsWhatItVerifies(t *testing.T) {
	t.Parallel()

	bundle, err := New(corekeys.EVM)
	require.NoError(t, err)
	keyring, err := NewOCR3Keyring(string(corekeys.EVM), bundle)
	require.NoError(t, err)

	digest := ocrtypes.ConfigDigest{0x01, 0x02, 0x03}
	report := ocr3types.ReportWithInfo[[]byte]{Report: []byte("a report")}
	const seqNr = uint64(42)

	signature, err := keyring.Sign(digest, seqNr, report)
	require.NoError(t, err)

	assert.True(t, keyring.Verify(keyring.PublicKey(), digest, seqNr, report, signature))

	// The bare key is not what a configuration carries, so it is not what Verify takes.
	assert.False(t, keyring.Verify(bundle.PublicKey(), digest, seqNr, report, signature))

	// A different round is a different signature.
	assert.False(t, keyring.Verify(keyring.PublicKey(), digest, seqNr+1, report, signature))
}

// The presented key is the encoded form of the bundle's own, so a configuration
// written from the same bundle lists this oracle by exactly this.
func TestOCR3KeyringPublicKeyIsEncoded(t *testing.T) {
	t.Parallel()

	bundle, err := New(corekeys.EVM)
	require.NoError(t, err)
	keyring, err := NewOCR3Keyring(string(corekeys.EVM), bundle)
	require.NoError(t, err)

	want, err := MarshalMultichainKeyBundle(map[string]KeyBundle{string(corekeys.EVM): bundle})
	require.NoError(t, err)
	assert.Equal(t, want, keyring.PublicKey())
}

// What a process signing on an oracle's behalf serves has to be the signature the
// oracle would have made itself.
func TestSignOCR3ReportMatchesKeyring(t *testing.T) {
	t.Parallel()

	bundle, err := New(corekeys.EVM)
	require.NoError(t, err)
	keyring, err := NewOCR3Keyring(string(corekeys.EVM), bundle)
	require.NoError(t, err)

	digest := ocrtypes.ConfigDigest{0x0a}
	report := ocrtypes.Report("a report")
	const seqNr = uint64(7)

	served, err := SignOCR3Report(bundle, digest, seqNr, report)
	require.NoError(t, err)

	own, err := keyring.Sign(digest, seqNr, ocr3types.ReportWithInfo[[]byte]{Report: report})
	require.NoError(t, err)

	assert.Equal(t, own, served)
}

func TestNewOCR3KeyringRejectsUnknownFamily(t *testing.T) {
	t.Parallel()

	bundle, err := New(corekeys.EVM)
	require.NoError(t, err)

	_, err = NewOCR3Keyring("not-a-chain", bundle)
	require.Error(t, err)

	_, err = NewOCR3Keyring(string(corekeys.EVM), nil)
	require.Error(t, err)
}
