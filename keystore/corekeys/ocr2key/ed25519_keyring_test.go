package ocr2key

import (
	"bytes"
	cryptorand "crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestAptosKeyRing_Sign_Verify(t *testing.T) {
	kr1, err := newEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	kr2, err := newEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	ctx := ocrtypes.ReportContext{}

	t.Run("can verify", func(t *testing.T) {
		report := ocrtypes.Report{}
		sig, err := kr1.Sign(ctx, report)
		require.NoError(t, err)
		t.Log(len(sig))
		result := kr2.Verify(kr1.PublicKey(), ctx, report, sig)
		require.True(t, result)
	})

	t.Run("invalid sig", func(t *testing.T) {
		report := ocrtypes.Report{}
		result := kr2.Verify(kr1.PublicKey(), ctx, report, []byte{0x01})
		require.False(t, result)
	})

	t.Run("invalid pubkey", func(t *testing.T) {
		report := ocrtypes.Report{}
		sig, err := kr1.Sign(ctx, report)
		require.NoError(t, err)
		result := kr2.Verify([]byte{0x01}, ctx, report, sig)
		require.False(t, result)
	})
}

func TestAptosKeyRing_Marshalling(t *testing.T) {
	kr1, err := newEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	m, err := kr1.Marshal()
	require.NoError(t, err)
	kr2 := ed25519Keyring{}
	err = kr2.Unmarshal(m)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(kr1.pubKey, kr2.pubKey))
	assert.True(t, bytes.Equal(kr1.privKey(), kr2.privKey()))

	// Invalid seed size should error
	require.Error(t, kr2.Unmarshal([]byte{0x01}))
}

// TestStellarKeyBundle_Sign_Verify exercises the chain-type wiring (not just the raw keyring):
// New(corekeys.Stellar) must produce an ed25519-backed OCR2 key bundle that reports
// ChainType == Stellar and round-trips sign/verify. Note Stellar signs the keccak256 CRE
// digest (keccakEd25519Keyring), not Aptos's blake2b digest; the keccak-specific behaviour
// is covered in keccak_ed25519_keyring_test.go. This guards the New() → Stellar mapping.
func TestStellarKeyBundle_Sign_Verify(t *testing.T) {
	kb1, err := New(corekeys.Stellar)
	require.NoError(t, err)
	require.Equal(t, corekeys.Stellar, kb1.ChainType())

	kb2, err := New(corekeys.Stellar)
	require.NoError(t, err)

	ctx := ocrtypes.ReportContext{}
	report := ocrtypes.Report{}

	sig, err := kb1.Sign(ctx, report)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	t.Run("cross-bundle verify succeeds", func(t *testing.T) {
		require.True(t, kb2.Verify(kb1.PublicKey(), ctx, report, sig))
	})
	t.Run("invalid sig fails", func(t *testing.T) {
		require.False(t, kb2.Verify(kb1.PublicKey(), ctx, report, []byte{0x01}))
	})
	t.Run("invalid pubkey fails", func(t *testing.T) {
		require.False(t, kb2.Verify([]byte{0x01}, ctx, report, sig))
	})
}
