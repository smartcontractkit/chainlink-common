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

// mockDecrypter implements crypto.Decrypter for testing
type mockDecrypter struct {
	publicKey    crypto.PublicKey
	decryptData  []byte
	decryptError error
}

func (m *mockDecrypter) Public() crypto.PublicKey {
	return m.publicKey
}

func (m *mockDecrypter) Decrypt(rand io.Reader, msg []byte, opts crypto.DecrypterOpts) (plaintext []byte, err error) {
	if m.decryptError != nil {
		return nil, m.decryptError
	}
	return m.decryptData, nil
}

func TestSingleAccountSigner_NewSingleAccountSigner(t *testing.T) {
	account := "account1"
	singleSigner, err := core.NewSignerDecrypter(&account, &mockSigner{}, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, singleSigner)
}

func TestSingleAccountSigner_Accounts(t *testing.T) {
	t.Run("returns all accounts", func(t *testing.T) {
		expectedAccounts := []string{"account1"}
		signer := &mockSigner{}
		account := "account1"
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
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
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
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
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
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
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
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
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, "non_existent_account", data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "account not found: non_existent_account")
	})

	t.Run("returns error for no account", func(t *testing.T) {
		singleSigner, err := core.NewSignerDecrypter(nil, nil, nil, nil)
		require.NoError(t, err)

		accounts, err := singleSigner.Accounts(context.Background())
		assert.ErrorContains(t, err, "no accounts found")
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
		singleSigner, err := core.NewSignerDecrypter(&account, signer, nil, nil)
		require.NoError(t, err)

		ctx := context.Background()
		data := []byte("test_data")
		signature, err := singleSigner.Sign(ctx, account, data)

		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "signing failed")
	})
}

func TestSignerDecrypter_Decrypt(t *testing.T) {
	t.Run("successfully decrypts with valid account", func(t *testing.T) {
		expectedDecrypted := []byte("decrypted_data")
		decrypter := &mockDecrypter{decryptData: expectedDecrypted}
		decryptAccount := "decrypt_account"
		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, &decryptAccount, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		encrypted := []byte("encrypted_data")
		decrypted, err := signerDecrypter.Decrypt(ctx, "decrypt_account", encrypted)

		require.NoError(t, err)
		assert.Equal(t, expectedDecrypted, decrypted)
	})

	t.Run("returns error for non-existent decrypt account", func(t *testing.T) {
		decrypter := &mockDecrypter{}
		decryptAccount := "decrypt_account"
		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, &decryptAccount, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		encrypted := []byte("encrypted_data")
		decrypted, err := signerDecrypter.Decrypt(ctx, "non_existent_account", encrypted)

		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.Contains(t, err.Error(), "account not found: non_existent_account")
	})

	t.Run("returns error for no decrypt account", func(t *testing.T) {
		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, nil, nil)
		require.NoError(t, err)

		ctx := context.Background()
		encrypted := []byte("encrypted_data")
		decrypted, err := signerDecrypter.Decrypt(ctx, "decrypt_account", encrypted)

		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.Contains(t, err.Error(), "account not found: decrypt_account")
	})

	t.Run("propagates decrypter error", func(t *testing.T) {
		expectedError := fmt.Errorf("decryption failed")
		decrypter := &mockDecrypter{decryptError: expectedError}
		decryptAccount := "decrypt_account"
		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, &decryptAccount, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		encrypted := []byte("encrypted_data")
		decrypted, err := signerDecrypter.Decrypt(ctx, decryptAccount, encrypted)

		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.Contains(t, err.Error(), "decryption failed")
	})
}

func TestSignerDecrypter_Accounts_WithBothSignerAndDecrypter(t *testing.T) {
	t.Run("returns both sign and decrypt accounts", func(t *testing.T) {
		signAccount := "sign_account"
		decryptAccount := "decrypt_account"
		signer := &mockSigner{}
		decrypter := &mockDecrypter{}
		signerDecrypter, err := core.NewSignerDecrypter(&signAccount, signer, &decryptAccount, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		accounts, err := signerDecrypter.Accounts(ctx)

		require.NoError(t, err)
		assert.Len(t, accounts, 2)
		assert.Contains(t, accounts, signAccount)
		assert.Contains(t, accounts, decryptAccount)
	})

	t.Run("returns only sign account when no decrypt account", func(t *testing.T) {
		signAccount := "sign_account"
		signer := &mockSigner{}
		signerDecrypter, err := core.NewSignerDecrypter(&signAccount, signer, nil, nil)
		require.NoError(t, err)

		ctx := context.Background()
		accounts, err := signerDecrypter.Accounts(ctx)

		require.NoError(t, err)
		assert.Len(t, accounts, 1)
		assert.Contains(t, accounts, signAccount)
	})

	t.Run("returns only decrypt account when no sign account", func(t *testing.T) {
		decryptAccount := "decrypt_account"
		decrypter := &mockDecrypter{}
		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, &decryptAccount, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		accounts, err := signerDecrypter.Accounts(ctx)

		require.NoError(t, err)
		assert.Len(t, accounts, 1)
		assert.Contains(t, accounts, decryptAccount)
	})

	t.Run("handles same account for both sign and decrypt", func(t *testing.T) {
		account := "same_account"
		signer := &mockSigner{}
		decrypter := &mockDecrypter{}
		signerDecrypter, err := core.NewSignerDecrypter(&account, signer, &account, decrypter)
		require.NoError(t, err)

		ctx := context.Background()
		accounts, err := signerDecrypter.Accounts(ctx)

		require.NoError(t, err)
		assert.Len(t, accounts, 2) // Should contain both entries even if same account name
		assert.Contains(t, accounts, account)
	})
}

type boxDecrypter struct {
	privateKey *[32]byte
	publicKey  *[32]byte
}

var _ crypto.Decrypter = (*boxDecrypter)(nil)

func (b *boxDecrypter) Public() crypto.PublicKey {
	pubKeyBytes := b.publicKey[:]
	return crypto.PublicKey(pubKeyBytes)
}

func (b *boxDecrypter) Decrypt(_ io.Reader, ciphertext []byte, _ crypto.DecrypterOpts) ([]byte, error) {
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
		singleSigner, err := core.NewSignerDecrypter(&account, privKey, nil, nil)
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

		signerDecrypter, err := core.NewSignerDecrypter(nil, nil, &account, &boxDecrypter{
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
