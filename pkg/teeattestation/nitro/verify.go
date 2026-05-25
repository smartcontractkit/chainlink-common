package nitro

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/fxamacker/cbor/v2"
)

var (
	errBadCOSESign1Structure         = errors.New("data is not a COSE Sign1 array")
	errEmptyProtectedSection         = errors.New("COSE Sign1 protected section is nil or empty")
	errEmptyPayloadSection           = errors.New("COSE Sign1 payload section is nil or empty")
	errEmptySignatureSection         = errors.New("COSE Sign1 signature section is nil or empty")
	errUnsupportedSignatureAlgorithm = errors.New("COSE Sign1 algorithm is not ECDSA P-384")
	errBadAttestationDocument        = errors.New("bad attestation document")
	errMandatoryFieldsMissing        = errors.New("attestation document is missing mandatory fields")
	errBadDigest                     = errors.New("attestation digest is not SHA384")
	errBadTimestamp                  = errors.New("attestation timestamp is 0")
	errBadPCRs                       = errors.New("attestation pcrs is less than 1 or more than 32")
	errBadPCRIndex                   = errors.New("attestation pcr index is not in [0, 32)")
	errBadPCRValue                   = errors.New("attestation pcr value length is invalid")
	errBadCABundle                   = errors.New("attestation cabundle is empty")
	errBadCABundleItem               = errors.New("attestation cabundle item is empty or too large")
	errBadPublicKey                  = errors.New("attestation public_key length is invalid")
	errBadUserData                   = errors.New("attestation user_data length is invalid")
	errBadNonce                      = errors.New("attestation nonce length is invalid")
	errBadCertificatePublicKeyAlgo   = errors.New("attestation certificate public key algorithm is not ECDSA")
	errBadCertificateSigningAlgo     = errors.New("attestation certificate signature algorithm is not ECDSAWithSHA384")
	errBadSignature                  = errors.New("attestation signature does not match certificate")
)

type verifyResult struct {
	document    *attestationDocument
	signatureOK bool
}

type attestationDocument struct {
	ModuleID    string          `cbor:"module_id"`
	Timestamp   uint64          `cbor:"timestamp"`
	Digest      string          `cbor:"digest"`
	PCRs        map[uint][]byte `cbor:"pcrs"`
	Certificate []byte          `cbor:"certificate"`
	CABundle    [][]byte        `cbor:"cabundle"`
	PublicKey   []byte          `cbor:"public_key,omitempty"`
	UserData    []byte          `cbor:"user_data,omitempty"`
	Nonce       []byte          `cbor:"nonce,omitempty"`
}

type coseProtectedHeader struct {
	Alg any `cbor:"1,keyasint,omitempty"`
}

type coseSign1 struct {
	_           struct{} `cbor:",toarray"` //nolint:revive // idiomatic CBOR array encoding
	Protected   []byte
	Unprotected cbor.RawMessage
	Payload     []byte
	Signature   []byte
}

type coseSignatureInput struct {
	_           struct{} `cbor:",toarray"` //nolint:revive // idiomatic CBOR array encoding
	Context     string
	Protected   []byte
	ExternalAAD []byte
	Payload     []byte
}

func verifyAttestationDocument(data []byte, roots *x509.CertPool, currentTime time.Time) (*verifyResult, error) {
	var sign1 coseSign1
	if err := cbor.Unmarshal(data, &sign1); err != nil {
		return nil, errBadCOSESign1Structure
	}
	if len(sign1.Protected) == 0 {
		return nil, errEmptyProtectedSection
	}
	if len(sign1.Payload) == 0 {
		return nil, errEmptyPayloadSection
	}
	if len(sign1.Signature) == 0 {
		return nil, errEmptySignatureSection
	}

	var protected coseProtectedHeader
	if err := cbor.Unmarshal(sign1.Protected, &protected); err != nil {
		return nil, errBadCOSESign1Structure
	}
	if err := validateProtectedAlgorithm(protected.Alg); err != nil {
		return nil, err
	}

	var doc attestationDocument
	if err := cbor.Unmarshal(sign1.Payload, &doc); err != nil {
		return nil, errBadAttestationDocument
	}
	if err := validateAttestationPayload(&doc); err != nil {
		return nil, err
	}

	leafCert, intermediates, err := parseCertificateChain(&doc)
	if err != nil {
		return nil, err
	}
	if currentTime.IsZero() {
		currentTime = time.Now()
	}
	if _, err := leafCert.Verify(x509.VerifyOptions{
		Intermediates: intermediates,
		Roots:         roots,
		CurrentTime:   currentTime,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, err
	}

	sigStructure, err := cbor.Marshal(&coseSignatureInput{
		Context:     "Signature1",
		Protected:   sign1.Protected,
		ExternalAAD: []byte{},
		Payload:     sign1.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("build signature structure: %w", err)
	}

	pubKey, ok := leafCert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errBadCertificatePublicKeyAlgo
	}
	signatureOK := verifyECDSASignature(pubKey, sigStructure, sign1.Signature)
	if !signatureOK {
		return &verifyResult{document: &doc, signatureOK: false}, errBadSignature
	}

	return &verifyResult{document: &doc, signatureOK: true}, nil
}

func validateProtectedAlgorithm(alg any) error {
	switch v := alg.(type) {
	case int64:
		if v == -35 {
			return nil
		}
	case string:
		if v == "ES384" {
			return nil
		}
	}
	return errUnsupportedSignatureAlgorithm
}

func validateAttestationPayload(doc *attestationDocument) error {
	if doc.ModuleID == "" || doc.Digest == "" || doc.Timestamp == 0 || doc.PCRs == nil || doc.Certificate == nil || doc.CABundle == nil {
		return errMandatoryFieldsMissing
	}
	if doc.Digest != "SHA384" {
		return errBadDigest
	}
	if doc.Timestamp < 1 {
		return errBadTimestamp
	}
	if len(doc.PCRs) < 1 || len(doc.PCRs) > 32 {
		return errBadPCRs
	}
	for idx, value := range doc.PCRs {
		if idx > 31 {
			return errBadPCRIndex
		}
		if value == nil || (len(value) != 32 && len(value) != 48 && len(value) != 64) {
			return errBadPCRValue
		}
	}
	if len(doc.CABundle) < 1 {
		return errBadCABundle
	}
	for _, item := range doc.CABundle {
		if item == nil || len(item) < 1 || len(item) > 1024 {
			return errBadCABundleItem
		}
	}
	if doc.PublicKey != nil && len(doc.PublicKey) > 1024 {
		return errBadPublicKey
	}
	if doc.UserData != nil && len(doc.UserData) > 1024 {
		return errBadUserData
	}
	if doc.Nonce != nil && len(doc.Nonce) > 1024 {
		return errBadNonce
	}
	return nil
}

func parseCertificateChain(doc *attestationDocument) (*x509.Certificate, *x509.CertPool, error) {
	leafCert, err := x509.ParseCertificate(doc.Certificate)
	if err != nil {
		return nil, nil, err
	}
	if leafCert.PublicKeyAlgorithm != x509.ECDSA {
		return nil, nil, errBadCertificatePublicKeyAlgo
	}
	if leafCert.SignatureAlgorithm != x509.ECDSAWithSHA384 {
		return nil, nil, errBadCertificateSigningAlgo
	}

	intermediates := x509.NewCertPool()
	for _, der := range doc.CABundle {
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, nil, err
		}
		intermediates.AddCert(cert)
	}
	return leafCert, intermediates, nil
}

// curveKeySize returns the byte length of one ECDSA signature component
// (r or s) for the given curve: ceil(bitSize / 8). This is the correct
// size per RFC 9053 section 2.1, and differs from the hash length for P-521
// (key component = 66 bytes, hash = 64 bytes).
func curveKeySize(publicKey *ecdsa.PublicKey) int {
	return (publicKey.Curve.Params().BitSize + 7) / 8
}

func verifyECDSASignature(publicKey *ecdsa.PublicKey, sigStructure, signature []byte) bool {
	keySize := curveKeySize(publicKey)
	if len(signature) != 2*keySize {
		return false
	}

	hash, ok := hashForCurve(publicKey, sigStructure)
	if !ok {
		return false
	}

	r := new(big.Int).SetBytes(signature[:keySize])
	s := new(big.Int).SetBytes(signature[keySize:])
	return ecdsa.Verify(publicKey, hash, r, s)
}

func hashForCurve(publicKey *ecdsa.PublicKey, sigStructure []byte) ([]byte, bool) {
	switch publicKey.Curve.Params().Name {
	case "P-224":
		sum := sha256.Sum224(sigStructure)
		return sum[:], true
	case "P-256":
		sum := sha256.Sum256(sigStructure)
		return sum[:], true
	case "P-384":
		sum := sha512.Sum384(sigStructure)
		return sum[:], true
	case "P-521":
		sum := sha512.Sum512(sigStructure)
		return sum[:], true
	default:
		return nil, false
	}
}
