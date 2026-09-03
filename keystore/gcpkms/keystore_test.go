package gcpkms_test

import (
	"crypto/ed25519"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
	gcpkms "github.com/smartcontractkit/chainlink-common/keystore/gcpkms"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

const (
	keyRingName = "projects/test-project/locations/us-central1/keyRings/test-ring"
	keyName     = keyRingName + "/cryptoKeys/test-key"
	keyName2    = keyRingName + "/cryptoKeys/test-key-2"

	// Key names are CryptoKeyVersion resource names: a name always names exactly one version.
	keyVersion1  = keyName + "/cryptoKeyVersions/1"
	key2Version1 = keyName2 + "/cryptoKeyVersions/1"
)

func TestGCPKMSKeystore(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	key2, err := crypto.GenerateKey()
	require.NoError(t, err)
	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{
			KeyType:    keystore.ECDSA_S256,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(key)),
			KeyID:      keyName,
		},
		{
			KeyType:    keystore.ECDSA_S256,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(key2)),
			KeyID:      keyName2,
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient)
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("GetKeys", func(t *testing.T) {
		t.Run("no key names is rejected", func(t *testing.T) {
			_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
			require.ErrorContains(t, err, "does not list key rings")
		})
		t.Run("specific keys are sorted by name", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{keyVersion1, key2Version1},
			})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 2)
			require.Equal(t, key2Version1, resp.Keys[0].KeyInfo.Name)
			require.Equal(t, keyVersion1, resp.Keys[1].KeyInfo.Name)
			require.Equal(t, keystore.ECDSA_S256, resp.Keys[1].KeyInfo.KeyType)
			require.Equal(t, crypto.FromECDSAPub(&key.PublicKey), resp.Keys[1].KeyInfo.PublicKey)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{"projects/p/locations/l/keyRings/r/cryptoKeys/nope/cryptoKeyVersions/1"},
			})
			require.Error(t, err)
		})
		t.Run("bare crypto key name is rejected", func(t *testing.T) {
			_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{keyName},
			})
			require.ErrorContains(t, err, "not a CryptoKeyVersion resource name")
		})
	})

	t.Run("SignVerify", func(t *testing.T) {
		t.Run("invalid sign request", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyVersion1,
				Data:    make([]byte, 31), // 31 byte digest
			})
			require.Error(t, err)
			require.ErrorIs(t, err, keystore.ErrInvalidSignRequest)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: "projects/p/locations/l/keyRings/r/cryptoKeys/nope/cryptoKeyVersions/1",
				Data:    make([]byte, 32), // 32 byte digest
			})
			require.Error(t, err)
		})
		t.Run("bare crypto key name is rejected", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyName,
				Data:    make([]byte, 32),
			})
			require.ErrorContains(t, err, "not a CryptoKeyVersion resource name")
		})
		t.Run("success", func(t *testing.T) {
			signResp, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyVersion1,
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

func TestGCPKMSKeystore_Ed25519(t *testing.T) {
	_, ed25519PrivKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	ed25519PubKey := ed25519PrivKey.Public().(ed25519.PublicKey)

	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{
			KeyType:    keystore.Ed25519,
			KeyID:      keyName,
			PrivateKey: internal.NewRaw(ed25519PrivKey),
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient)
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("GetKeys", func(t *testing.T) {
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{keyVersion1},
		})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, keyVersion1, resp.Keys[0].KeyInfo.Name)
		require.Equal(t, keystore.Ed25519, resp.Keys[0].KeyInfo.KeyType)
		require.Equal(t, []byte(ed25519PubKey), resp.Keys[0].KeyInfo.PublicKey)
	})

	t.Run("SignVerify", func(t *testing.T) {
		// Ed25519 can sign arbitrary length messages
		testData := []byte("hello, world")
		signResp, err := ks.Sign(ctx, keystore.SignRequest{
			KeyName: keyVersion1,
			Data:    testData,
		})
		require.NoError(t, err)
		require.NotNil(t, signResp.Signature)
		require.Len(t, signResp.Signature, 64) // Ed25519 signatures are 64 bytes

		verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
			KeyType:   keystore.Ed25519,
			PublicKey: ed25519PubKey,
			Data:      testData,
			Signature: signResp.Signature,
		})
		require.NoError(t, err)
		require.True(t, verifyResp.Valid)
	})
}

// Requesting a version this keystore cannot use e.g. an unsupported algorithm or a disabled
// version surfaces the error from both GetKeys and Sign.
func TestGCPKMSKeystore_UnusableVersions(t *testing.T) {
	p256Key, err := crypto.GenerateKey()
	require.NoError(t, err)
	disabledKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{
			// A signing key on a curve this keystore does not support.
			KeyType:    keystore.ECDSA_S256,
			KeyID:      keyName,
			Algorithm:  kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(p256Key)),
		},
		{
			KeyType:    keystore.ECDSA_S256,
			KeyID:      keyName2,
			State:      kmspb.CryptoKeyVersion_DISABLED,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(disabledKey)),
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient)
	require.NoError(t, err)
	ctx := t.Context()

	_, err = ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyVersion1}})
	require.ErrorContains(t, err, "unsupported Cloud KMS key algorithm")
	_, err = ks.Sign(ctx, keystore.SignRequest{KeyName: keyVersion1, Data: make([]byte, 32)})
	require.ErrorContains(t, err, "unsupported Cloud KMS key algorithm")

	_, err = ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{key2Version1}})
	require.ErrorContains(t, err, "is not enabled")
	_, err = ks.Sign(ctx, keystore.SignRequest{KeyName: key2Version1, Data: make([]byte, 32)})
	require.ErrorContains(t, err, "is not enabled")
}

func TestGCPKMSKeystore_InvalidEd25519Key(t *testing.T) {
	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{
			KeyType:    keystore.Ed25519,
			KeyID:      keyName,
			PrivateKey: internal.NewRaw(make([]byte, ed25519.SeedSize)), // seed, not a private key
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient)
	require.NoError(t, err)

	// Must error rather than panic inside crypto/ed25519.
	_, err = ks.Sign(t.Context(), keystore.SignRequest{KeyName: keyVersion1, Data: []byte("hello")})
	require.ErrorContains(t, err, "invalid Ed25519 private key length")
}

// A CryptoKeyVersion name names one version forever, so a rotation can never change which key a
// configured name signs with: the public key a caller already holds stays the one that verifies
// its signatures. Adopting a rotation means configuring the new version's name.
func TestGCPKMSKeystore_Rotation(t *testing.T) {
	originalKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	rotatedKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{KeyType: keystore.ECDSA_S256, KeyID: keyName, VersionNumber: 1, PrivateKey: internal.NewRaw(crypto.FromECDSA(originalKey))},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient)
	require.NoError(t, err)
	ctx := t.Context()

	resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyVersion1}})
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	publicKey := resp.Keys[0].KeyInfo.PublicKey
	require.Equal(t, crypto.FromECDSAPub(&originalKey.PublicKey), publicKey)

	// Cloud KMS gains a newer enabled version after the caller has read the public key.
	require.NoError(t, fakeClient.AddVersion(gcpkms.Key{
		KeyType:       keystore.ECDSA_S256,
		KeyID:         keyName,
		VersionNumber: 2,
		PrivateKey:    internal.NewRaw(crypto.FromECDSA(rotatedKey)),
	}))

	signResp, err := ks.Sign(ctx, keystore.SignRequest{KeyName: keyVersion1, Data: make([]byte, 32)})
	require.NoError(t, err)

	verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
		KeyType:   keystore.ECDSA_S256,
		PublicKey: publicKey,
		Data:      make([]byte, 32),
		Signature: signResp.Signature,
	})
	require.NoError(t, err)
	require.True(t, verifyResp.Valid, "signature must verify against the public key GetKeys reported")

	// GetKeys keeps reporting the configured version too, so the two never diverge.
	resp, err = ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyVersion1}})
	require.NoError(t, err)
	require.Equal(t, publicKey, resp.Keys[0].KeyInfo.PublicKey)
}
