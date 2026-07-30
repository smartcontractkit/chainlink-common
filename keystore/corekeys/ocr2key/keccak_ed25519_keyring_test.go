package ocr2key

import (
	"bytes"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestKeccakEd25519Keyring_SignVerify(t *testing.T) {
	kr1, err := newKeccakEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	kr2, err := newKeccakEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	ctx := ocrtypes.ReportContext{}

	t.Run("can verify", func(t *testing.T) {
		report := ocrtypes.Report{0x01, 0x02, 0x03}
		sig, err := kr1.Sign(ctx, report)
		require.NoError(t, err)
		// pubkey(32) || ed25519 signature(64)
		require.Len(t, sig, ed25519.PublicKeySize+ed25519.SignatureSize)
		assert.True(t, kr2.Verify(kr1.PublicKey(), ctx, report, sig))
		// Independently pin the signed message to the keccak256 ReportToSigData digest
		// (what the Soroban forwarder recomputes) — a blake2b signature would fail here.
		assert.True(t, ed25519.Verify(ed25519.PublicKey(kr1.PublicKey()), ReportToSigData(ctx, report), sig[ed25519.PublicKeySize:]))
	})

	t.Run("invalid sig", func(t *testing.T) {
		report := ocrtypes.Report{0x01}
		assert.False(t, kr2.Verify(kr1.PublicKey(), ctx, report, []byte{0x01}))
	})

	t.Run("invalid pubkey", func(t *testing.T) {
		report := ocrtypes.Report{0x01}
		sig, err := kr1.Sign(ctx, report)
		require.NoError(t, err)
		assert.False(t, kr2.Verify([]byte{0x01}, ctx, report, sig))
	})

	t.Run("tampered report fails", func(t *testing.T) {
		report := ocrtypes.Report{0x01, 0x02}
		sig, err := kr1.Sign(ctx, report)
		require.NoError(t, err)
		assert.False(t, kr1.Verify(kr1.PublicKey(), ctx, ocrtypes.Report{0x01, 0x03}, sig))
	})
}

func TestKeccakEd25519Keyring_Sign3Verify3(t *testing.T) {
	kr1, err := newKeccakEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)
	kr2, err := newKeccakEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)

	digest, err := types.BytesToConfigDigest(mustRandBytes(32))
	require.NoError(t, err)
	seqNr := rand.Uint64()
	r := ocrtypes.Report(mustRandBytes(rand.Intn(1024)))

	t.Run("can verify", func(t *testing.T) {
		sig, err := kr1.Sign3(digest, seqNr, r)
		require.NoError(t, err)
		assert.True(t, kr2.Verify3(kr1.PublicKey(), digest, seqNr, r, sig))
		// Independently pin the signed message to the keccak256 ReportToSigData3 digest.
		assert.True(t, ed25519.Verify(ed25519.PublicKey(kr1.PublicKey()), ReportToSigData3(digest, seqNr, r), sig[ed25519.PublicKeySize:]))
	})

	t.Run("invalid sig", func(t *testing.T) {
		assert.False(t, kr2.Verify3(kr1.PublicKey(), digest, seqNr, r, []byte{0x01}))
	})

	t.Run("invalid pubkey", func(t *testing.T) {
		sig, err := kr1.Sign3(digest, seqNr, r)
		require.NoError(t, err)
		assert.False(t, kr2.Verify3([]byte{0x01}, digest, seqNr, r, sig))
	})
}

func TestKeccakEd25519Keyring_Marshalling(t *testing.T) {
	kr1, err := newKeccakEd25519Keyring(cryptorand.Reader)
	require.NoError(t, err)

	m, err := kr1.Marshal()
	require.NoError(t, err)

	kr2 := &keccakEd25519Keyring{ed25519Keyring: &ed25519Keyring{}}
	require.NoError(t, kr2.Unmarshal(m))

	assert.True(t, bytes.Equal(kr1.pubKey, kr2.pubKey))
	assert.True(t, bytes.Equal(kr1.privKey(), kr2.privKey()))

	// signature from the reloaded keyring still verifies (keccak digest preserved).
	ctx := ocrtypes.ReportContext{}
	report := ocrtypes.Report{0x09}
	sig, err := kr2.Sign(ctx, report)
	require.NoError(t, err)
	assert.True(t, kr1.Verify(kr2.PublicKey(), ctx, report, sig))

	// Invalid seed size should error.
	assert.Error(t, kr2.Unmarshal([]byte{0x01}))
}
