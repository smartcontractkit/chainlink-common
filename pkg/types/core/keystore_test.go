package core_test

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// mockSigner implements crypto.Signer for testing
type mockSigner struct {
	publicKey crypto.PublicKey
	signData  []byte
	signError error
}

func (m *mockSigner) Public() crypto.PublicKey {
	return m.publicKey
}

func (m *mockSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	if m.signError != nil {
		return nil, m.signError
	}
	return m.signData, nil
}

func TestSingleAccountSigner_NewSingleAccountSigner(t *testing.T) {
	account := "account1"
	singleSigner, err := core.NewSingleAccountSigner(&account, &mockSigner{})
	require.NoError(t, err)
	assert.NotNil(t, singleSigner)
}

func TestSingleAccountSigner_Accounts(t *testing.T) {
	t.Run("returns all accounts", func(t *testing.T) {
		expectedAccounts := []string{"account1"}
		signer := &mockSigner{}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		accounts, err := singleSigner.Accounts(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedAccounts, accounts)
	})
}

func TestSingleAccountSigner_Sign(t *testing.T) {
	t.Run("successfully signs with valid account", func(t *testing.T) {
		expectedSignature := []byte("signature_data")
		signer := &mockSigner{signData: expectedSignature}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "account1", data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("successfully signs with second account", func(t *testing.T) {
		expectedSignature := []byte("second_signature")
		signer := &mockSigner{signData: expectedSignature}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, account, data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("successfully signs with second account", func(t *testing.T) {
		expectedSignature := []byte("second_signature")
		signer := &mockSigner{signData: expectedSignature}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, account, data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("returns error for non-existent account", func(t *testing.T) {
		signer := &mockSigner{}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "non_existent_account", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "account not found: non_existent_account")
	})

	t.Run("returns error for no account", func(t *testing.T) {
		singleSigner, err := core.NewSingleAccountSigner(nil, nil)
		require.NoError(t, err)

		accounts, err := singleSigner.Accounts(context.Background())
		assert.ErrorContains(t, err, "account is nil")
		assert.Empty(t, accounts)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "account1", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "account not found: account1")
	})

	t.Run("propagates signer error", func(t *testing.T) {
		expectedError := fmt.Errorf("signing failed")
		signer := &mockSigner{signError: expectedError}
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, account, data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "signing failed")
	})
}

func TestSingleAccountSigner_Integration(t *testing.T) {
	t.Run("real ed25519 keys integration", func(t *testing.T) {
		privKey := ed25519.NewKeyFromSeed([]byte("test_seed_that_is_32_bytes_long!"))
		account := "account1"
		singleSigner, err := core.NewSingleAccountSigner(&account, privKey)
		require.NoError(t, err)

		ctx := context.Background()
		testData := []byte("integration test data")

		signature1, err := singleSigner.Sign(ctx, account, testData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		valid := ed25519.Verify(privKey.Public().(ed25519.PublicKey), testData, signature1)
		assert.True(t, valid, "signature should be valid")

		_, err = singleSigner.Decrypt(ctx, account, []byte("encrypted_data"))
		require.ErrorContains(t, err, "decrypt not supported for single account signer")
	})
}

type boxDecrypter struct {
	privateKey *[32]byte
	publicKey  *[32]byte
}

var _ core.Decrypter = (*boxDecrypter)(nil)

func (b *boxDecrypter) Public() crypto.PublicKey {
	pubKeyBytes := b.publicKey[:]
	return crypto.PublicKey(pubKeyBytes)
}

func (b *boxDecrypter) Decrypt(ciphertext []byte) ([]byte, error) {
	msg, ok := box.OpenAnonymous(nil, ciphertext, b.publicKey, b.privateKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}
	return msg, nil
}

func TestSingleAccountSignerDecrypter_Integration(t *testing.T) {
	t.Run("real ed25519 keys integration", func(t *testing.T) {
		privKey := ed25519.NewKeyFromSeed([]byte("test_seed_that_is_32_bytes_long!"))
		account := "sign_account1"
		singleSigner, err := core.NewSignerDecrypter(&account, privKey, nil)
		require.NoError(t, err)

		ctx := context.Background()
		testData := []byte("integration test data")

		signature1, err := singleSigner.Sign(ctx, account, testData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		valid := ed25519.Verify(privKey.Public().(ed25519.PublicKey), testData, signature1)
		assert.True(t, valid, "signature should be valid")
	})

	t.Run("real nacl/box keys decrypt integration", func(t *testing.T) {
		pubKey, privKey, err := box.GenerateKey(rand.Reader)
		require.NoError(t, err)
		account := "decrypt_account1"

		signerDecrypter, err := core.NewSignerDecrypter(&account, nil, &boxDecrypter{
			privateKey: privKey,
			publicKey:  pubKey,
		})
		require.NoError(t, err)

		msg := []byte("message")
		encrypted, err := box.SealAnonymous(nil, msg, pubKey, rand.Reader)
		require.NoError(t, err)

		ctx := context.Background()
		decrypted, err := signerDecrypter.Decrypt(ctx, account, encrypted)
		require.NoError(t, err)
		assert.Equal(t, msg, decrypted)
	})
}
