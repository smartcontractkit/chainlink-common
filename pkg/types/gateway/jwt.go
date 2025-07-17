package gateway

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"time"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/golang-jwt/jwt/v5"

	jsonrpc "github.com/smartcontractkit/chainlink-common/pkg/jsonrpc2"
)

const (
	defaultJWTExpiryDuration = time.Hour
)

// Option is a function type that allows configuring CreateRequestJWT.
type Option func(*jwtOptions)

type jwtOptions struct {
	expiryDuration *time.Duration
	issuer         *string  // New field for optional issuer
	audience       []string // New field for optional audience
	subject        *string  // New field for optional subject
}

func WithExpiry(d time.Duration) Option {
	return func(opts *jwtOptions) {
		opts.expiryDuration = &d
	}
}

func WithIssuer(issuer string) Option {
	return func(opts *jwtOptions) {
		opts.issuer = &issuer
	}
}

func WithAudience(audience []string) Option {
	return func(opts *jwtOptions) {
		opts.audience = audience
	}
}

func WithSubject(subject string) Option {
	return func(opts *jwtOptions) {
		opts.subject = &subject
	}
}

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

	pubKeyBytes := append(ecdsaKey.X.Bytes(), ecdsaKey.Y.Bytes()...)
	fullPubKey := append([]byte{0x04}, pubKeyBytes...)
	secpPubKey, err := secp.ParsePubKey(fullPubKey)
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

// CreateRequestJWT creates an unsigned JWT for a JSON-RPC request
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
//		digest: "<request-digest>",      // 32 byte hex string with "0x" prefix
//		iss: "<compresssed-public-key>", // compressed ECDSA public key in hex format with "0x" prefix
//		exp: <timestamp>,                // expiration time (Unix timestamp)
//		iat: <timestamp>                 // issued at time (Unix timestamp)
//	}
//
// sample payload:
//
//	{
//	  "digest": "0x4a1f2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a",
//	  "iss": "0x03a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b",
//	  "exp": 1717600000,
//	  "iat": 1717596400
//	}
//
// signature: ECDSA signature of the header and payload using the private key
func CreateRequestJWT[T any](req jsonrpc.Request[T], key *ecdsa.PublicKey, opts ...Option) (*jwt.Token, error) {
	if key == nil {
		return nil, errors.New("public key cannot be nil")
	}

	// Apply options
	options := &jwtOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Set defaults if not provided
	expiryDuration := defaultJWTExpiryDuration
	if options.expiryDuration != nil {
		expiryDuration = *options.expiryDuration
	}

	digest, err := req.Digest()
	if err != nil {
		return nil, err
	}

	var issuer string
	if options.issuer != nil {
		issuer = *options.issuer
	}

	var subject string
	if options.subject != nil {
		subject = *options.subject
	}

	var audience []string
	if options.audience != nil {
		audience = options.audience
	}

	now := time.Now()
	claims := JWTClaims{
		Digest: "0x" + digest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings(audience),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	return jwt.NewWithClaims(SigningMethodES256K, claims), nil
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
	reqDigest, err := req.Digest()
	if err != nil {
		return nil, err
	}
	if claims.Digest != "0x"+reqDigest {
		return nil, errors.New("JWT digest does not match request digest")
	}
	return claims, nil
}
