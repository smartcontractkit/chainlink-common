package nitro

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
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

// --- RFC / cose-wg test vectors ---
//
// These vectors come from the COSE Working Group Examples repository
// (https://github.com/cose-wg/Examples), which is the normative test
// suite for RFC 9052/9053 (and predecessor RFC 8152).
//
// verifyCOSESign1WithKey exercises the same code path as
// verifyAttestationDocument (CBOR unmarshal -> Sig_structure
// construction -> verifyECDSASignature) but without attestation-
// specific payload validation, so we can test against standard vectors.

func verifyCOSESign1WithKey(t *testing.T, coseHex string, pub *ecdsa.PublicKey) bool {
	t.Helper()

	data, err := hex.DecodeString(coseHex)
	require.NoError(t, err)

	var sign1 coseSign1
	require.NoError(t, cbor.Unmarshal(data, &sign1))

	sigStructure, err := cbor.Marshal(&coseSignatureInput{
		Context:     "Signature1",
		Protected:   sign1.Protected,
		ExternalAAD: []byte{},
		Payload:     sign1.Payload,
	})
	require.NoError(t, err)

	return verifyECDSASignature(pub, sigStructure, sign1.Signature)
}

func ecdsaPubKey(curve elliptic.Curve, xHex, yHex string) *ecdsa.PublicKey {
	x := new(big.Int)
	x.SetString(xHex, 16)
	y := new(big.Int)
	y.SetString(yHex, 16)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

// P-256 key from cose-wg/Examples (kid "11").
var coseWGKeyP256 = ecdsaPubKey(
	elliptic.P256(),
	"bac5b11cad8f99f9c72b05cf4b9e26d244dc189f745228255a219a86d6a09eff",
	"20138bf82dc1b6d562be0fa54ab7804a3a64b6d72ccfed6b6fb6ed28bbfc117e",
)

// P-384 key from cose-wg/Examples (kid "P384").
var coseWGKeyP384 = ecdsaPubKey(
	elliptic.P384(),
	"9132723f6292b010619dbe248d698c17b58756c639e7150f81bee4eb8ac37236ad0a1a19d67be32a66263e1e524d129c",
	"98cd3078c554d832ac603c4326410ff61662459b41f1f3df5dbcc83598ff7c5ed8411ca735679d1c4cb3009397d9ef2c",
)

// P-521 key from cose-wg/Examples (kid "bilbo.baggins@hobbiton.example").
var coseWGKeyP521 = ecdsaPubKey(
	elliptic.P521(),
	"0072992cb3ac08ecf3e5c63dedec0d51a8c1f79ef2f82f94f3c737bf5de7986671eac625fe8257bbd0394644caaa3aaf8f27a4585fbbcad0f2457620085e5c8f42ad",
	"01dca6947bce88bc5790485ac97427342bc35f887d86d65a089377e247e60baa55e4e8501e2ada5724ac51d6909008033ebc10ac999b9d7f5cc2519f3fe1ea1d9475",
)

// TestCOSESign1_RFC8152_AppendixC21_ES256 verifies against the canonical
// RFC 8152 Appendix C.2.1 test vector (also sign-pass-03 in cose-wg/Examples).
// ES256 (P-256), no external AAD. The hex includes the CBOR tag 18 prefix.
func TestCOSESign1_RFC8152_AppendixC21_ES256(t *testing.T) {
	// COSE_Sign1 with CBOR tag 18.
	coseHex := "D28443A10126A10442313154546869732069732074686520636F6E74656E742E" +
		"58408EB33E4CA31D1C465AB05AAC34CC6B23D58FEF5C083106C4D25A91AEF0B0117E" +
		"2AF9A291AA32E14AB834DC56ED2A223444547E01F11D3B0916E5A4C345CACB36"

	assert.True(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP256),
		"RFC 8152 C.2.1 ES256 vector must verify")
}

// TestCOSESign1_ES384 verifies ecdsa-sig-02 from cose-wg/Examples.
// ES384 (P-384), no external AAD.
func TestCOSESign1_ES384(t *testing.T) {
	coseHex := "D28444A1013822A104445033383454546869732069732074686520636F6E74656E742E" +
		"58605F150ABD1C7D25B32065A14E05D6CB1F665D10769FF455EA9A2E0ADAB5DE63838D" +
		"B257F0949C41E13330E110EBA7B912F34E1546FB1366A2568FAA91EC3E6C8D42F4A67A" +
		"0EDF731D88C9AEAD52258B2E2C4740EF614F02E9D91E9B7B59622A3C"

	assert.True(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP384),
		"cose-wg ecdsa-sig-02 ES384 vector must verify")
}

// TestCOSESign1_ES512 verifies ecdsa-sig-03 from cose-wg/Examples.
// ES512 (P-521), no external AAD.
func TestCOSESign1_ES512(t *testing.T) {
	coseHex := "D28444A1013823A104581E62696C626F2E62616767696E7340686F626269746F6E2E" +
		"6578616D706C6554546869732069732074686520636F6E74656E742E588401664DD696" +
		"2091B5100D6E1833D503539330EC2BC8FD3E8996950CE9F70259D9A30F73794F603B0D" +
		"3E7C5E9C4C2A57E10211F76E79DF8FFD1B79D7EF5B9FA7DA109001965FA2D37E093BB" +
		"13C040399C467B3B9908C09DB2B0F1F4996FE07BB02AAA121A8E1C671F3F997ADE7D65" +
		"1081017057BD3A8A5FBF394972EA71CFDC15E6F8FE2E1"

	assert.True(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP521),
		"cose-wg ecdsa-sig-03 ES512 vector must verify")
}

// TestCOSESign1_FailTamperedPayload verifies sign-fail-02 from cose-wg/Examples.
// The last byte of the payload was changed from 0x2E ('.') to 0x2F ('/').
// The signature was computed over the original payload, so verification must fail.
func TestCOSESign1_FailTamperedPayload(t *testing.T) {
	coseHex := "D28443A10126A10442313154546869732069732074686520636F6E74656E742F" +
		"58408EB33E4CA31D1C465AB05AAC34CC6B23D58FEF5C083106C4D25A91AEF0B0117E" +
		"2AF9A291AA32E14AB834DC56ED2A223444547E01F11D3B0916E5A4C345CACB36"

	assert.False(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP256),
		"sign-fail-02: tampered payload must not verify")
}

// TestCOSESign1_FailModifiedProtectedHeader verifies sign-fail-06 from
// cose-wg/Examples. An extra attribute (ctyp: 0) was added to the protected
// header after signing. The Sig_structure changes, so the signature is invalid.
func TestCOSESign1_FailModifiedProtectedHeader(t *testing.T) {
	coseHex := "D28445A201260300A10442313154546869732069732074686520636F6E74656E742E" +
		"58408EB33E4CA31D1C465AB05AAC34CC6B23D58FEF5C083106C4D25A91AEF0B0117E" +
		"2AF9A291AA32E14AB834DC56ED2A223444547E01F11D3B0916E5A4C345CACB36"

	assert.False(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP256),
		"sign-fail-06: modified protected header must not verify")
}

// TestCOSESign1_WrongKeyRejectsValidVector uses the valid RFC 8152 C.2.1
// vector but verifies with the P-384 key. Must fail.
func TestCOSESign1_WrongKeyRejectsValidVector(t *testing.T) {
	coseHex := "D28443A10126A10442313154546869732069732074686520636F6E74656E742E" +
		"58408EB33E4CA31D1C465AB05AAC34CC6B23D58FEF5C083106C4D25A91AEF0B0117E" +
		"2AF9A291AA32E14AB834DC56ED2A223444547E01F11D3B0916E5A4C345CACB36"

	assert.False(t, verifyCOSESign1WithKey(t, coseHex, coseWGKeyP384),
		"valid ES256 vector must not verify with P-384 key")
}
