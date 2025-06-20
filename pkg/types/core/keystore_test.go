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

func TestMultiAccountSigner_NewMultiAccountSigner(t *testing.T) {
	accounts := []string{"account1", "account2"}
	signers := []crypto.Signer{&mockSigner{}, &mockSigner{}}

	multiSigner := core.NewMultiAccountSigner(accounts, signers)

	assert.NotNil(t, multiSigner)
}

func TestMultiAccountSigner_Accounts(t *testing.T) {
	t.Run("returns all accounts", func(t *testing.T) {
		expectedAccounts := []string{"account1", "account2", "account3"}
		signers := []crypto.Signer{&mockSigner{}, &mockSigner{}, &mockSigner{}}

		multiSigner := core.NewMultiAccountSigner(expectedAccounts, signers)

		ctx := context.Background()
		accounts, err := multiSigner.Accounts(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedAccounts, accounts)
	})

	t.Run("returns empty slice when no accounts", func(t *testing.T) {
		multiSigner := core.NewMultiAccountSigner([]string{}, []crypto.Signer{})

		ctx := context.Background()
		accounts, err := multiSigner.Accounts(ctx)

		require.NoError(t, err)
		assert.Empty(t, accounts)
	})
}

func TestMultiAccountSigner_Sign(t *testing.T) {
	t.Run("successfully signs with valid account", func(t *testing.T) {
		expectedSignature := []byte("signature_data")
		accounts := []string{"account1", "account2"}
		signers := []crypto.Signer{
			&mockSigner{signData: expectedSignature},
			&mockSigner{signData: []byte("other_signature")},
		}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := multiSigner.Sign(ctx, "account1", data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("successfully signs with second account", func(t *testing.T) {
		expectedSignature := []byte("second_signature")
		accounts := []string{"account1", "account2"}
		signers := []crypto.Signer{
			&mockSigner{signData: []byte("first_signature")},
			&mockSigner{signData: expectedSignature},
		}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := multiSigner.Sign(ctx, "account2", data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("returns error for non-existent account", func(t *testing.T) {
		accounts := []string{"account1", "account2"}
		signers := []crypto.Signer{&mockSigner{}, &mockSigner{}}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := multiSigner.Sign(ctx, "non_existent_account", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "account not found: non_existent_account")
	})

	t.Run("propagates signer error", func(t *testing.T) {
		expectedError := fmt.Errorf("signing failed")
		accounts := []string{"account1"}
		signers := []crypto.Signer{
			&mockSigner{signError: expectedError},
		}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := multiSigner.Sign(ctx, "account1", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Equal(t, expectedError, err)
	})

	t.Run("handles empty data", func(t *testing.T) {
		expectedSignature := []byte("empty_data_signature")
		accounts := []string{"account1"}
		signers := []crypto.Signer{
			&mockSigner{signData: expectedSignature},
		}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		signature, err := multiSigner.Sign(ctx, "account1", []byte{})

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("handles nil data", func(t *testing.T) {
		expectedSignature := []byte("nil_data_signature")
		accounts := []string{"account1"}
		signers := []crypto.Signer{
			&mockSigner{signData: expectedSignature},
		}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		signature, err := multiSigner.Sign(ctx, "account1", nil)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})
}

func TestMultiAccountSigner_Integration(t *testing.T) {
	t.Run("real ed25519 keys integration", func(t *testing.T) {
		// Generate real ed25519 keys for integration testing
		pubKey1, privKey1, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)
		pubKey2, privKey2, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		// Create real signers
		signer1 := &ed25519Signer{privateKey: privKey1, publicKey: pubKey1}
		signer2 := &ed25519Signer{privateKey: privKey2, publicKey: pubKey2}

		accounts := []string{"key1", "key2"}
		signers := []crypto.Signer{signer1, signer2}

		multiSigner := core.NewMultiAccountSigner(accounts, signers)

		ctx := context.Background()
		testData := []byte("integration test data")

		// Test signing with first key
		signature1, err := multiSigner.Sign(ctx, "key1", testData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		// Verify signature
		valid := ed25519.Verify(pubKey1, testData, signature1)
		assert.True(t, valid, "signature should be valid")

		// Test signing with second key
		signature2, err := multiSigner.Sign(ctx, "key2", testData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature2)

		// Verify second signature
		valid = ed25519.Verify(pubKey2, testData, signature2)
		assert.True(t, valid, "signature should be valid")

		// Signatures should be different
		assert.NotEqual(t, signature1, signature2)
	})
}

// ed25519Signer is a real implementation for integration testing
type ed25519Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func (e *ed25519Signer) Public() crypto.PublicKey {
	return e.publicKey
}

func (e *ed25519Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return ed25519.Sign(e.privateKey, digest), nil
}
