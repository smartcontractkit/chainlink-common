package vaultcrypto

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"github.com/smartcontractkit/tdh2/go/tdh2/tdh2easy"

	vaultcommon "github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/vault"
)

var label = [32]byte{31: 1}

type recipient struct {
	pub  *[32]byte
	priv *[32]byte
}

func newRecipient(t *testing.T) recipient {
	pub, priv, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return recipient{pub: pub, priv: priv}
}

func (r recipient) hex() string { return hex.EncodeToString(r.pub[:]) }

func (r recipient) decrypter() Decrypter {
	return DecryptFunc(func(b []byte) ([]byte, error) {
		out, ok := box.OpenAnonymous(nil, b, r.pub, r.priv)
		if !ok {
			return nil, assert.AnError
		}
		return out, nil
	})
}

func setup(t *testing.T, k, n int, secret string) (*tdh2easy.PublicKey, []*tdh2easy.PrivateShare, []byte) {
	_, pk, shares, err := tdh2easy.GenerateKeys(k, n)
	require.NoError(t, err)
	ct, err := tdh2easy.EncryptWithLabel(pk, []byte(secret), label)
	require.NoError(t, err)
	ctb, err := ct.Marshal()
	require.NoError(t, err)
	return pk, shares, ctb
}

func TestGenerateShareEncryptAndDecrypt(t *testing.T) {
	const k, n = 2, 4
	pk, keyShares, ctb := setup(t, k, n, "my secret")
	r := newRecipient(t)

	var encrypted [][]byte
	for _, ks := range keyShares[:k] {
		share, err := GeneratePlaintextShare(pk, ks, ctb, label)
		require.NoError(t, err)
		enc, err := EncryptShareBinary(share, r.hex())
		require.NoError(t, err)
		encrypted = append(encrypted, enc)
	}

	plaintext, rejected, err := DecryptSecret(ctb, pk, k, encrypted, r.decrypter())
	require.NoError(t, err)
	assert.Empty(t, rejected)
	assert.Equal(t, "my secret", string(plaintext))
}

func TestGeneratePlaintextShare_Errors(t *testing.T) {
	pk, keyShares, ctb := setup(t, 2, 4, "my secret")

	_, err := GeneratePlaintextShare(pk, keyShares[0], ctb, [32]byte{31: 2})
	require.ErrorContains(t, err, "failed to verify label on secret")

	_, err = GeneratePlaintextShare(pk, keyShares[0], []byte("garbage"), label)
	require.ErrorContains(t, err, "failed to unmarshal ciphertext")

	_, err = GeneratePlaintextShare(nil, keyShares[0], ctb, label)
	require.ErrorContains(t, err, "public key is not set")

	_, err = GeneratePlaintextShare(pk, nil, ctb, label)
	require.ErrorContains(t, err, "private key share is not set")
}

func TestEncryptShareBinary_InvalidKey(t *testing.T) {
	_, err := EncryptShareBinary([]byte("share"), "zz")
	require.ErrorIs(t, err, ErrInvalidRecipientKey)

	_, err = EncryptShareBinary([]byte("share"), "abcd")
	require.ErrorIs(t, err, ErrInvalidRecipientKey)
}

func TestDecryptSecret_SkipsBadShares(t *testing.T) {
	const k, n = 2, 4
	pk, keyShares, ctb := setup(t, k, n, "my secret")
	r := newRecipient(t)
	other := newRecipient(t)

	encryptFor := func(ks *tdh2easy.PrivateShare, to recipient) []byte {
		share, err := GeneratePlaintextShare(pk, ks, ctb, label)
		require.NoError(t, err)
		enc, err := EncryptShareBinary(share, to.hex())
		require.NoError(t, err)
		return enc
	}

	// A share addressed to someone else and a garbage share are rejected.
	garbage, err := EncryptShareBinary([]byte("not a share"), r.hex())
	require.NoError(t, err)
	bad := [][]byte{encryptFor(keyShares[0], other), garbage}

	_, rejected, err := DecryptSecret(ctb, pk, k, append(bad, encryptFor(keyShares[0], r)), r.decrypter())
	require.ErrorContains(t, err, "not enough decryption shares")
	assert.Len(t, rejected, 2)

	plaintext, rejected, err := DecryptSecret(ctb, pk, k, append(bad, encryptFor(keyShares[0], r), encryptFor(keyShares[1], r)), r.decrypter())
	require.NoError(t, err)
	assert.Len(t, rejected, 2)
	assert.Equal(t, "my secret", string(plaintext))
}

func TestSharesForKey(t *testing.T) {
	entries := []*vaultcommon.EncryptedShares{
		{EncryptionKey: "a", BinaryShares: [][]byte{{1}, {2}}},
		{EncryptionKey: "b", BinaryShares: [][]byte{{3}}},
		{EncryptionKey: "a", Shares: []string{"04"}},
		{EncryptionKey: "a", Shares: []string{"05"}, BinaryShares: [][]byte{{6}}},
	}

	got, err := SharesForKey(entries, "a")
	require.NoError(t, err)
	assert.Equal(t, [][]byte{{1}, {2}, {4}, {6}}, got)

	got, err = SharesForKey(entries, "c")
	require.NoError(t, err)
	assert.Empty(t, got)

	_, err = SharesForKey([]*vaultcommon.EncryptedShares{{EncryptionKey: "a", Shares: []string{"zz"}}}, "a")
	require.Error(t, err)
}
