package core_test

import (
	"context"
	"crypto"
	"crypto/ed25519"
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

func TestSingleAccountSigner_NewSingleAccountSigner(t *testing.T) {
	singleSigner, err := core.NewSingleAccountSigner("account1", &mockSigner{})
	require.NoError(t, err)
	assert.NotNil(t, singleSigner)
}

func TestSingleAccountSigner_Accounts(t *testing.T) {
	t.Run("returns all accounts", func(t *testing.T) {
		expectedAccounts := []string{"account1"}
		signer := &mockSigner{}

		singleSigner, err := core.NewSingleAccountSigner("account1", signer)
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

		singleSigner, err := core.NewSingleAccountSigner("account1", signer)
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

		singleSigner, err := core.NewSingleAccountSigner("account2", signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "account2", data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("successfully signs with second account", func(t *testing.T) {
		expectedSignature := []byte("second_signature")
		signer := &mockSigner{signData: expectedSignature}

		singleSigner, err := core.NewSingleAccountSigner("account2", signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "account2", data)

		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("returns error for non-existent account", func(t *testing.T) {
		signer := &mockSigner{}
		singleSigner, err := core.NewSingleAccountSigner("account1", signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "non_existent_account", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "account not found: non_existent_account")
	})

	t.Run("propagates signer error", func(t *testing.T) {
		expectedError := fmt.Errorf("signing failed")
		signer := &mockSigner{signError: expectedError}
		singleSigner, err := core.NewSingleAccountSigner("account1", signer)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "account1", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "signing failed")
	})
}

func TestSingleAccountSigner_Integration(t *testing.T) {
	t.Run("real ed25519 keys integration", func(t *testing.T) {
		privKey := ed25519.NewKeyFromSeed([]byte("test_seed_that_is_32_bytes_long!"))
		singleSigner, err := core.NewSingleAccountSigner("key1", privKey)
		require.NoError(t, err)

		ctx := context.Background()
		testData := []byte("integration test data")

		signature1, err := singleSigner.Sign(ctx, "key1", testData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		valid := ed25519.Verify(privKey.Public().(ed25519.PublicKey), testData, signature1)
		assert.True(t, valid, "signature should be valid")
	})
}
