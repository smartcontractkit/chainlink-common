package kms_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/keystore"
	kms "github.com/smartcontractkit/chainlink-common/keystore/kms"
	"github.com/stretchr/testify/require"
)

func TestKMSKeystore(t *testing.T) {
	keyID, keyID2 := "test-key-id", "test-key-id-2"
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	key2, err := crypto.GenerateKey()
	require.NoError(t, err)
	fakeClient, err := kms.NewFakeKMSClient([]kms.Key{
		{
			PrivateKey: key,
			KeyID:      keyID,
		},
		{
			PrivateKey: key2,
			KeyID:      keyID2,
		},
	})
	require.NoError(t, err)
	ks, err := kms.NewKeystore(fakeClient)
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("GetKeys", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 2)
			require.Equal(t, keyID, resp.Keys[0].KeyInfo.Name)
			require.Equal(t, keyID2, resp.Keys[1].KeyInfo.Name)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{"no-such-key"},
			})
			require.Error(t, err)
		})
		t.Run("specific keys", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{keyID},
			})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 1)
			require.Equal(t, keyID, resp.Keys[0].KeyInfo.Name)
		})
	})

	t.Run("SignVerify", func(t *testing.T) {
		t.Run("invalid sign request", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyID,
				Data:    make([]byte, 31), // 31 byte digest
			})
			require.Error(t, err)
			require.ErrorIs(t, err, keystore.ErrInvalidSignRequest)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: "no-such-key",
				Data:    make([]byte, 32), // 32 byte digest
			})
			require.Error(t, err)
		})
		t.Run("success", func(t *testing.T) {
			signResp, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyID,
				Data:    make([]byte, 32), // 32 byte digest
			})
			require.NoError(t, err)
			require.NotNil(t, signResp.Signature)
			verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
				KeyType:   keystore.ECDSA_S256,
				PublicKey: crypto.FromECDSAPub(&key.PublicKey),
				Data:      make([]byte, 32), // 32 byte digest
				Signature: signResp.Signature,
			})
			require.NoError(t, err)
			require.True(t, verifyResp.Valid)
		})
	})
}
