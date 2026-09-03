package kms_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	kms "github.com/smartcontractkit/chainlink-common/keystore/kms"
)

func TestSEC1ToASN1PublicKey(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Geth library uses SEC1 format.
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)
	require.Len(t, sec1PubKey, 65)
	require.Equal(t, byte(0x04), sec1PubKey[0])

	// Convert to ASN.1
	asn1PubKey, err := kms.SEC1ToASN1PublicKey(sec1PubKey)
	require.NoError(t, err)

	// Convert back to SEC1
	sec1PubKey2, err := kms.ASN1ToSEC1PublicKey(asn1PubKey)
	require.NoError(t, err)
	require.Len(t, sec1PubKey2, 65)
	require.Equal(t, byte(0x04), sec1PubKey2[0])
	pubKey := privateKey.PublicKey
	require.Equal(t, pubKey.X.Bytes(), sec1PubKey2[1:33])
	require.Equal(t, pubKey.Y.Bytes(), sec1PubKey2[33:65])
}

func TestASN1SignatureToSEC1Signature(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)

	hash := crypto.Keccak256Hash([]byte("test"))

	sig, err := crypto.Sign(hash[:], privateKey)
	require.NoError(t, err)

	asn1Sig, err := kms.SEC1ToASN1Sig(sig)
	require.NoError(t, err)

	// We pass the expected SEC1 public key for verification.
	sec1Sig, err := kms.ASN1ToSEC1Sig(asn1Sig, sec1PubKey, hash[:])
	require.NoError(t, err)
	require.Len(t, sec1Sig, 65)
	require.Equal(t, sig, sec1Sig)
}

// TestASN1SignatureToSEC1SignatureHighS exercises the EIP-2 high-S normalization branch in
// ASN1ToSEC1Sig. go-ethereum's crypto.Sign always returns a low-S signature (S <= N/2), so the
// happy path never covers it. A non-normalizing signer such as Cloud KMS emits S > N/2 roughly
// half the time, which is exactly the input this test builds: same R, S negated (S' = N - S).
func TestASN1SignatureToSEC1SignatureHighS(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)

	hash := crypto.Keccak256Hash([]byte("high-S test"))

	// crypto.Sign normalizes S to <= N/2 (decred SignCompact). Confirm the assumption.
	sig, err := crypto.Sign(hash[:], privateKey)
	require.NoError(t, err)
	require.Len(t, sig, 65)

	n := crypto.S256().Params().N
	halfN := new(big.Int).Div(n, big.NewInt(2))
	s := new(big.Int).SetBytes(sig[32:64])
	require.LessOrEqual(t, s.Cmp(halfN), 0, "crypto.Sign must produce a low-S signature")

	// Negate S to obtain the high-S (S > N/2) value a non-normalizing signer would emit.
	highS := new(big.Int).Sub(n, s)
	require.Positive(t, highS.Cmp(halfN), "negating S must yield a high-S signature")

	// Rebuild a SEC1 signature carrying the high S, then DER-encode it as Cloud KMS would.
	highSSec1 := slices.Concat(sig[:32], highS.FillBytes(make([]byte, 32)), []byte{0})
	highSDER, err := kms.SEC1ToASN1Sig(highSSec1)
	require.NoError(t, err)

	// ASN1ToSEC1Sig must flip S back below N/2 per EIP-2 and recover the correct V.
	sec1Sig, err := kms.ASN1ToSEC1Sig(highSDER, sec1PubKey, hash[:])
	require.NoError(t, err)
	require.Len(t, sec1Sig, 65)

	// The normalized signature must be byte-for-byte the original low-S signature (same R, S, V).
	require.Equal(t, sig, sec1Sig)
	require.Equal(t, sig[64], sec1Sig[64], "recovery id must match the low-S signature")

	recovered, err := crypto.Ecrecover(hash[:], sec1Sig)
	require.NoError(t, err)
	require.Equal(t, sec1PubKey, recovered, "normalized high-S signature must recover the signer")
}
