package nitro

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// coseSign encodes an ECDSA signature as r||s with each component
// zero-padded to keySize bytes, matching the COSE Sign1 encoding
// (RFC 9053 section 2.1).
func coseSign(t *testing.T, key *ecdsa.PrivateKey, hash []byte) []byte {
	t.Helper()
	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	require.NoError(t, err)

	keySize := (key.Curve.Params().BitSize + 7) / 8
	sig := make([]byte, 2*keySize)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[keySize-len(rBytes):keySize], rBytes)
	copy(sig[2*keySize-len(sBytes):2*keySize], sBytes)
	return sig
}

func TestHashForCurve_AllCurves(t *testing.T) {
	curves := []struct {
		name     string
		curve    elliptic.Curve
		hashSize int
	}{
		{"P-224", elliptic.P224(), sha256.Size224},
		{"P-256", elliptic.P256(), sha256.Size},
		{"P-384", elliptic.P384(), sha512.Size384},
		{"P-521", elliptic.P521(), sha512.Size},
	}

	payload := []byte("test payload for hashing")

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			key, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			require.NoError(t, err)

			hash, ok := hashForCurve(&key.PublicKey, payload)
			assert.True(t, ok, "hashForCurve should succeed for %s", tc.name)
			assert.Len(t, hash, tc.hashSize, "wrong hash length for %s", tc.name)
		})
	}
}

func TestHashForCurve_UnsupportedCurve(t *testing.T) {
	// Use a custom curve params to simulate an unsupported curve.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Mutate the curve name to something unsupported.
	fakeCurve := *key.PublicKey.Curve.Params()
	fakeCurve.Name = "P-999"

	fakeKey := &ecdsa.PublicKey{
		Curve: &fakeCurve,
		X:     key.PublicKey.X,
		Y:     key.PublicKey.Y,
	}

	hash, ok := hashForCurve(fakeKey, []byte("data"))
	assert.False(t, ok)
	assert.Nil(t, hash)
}

func TestVerifyECDSASignature_RoundTrip(t *testing.T) {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P-224", elliptic.P224()},
		{"P-256", elliptic.P256()},
		{"P-384", elliptic.P384()},
		{"P-521", elliptic.P521()},
	}

	payload := []byte("COSE Signature1 structure payload")

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			key, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			require.NoError(t, err)

			hash, ok := hashForCurve(&key.PublicKey, payload)
			require.True(t, ok)

			sig := coseSign(t, key, hash)
			assert.True(t, verifyECDSASignature(&key.PublicKey, payload, sig),
				"valid signature should verify for %s", tc.name)
		})
	}
}

func TestVerifyECDSASignature_WrongPayload(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	hash, _ := hashForCurve(&key.PublicKey, []byte("original payload"))
	sig := coseSign(t, key, hash)

	assert.False(t, verifyECDSASignature(&key.PublicKey, []byte("different payload"), sig),
		"signature for different payload should not verify")
}

func TestVerifyECDSASignature_WrongKey(t *testing.T) {
	key1, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	key2, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	payload := []byte("payload")
	hash, _ := hashForCurve(&key1.PublicKey, payload)
	sig := coseSign(t, key1, hash)

	assert.False(t, verifyECDSASignature(&key2.PublicKey, payload, sig),
		"signature should not verify with wrong public key")
}

func TestVerifyECDSASignature_TamperedSignature(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	payload := []byte("payload")
	hash, _ := hashForCurve(&key.PublicKey, payload)
	sig := coseSign(t, key, hash)

	// Flip a bit in the signature.
	sig[len(sig)/2] ^= 0x01

	assert.False(t, verifyECDSASignature(&key.PublicKey, payload, sig),
		"tampered signature should not verify")
}

func TestVerifyECDSASignature_WrongLength(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	payload := []byte("payload")

	// Too short.
	assert.False(t, verifyECDSASignature(&key.PublicKey, payload, []byte("short")))

	// Too long.
	assert.False(t, verifyECDSASignature(&key.PublicKey, payload, make([]byte, 200)))
}

func TestCurveKeySize(t *testing.T) {
	tests := []struct {
		name     string
		curve    elliptic.Curve
		expected int
	}{
		{"P-224", elliptic.P224(), 28},
		{"P-256", elliptic.P256(), 32},
		{"P-384", elliptic.P384(), 48},
		{"P-521", elliptic.P521(), 66}, // ceil(521/8) = 66, NOT 64
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, curveKeySize(&key.PublicKey))
		})
	}
}

// TestP521SignatureSize verifies that the P-521 signature component size
// (66 bytes) differs from the SHA-512 hash size (64 bytes). This is the
// specific edge case that was previously broken: using hash length to
// determine signature component size produces the wrong answer for P-521.
func TestP521SignatureSize(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err)

	hash, ok := hashForCurve(&key.PublicKey, []byte("data"))
	require.True(t, ok)
	assert.Len(t, hash, 64, "SHA-512 hash should be 64 bytes")

	keySize := curveKeySize(&key.PublicKey)
	assert.Equal(t, 66, keySize, "P-521 key component should be 66 bytes")
	assert.NotEqual(t, len(hash), keySize,
		"P-521 key size must differ from hash size; using hash length for signature components is wrong")

	// Now verify that a real P-521 sign/verify roundtrip works.
	sig := coseSign(t, key, hash)
	assert.Len(t, sig, 132, "P-521 COSE signature should be 132 bytes (2x66)")
	assert.True(t, verifyECDSASignature(&key.PublicKey, []byte("data"), sig))

	// A 128-byte signature (2x64, the old broken size) must be rejected,
	// regardless of content.
	wrongSig := make([]byte, 128)
	assert.False(t, verifyECDSASignature(&key.PublicKey, []byte("data"), wrongSig),
		"128-byte signature should be rejected for P-521")
}

// TestVerifyECDSASignature_DERSignatureRejected ensures that standard
// ASN.1/DER-encoded signatures are rejected. COSE uses raw r||s encoding,
// not DER. This guards against accidentally accepting the wrong format.
func TestVerifyECDSASignature_DERSignatureRejected(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	payload := []byte("payload")
	hash, _ := hashForCurve(&key.PublicKey, payload)

	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	require.NoError(t, err)

	// Build a DER-encoded signature (ASN.1 SEQUENCE of two INTEGERs).
	derSig := marshalDER(r, s)

	assert.False(t, verifyECDSASignature(&key.PublicKey, payload, derSig),
		"DER-encoded signature must be rejected; COSE uses raw r||s")
}

// marshalDER produces a minimal ASN.1 DER encoding of an ECDSA signature.
func marshalDER(r, s *big.Int) []byte {
	encodeInt := func(v *big.Int) []byte {
		b := v.Bytes()
		if len(b) > 0 && b[0]&0x80 != 0 {
			b = append([]byte{0x00}, b...)
		}
		return append([]byte{0x02, byte(len(b))}, b...)
	}
	rEnc := encodeInt(r)
	sEnc := encodeInt(s)
	seq := append(rEnc, sEnc...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}
