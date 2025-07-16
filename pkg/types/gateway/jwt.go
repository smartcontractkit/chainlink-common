package gateway

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/golang-jwt/jwt/v5"

	jsonrpc "github.com/smartcontractkit/chainlink-common/pkg/jsonrpc2"
)

var SigningMethodES256K *SigningMethodSecp256k1

func init() {
	SigningMethodES256K = &SigningMethodSecp256k1{}
	// golang-jwt library does not support ES256K (ECDSA on the secp256k1 curve)
	// so registering a custom implementation of the signing method here
	jwt.RegisterSigningMethod(SigningMethodES256K.Alg(), func() jwt.SigningMethod {
		return SigningMethodES256K
	})
}

// SigningMethodSecp256k1 implements the ES256K signing method for JWT
type SigningMethodSecp256k1 struct{}

func (m *SigningMethodSecp256k1) Alg() string {
	return "ES256K"
}

func (m *SigningMethodSecp256k1) Sign(signingString string, key interface{}) ([]byte, error) {
	var ecdsaKey *ecdsa.PrivateKey
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		ecdsaKey = k
	default:
		return nil, jwt.ErrInvalidKeyType
	}
	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	hash := hasher.Sum(nil)

	secpPrivKey := secp.PrivKeyFromBytes(ecdsaKey.D.Bytes())
	signature := secpecdsa.Sign(secpPrivKey, hash)

	return signature.Serialize(), nil
}

func (m *SigningMethodSecp256k1) Verify(signingString string, signature []byte, key interface{}) error {
	var ecdsaKey *ecdsa.PublicKey
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		ecdsaKey = k
	default:
		return jwt.ErrInvalidKeyType
	}

	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	hash := hasher.Sum(nil)

	secpPubKey, err := secp.ParsePubKey(append([]byte{0x04}, append(ecdsaKey.X.Bytes(), ecdsaKey.Y.Bytes()...)...))
	if err != nil {
		return err
	}
	sig, err := secpecdsa.ParseDERSignature(signature)
	if err != nil {
		return err
	}
	if !sig.Verify(hash, secpPubKey) {
		return jwt.ErrSignatureInvalid
	}

	return nil
}

type JWTClaims struct {
	Digest string `json:"digest"`
	jwt.RegisteredClaims
}

// CreateRequestJWT creates a signed JWT for a JSON-RPC request
// JWT has 3 parts: header, payload, and signature as shown below
// header:
//
//	{
//		alg: "ES256K",
//		typ: "JWT"
//	}
//
// payload:
//
//	{
//		digest: "<request-digest>"
//		iss: "<compressed-public-key>",
//		exp: <timestamp>,
//		iat: <timestamp>
//	}
//
// signature: ECDSA signature of the header and payload using the private key
func CreateRequestJWT[T any](req jsonrpc.Request[T], key *ecdsa.PrivateKey, expiryDuration time.Duration) (string, error) {
	if key == nil {
		return "", errors.New("private key cannot be nil")
	}
	digest, err := req.Digest()
	if err != nil {
		return "", err
	}
	now := time.Now()
	compressed, err := compressedECDSAPubKey(&key.PublicKey)
	if err != nil {
		return "", err
	}
	claims := JWTClaims{
		Digest: digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    hex.EncodeToString(compressed),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(SigningMethodES256K, claims)
	return token.SignedString(key)
}

// VerifyRequestJWT verifies a signed JWT for a JSON-RPC request
// It verifies the signature, checks the issuer, validates the digest
// and performs all validations done by jwt.ParseWithClaims() including expiration checks
func VerifyRequestJWT[T any](tokenString string, key *ecdsa.PublicKey, req jsonrpc.Request[T]) (*JWTClaims, error) {
	if key == nil {
		return nil, errors.New("public key cannot be nil")
	}
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*SigningMethodSecp256k1); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	compressed, err := compressedECDSAPubKey(key)
	if err != nil {
		return nil, err
	}
	if claims.Issuer != hex.EncodeToString(compressed) {
		return nil, jwt.ErrTokenInvalidIssuer
	}
	reqDigest, err := req.Digest()
	if err != nil {
		return nil, err
	}
	if claims.Digest != reqDigest {
		return nil, errors.New("JWT digest does not match request digest")
	}
	return claims, nil
}

func compressedECDSAPubKey(pubKey *ecdsa.PublicKey) ([]byte, error) {
	if pubKey == nil || pubKey.X == nil || pubKey.Y == nil {
		return nil, errors.New("invalid public key")
	}
	var x, y secp256k1.FieldVal
	x.SetByteSlice(pubKey.X.Bytes())
	y.SetByteSlice(pubKey.Y.Bytes())
	return secp.NewPublicKey(&x, &y).SerializeCompressed(), nil
}
