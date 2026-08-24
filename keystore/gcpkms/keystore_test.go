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
	ks, err := gcpkms.NewKeystore(fakeClient, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("GetKeys", func(t *testing.T) {
		t.Run("listing all keys", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 2)
			require.Equal(t, keyName, resp.Keys[0].KeyInfo.Name)
			require.Equal(t, keyName2, resp.Keys[1].KeyInfo.Name)
		})
		t.Run("specific keys preserve the requested order", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{keyName2, keyName},
			})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 2)
			require.Equal(t, keyName2, resp.Keys[0].KeyInfo.Name)
			require.Equal(t, keyName, resp.Keys[1].KeyInfo.Name)
			require.Equal(t, keystore.ECDSA_S256, resp.Keys[1].KeyInfo.KeyType)
			require.Equal(t, crypto.FromECDSAPub(&key.PublicKey), resp.Keys[1].KeyInfo.PublicKey)
		})
		t.Run("explicit key version", func(t *testing.T) {
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{keyName + "/cryptoKeyVersions/1"},
			})
			require.NoError(t, err)
			require.Len(t, resp.Keys, 1)
			require.Equal(t, keyName+"/cryptoKeyVersions/1", resp.Keys[0].KeyInfo.Name)
			require.Equal(t, crypto.FromECDSAPub(&key.PublicKey), resp.Keys[0].KeyInfo.PublicKey)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
				KeyNames: []string{"projects/p/locations/l/keyRings/r/cryptoKeys/nope"},
			})
			require.Error(t, err)
		})
	})

	t.Run("SignVerify", func(t *testing.T) {
		t.Run("invalid sign request", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyName,
				Data:    make([]byte, 31), // 31 byte digest
			})
			require.Error(t, err)
			require.ErrorIs(t, err, keystore.ErrInvalidSignRequest)
		})
		t.Run("no such key", func(t *testing.T) {
			_, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: "projects/p/locations/l/keyRings/r/cryptoKeys/nope",
				Data:    make([]byte, 32), // 32 byte digest
			})
			require.Error(t, err)
		})
		t.Run("success", func(t *testing.T) {
			signResp, err := ks.Sign(ctx, keystore.SignRequest{
				KeyName: keyName,
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
	ks, err := gcpkms.NewKeystore(fakeClient, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("GetKeys", func(t *testing.T) {
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{keyName},
		})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, keyName, resp.Keys[0].KeyInfo.Name)
		require.Equal(t, keystore.Ed25519, resp.Keys[0].KeyInfo.KeyType)
		require.Equal(t, []byte(ed25519PubKey), resp.Keys[0].KeyInfo.PublicKey)
	})

	t.Run("SignVerify", func(t *testing.T) {
		// Ed25519 can sign arbitrary length messages
		testData := []byte("hello, world")
		signResp, err := ks.Sign(ctx, keystore.SignRequest{
			KeyName: keyName,
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

// The keystore must never rely on CryptoKey.Primary: Cloud KMS only populates it for
// ENCRYPT_DECRYPT keys, so an asymmetric signing key's version has to be resolved by listing.
func TestGCPKMSKeystore_ResolvesLatestEnabledVersion(t *testing.T) {
	oldKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	newKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	disabledKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{KeyType: keystore.ECDSA_S256, KeyID: keyName, VersionNumber: 1, PrivateKey: internal.NewRaw(crypto.FromECDSA(oldKey))},
		{KeyType: keystore.ECDSA_S256, KeyID: keyName, VersionNumber: 2, PrivateKey: internal.NewRaw(crypto.FromECDSA(newKey))},
		{
			KeyType:       keystore.ECDSA_S256,
			KeyID:         keyName,
			VersionNumber: 3,
			State:         kmspb.CryptoKeyVersion_DISABLED,
			PrivateKey:    internal.NewRaw(crypto.FromECDSA(disabledKey)),
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("highest enabled version wins", func(t *testing.T) {
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyName}})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, crypto.FromECDSAPub(&newKey.PublicKey), resp.Keys[0].KeyInfo.PublicKey)
	})
	t.Run("signing uses the same version", func(t *testing.T) {
		signResp, err := ks.Sign(ctx, keystore.SignRequest{KeyName: keyName, Data: make([]byte, 32)})
		require.NoError(t, err)
		verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
			KeyType:   keystore.ECDSA_S256,
			PublicKey: crypto.FromECDSAPub(&newKey.PublicKey),
			Data:      make([]byte, 32),
			Signature: signResp.Signature,
		})
		require.NoError(t, err)
		require.True(t, verifyResp.Valid)
	})
	t.Run("pinning an older version", func(t *testing.T) {
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyName + "/cryptoKeyVersions/1"}})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, crypto.FromECDSAPub(&oldKey.PublicKey), resp.Keys[0].KeyInfo.PublicKey)
	})
	t.Run("pinning a disabled version fails", func(t *testing.T) {
		_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyName + "/cryptoKeyVersions/3"}})
		require.ErrorContains(t, err, "is not enabled")
	})
}

// A key ring is commonly shared, so listing it must skip keys this keystore cannot use instead of
// failing the whole call.
func TestGCPKMSKeystore_ListSkipsUnusableKeys(t *testing.T) {
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	otherKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	unsupportedKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	disabledKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{KeyType: keystore.ECDSA_S256, KeyID: keyName, PrivateKey: internal.NewRaw(crypto.FromECDSA(signingKey))},
		{
			// An encryption key that happens to live in the same key ring.
			KeyType:    keystore.ECDSA_S256,
			KeyID:      keyRingName + "/cryptoKeys/encrypt-key",
			Purpose:    kmspb.CryptoKey_ENCRYPT_DECRYPT,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(otherKey)),
		},
		{
			// A signing key on a curve this keystore does not support.
			KeyType:    keystore.ECDSA_S256,
			KeyID:      keyRingName + "/cryptoKeys/p256-key",
			Algorithm:  kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(unsupportedKey)),
		},
		{
			// A supported key whose only version has been disabled.
			KeyType:    keystore.ECDSA_S256,
			KeyID:      keyRingName + "/cryptoKeys/disabled-key",
			State:      kmspb.CryptoKeyVersion_DISABLED,
			PrivateKey: internal.NewRaw(crypto.FromECDSA(disabledKey)),
		},
	})
	require.NoError(t, err)
	ks, err := gcpkms.NewKeystore(fakeClient, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)
	ctx := t.Context()

	resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, keyName, resp.Keys[0].KeyInfo.Name)

	// Explicitly requesting an unusable key still surfaces the error.
	_, err = ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{keyRingName + "/cryptoKeys/p256-key"},
	})
	require.ErrorContains(t, err, "unsupported Cloud KMS key algorithm")
	_, err = ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{keyRingName + "/cryptoKeys/disabled-key"},
	})
	require.ErrorContains(t, err, "has no enabled version")
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
	ks, err := gcpkms.NewKeystore(fakeClient, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)

	// Must error rather than panic inside crypto/ed25519.
	_, err = ks.Sign(t.Context(), keystore.SignRequest{KeyName: keyName, Data: []byte("hello")})
	require.ErrorContains(t, err, "invalid Ed25519 private key length")
}

// A least-privilege deployment holds per-CryptoKey permissions only, with no ring-wide
// cloudkms.cryptoKeys.list. Such a client implements Client but not KeyRingLister, and must still be
// able to do everything except enumerate the ring.
type noListClient struct {
	gcpkms.Client // embedded as an interface: promotes only Client's methods, not ListCryptoKeys
}

func TestGCPKMSKeystore_ClientWithoutKeyRingLister(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	fakeClient, err := gcpkms.NewFakeGCPKMSClient([]gcpkms.Key{
		{KeyType: keystore.ECDSA_S256, KeyID: keyName, PrivateKey: internal.NewRaw(crypto.FromECDSA(key))},
	})
	require.NoError(t, err)

	var client gcpkms.Client = noListClient{Client: fakeClient}
	_, isLister := client.(gcpkms.KeyRingLister)
	require.False(t, isLister, "test double must not expose ListCryptoKeys")

	ks, err := gcpkms.NewKeystore(client, gcpkms.KeystoreOptions{KeyRingName: keyRingName})
	require.NoError(t, err)
	ctx := t.Context()

	t.Run("explicit key names work", func(t *testing.T) {
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{KeyNames: []string{keyName}})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, crypto.FromECDSAPub(&key.PublicKey), resp.Keys[0].KeyInfo.PublicKey)
	})
	t.Run("signing works", func(t *testing.T) {
		signResp, err := ks.Sign(ctx, keystore.SignRequest{KeyName: keyName, Data: make([]byte, 32)})
		require.NoError(t, err)
		verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
			KeyType:   keystore.ECDSA_S256,
			PublicKey: crypto.FromECDSAPub(&key.PublicKey),
			Data:      make([]byte, 32),
			Signature: signResp.Signature,
		})
		require.NoError(t, err)
		require.True(t, verifyResp.Valid)
	})
	t.Run("listing reports a clear error", func(t *testing.T) {
		_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
		require.ErrorContains(t, err, "does not implement KeyRingLister")
	})
}
