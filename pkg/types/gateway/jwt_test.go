package gateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	jsonrpc "github.com/smartcontractkit/chainlink-common/pkg/jsonrpc2"
)

func TestES256K(t *testing.T) {
	privKey, err := secp.GeneratePrivateKey()
	require.NoError(t, err)
	privKeyBytes := privKey.Key.Bytes()
	ecdsaPrivKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp.S256(),
			X:     privKey.PubKey().X(),
			Y:     privKey.PubKey().Y(),
		},
		D: new(big.Int).SetBytes(privKeyBytes[:]),
	}
	ecdsaPubKey := &ecdsaPrivKey.PublicKey

	t.Run("ES256K JWT creation and verification", func(t *testing.T) {
		req := jsonrpc.Request[map[string]interface{}]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-es256k",
			Method:  "test.method",
			Params: &map[string]interface{}{
				"param1": "value1",
				"param2": 42,
			},
		}

		token, err := CreateRequestJWT(req, ecdsaPubKey)
		require.NoError(t, err)
		signedToken, err := token.SignedString(ecdsaPrivKey)
		require.NoError(t, err)

		claims, err := VerifyRequestJWT(signedToken, &ecdsaPrivKey.PublicKey, req)
		require.NoError(t, err)
		require.NotNil(t, claims)

		expectedDigest, err := req.Digest()
		require.NoError(t, err)
		require.Equal(t, "0x"+expectedDigest, claims.Digest)

		parts := strings.Split(signedToken, ".")
		require.Len(t, parts, 3, "JWT should have 3 parts")

		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		require.NoError(t, err)

		var header map[string]interface{}
		err = json.Unmarshal(headerBytes, &header)
		require.NoError(t, err)

		require.Equal(t, "ES256K", header["alg"])
	})

	t.Run("ES256K signature verification", func(t *testing.T) {
		// Test the signing method directly
		signingMethod := SigningMethodES256K
		require.Equal(t, "ES256K", signingMethod.Alg())

		testString := "test.signing.string"

		signature, err := signingMethod.Sign(testString, ecdsaPrivKey)
		require.NoError(t, err)
		require.NotEmpty(t, signature)

		err = signingMethod.Verify(testString, signature, &ecdsaPrivKey.PublicKey)
		require.NoError(t, err)
	})

	t.Run("ES256K wrong key verification fails", func(t *testing.T) {
		privKey2, err := secp.GeneratePrivateKey()
		require.NoError(t, err)

		privKey2Bytes := privKey2.Key.Bytes()
		ecdsaPrivKey2 := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: secp.S256(),
				X:     privKey2.PubKey().X(),
				Y:     privKey2.PubKey().Y(),
			},
			D: new(big.Int).SetBytes(privKey2Bytes[:]),
		}

		testString := "test"
		req := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-wrong-key",
			Method:  "test.method",
			Params:  &testString,
		}
		token, err := CreateRequestJWT(req, ecdsaPubKey)
		require.NoError(t, err)
		signedToken, err := token.SignedString(ecdsaPrivKey)
		require.NoError(t, err)

		// Try to verify with second key (should fail)
		_, err = VerifyRequestJWT(signedToken, &ecdsaPrivKey2.PublicKey, req)
		require.Error(t, err)
	})
}

func TestCreateRequestJWT(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
	require.NoError(t, err)

	t.Run("successful JWT creation", func(t *testing.T) {
		req := jsonrpc.Request[map[string]interface{}]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params: &map[string]interface{}{
				"param1": "value1",
				"param2": 42,
			},
		}
		unsignedToken, err := CreateRequestJWT(req, &privateKey.PublicKey, WithExpiry(time.Hour), WithIssuer("issuer"), WithSubject("subject"), WithAudience([]string{"audience1", "audience2"}))
		require.NoError(t, err)
		signedToken, err := unsignedToken.SignedString(privateKey)
		require.NoError(t, err)

		token, err := jwt.ParseWithClaims(signedToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return &privateKey.PublicKey, nil
		})
		require.NoError(t, err)
		require.True(t, token.Valid)

		claims, ok := token.Claims.(*JWTClaims)
		require.True(t, ok)
		require.NotEmpty(t, claims.Digest)
		require.Equal(t, "issuer", claims.Issuer)
		require.NotNil(t, "subject", claims.Subject)
		require.Len(t, claims.Audience, 2)
		require.Equal(t, "audience1", claims.Audience[0])
		require.Equal(t, "audience2", claims.Audience[1])
		require.NotNil(t, claims.ExpiresAt)
		require.NotNil(t, claims.IssuedAt)

		expectedDigest, err := req.Digest()
		require.NoError(t, err)
		require.Equal(t, "0x"+expectedDigest, claims.Digest)
	})

	t.Run("different expiry durations", func(t *testing.T) {
		testParam := "test-param"
		req := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params:  &testParam,
		}

		testCases := []time.Duration{
			time.Minute,
			time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
		}

		for _, duration := range testCases {
			t.Run(duration.String(), func(t *testing.T) {
				unsignedToken, err := CreateRequestJWT(req, &privateKey.PublicKey, WithExpiry(duration))
				require.NoError(t, err)
				tokenString, err := unsignedToken.SignedString(privateKey)
				require.NoError(t, err)

				token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return &privateKey.PublicKey, nil
				})
				require.NoError(t, err)

				claims := token.Claims.(*JWTClaims)
				expectedExpiry := claims.IssuedAt.Add(duration)
				require.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, time.Second)
			})
		}
	})

	t.Run("invalid request - digest failure", func(t *testing.T) {
		// Create a request with unmarshalable data
		type UnmarshalableType struct {
			Channel chan string `json:"channel"`
		}
		req := jsonrpc.Request[UnmarshalableType]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params: &UnmarshalableType{
				Channel: make(chan string),
			},
		}

		_, err := CreateRequestJWT(req, &privateKey.PublicKey)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error marshaling JSON")
	})

	t.Run("nil private key", func(t *testing.T) {
		testParam := "test-param"
		req := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params:  &testParam,
		}

		_, err := CreateRequestJWT(req, nil)
		require.Error(t, err)
	})

	t.Run("different curve", func(t *testing.T) {
		// Generate a key with a different curve (P256 instead of secp256k1)
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		testParam := "test-param"
		req := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params:  &testParam,
		}

		unsigned, err := CreateRequestJWT(req, &privateKey.PublicKey)
		require.NoError(t, err)
		tokenString, err := unsigned.SignedString(privateKey)
		require.NoError(t, err)

		_, err = jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return &privateKey.PublicKey, nil
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid public key")
	})
}

func TestVerifyRequestJWT(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
	require.NoError(t, err)

	// Create a valid token for testing
	testParam := "test-param"
	req := jsonrpc.Request[string]{
		Version: jsonrpc.JsonRpcVersion,
		ID:      "test-id",
		Method:  "test.method",
		Params:  &testParam,
	}
	validToken, err := CreateRequestJWT(req, &privateKey.PublicKey)
	require.NoError(t, err)
	signedToken, err := validToken.SignedString(privateKey)
	require.NoError(t, err)

	t.Run("successful verification", func(t *testing.T) {
		claims, err := VerifyRequestJWT(signedToken, &privateKey.PublicKey, req)

		require.NoError(t, err)
		require.NotNil(t, claims)
		require.NotEmpty(t, claims.Digest)
		expectedDigest, err := req.Digest()
		require.NoError(t, err)
		require.Equal(t, "0x"+expectedDigest, claims.Digest)
	})

	t.Run("wrong public key", func(t *testing.T) {
		wrongKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		_, err = VerifyRequestJWT(signedToken, &wrongKey.PublicKey, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := VerifyRequestJWT("invalid.token.format", &privateKey.PublicKey, req)
		require.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken, err := CreateRequestJWT(req, &privateKey.PublicKey, WithExpiry(-time.Hour))
		require.NoError(t, err)
		expiredTokenString, err := expiredToken.SignedString(privateKey)
		require.NoError(t, err)

		_, err = VerifyRequestJWT(expiredTokenString, &privateKey.PublicKey, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "token is expired")
	})

	t.Run("invalid signing method", func(t *testing.T) {
		// Create a token with HMAC instead of ECDSA
		claims := JWTClaims{
			Digest: "test-digest",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "test-issuer",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("secret"))
		require.NoError(t, err)

		_, err = VerifyRequestJWT(tokenString, &privateKey.PublicKey, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("nil public key", func(t *testing.T) {
		_, err := VerifyRequestJWT(signedToken, nil, req)
		require.Error(t, err)
	})

	t.Run("token with invalid claims type", func(t *testing.T) {
		// Create a token with standard claims instead of JWTClaims
		standardClaims := jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}
		token := jwt.NewWithClaims(SigningMethodES256K, standardClaims)
		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)

		_, err = VerifyRequestJWT(tokenString, &privateKey.PublicKey, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "JWT digest does not match request digest")
	})

	t.Run("digest mismatch - different request", func(t *testing.T) {
		// Create a different request with different digest
		differentParam := "different-param"
		differentReq := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "different-id",
			Method:  "different.method",
			Params:  &differentParam,
		}

		// Try to verify the token against the different request
		_, err := VerifyRequestJWT(signedToken, &privateKey.PublicKey, differentReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "JWT digest does not match request digest")
	})

	t.Run("digest mismatch - modified request params", func(t *testing.T) {
		// Create a request with same structure but different params
		modifiedParam := "modified-test-param"
		modifiedReq := jsonrpc.Request[string]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params:  &modifiedParam,
		}

		// Try to verify the token against the modified request
		_, err := VerifyRequestJWT(signedToken, &privateKey.PublicKey, modifiedReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "JWT digest does not match request digest")
	})

	t.Run("request digest calculation failure", func(t *testing.T) {
		// Create a request with unmarshalable data
		type UnmarshalableType struct {
			Channel chan string `json:"channel"`
		}
		invalidReq := jsonrpc.Request[UnmarshalableType]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "test-id",
			Method:  "test.method",
			Params: &UnmarshalableType{
				Channel: make(chan string),
			},
		}

		_, err := VerifyRequestJWT(signedToken, &privateKey.PublicKey, invalidReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error marshaling JSON")
	})
}

func TestJWTEndToEnd(t *testing.T) {
	t.Run("complete flow", func(t *testing.T) {
		// Generate key pair
		privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		// Create a request
		req := jsonrpc.Request[map[string]interface{}]{
			Version: jsonrpc.JsonRpcVersion,
			ID:      "e2e-test",
			Method:  "trigger.action",
			Params: &map[string]interface{}{
				"action":    "send_email",
				"recipient": "test@example.com",
				"data": map[string]interface{}{
					"subject": "Test Email",
					"body":    "This is a test email",
					"urgent":  true,
				},
			},
		}

		token, err := CreateRequestJWT(req, &privateKey.PublicKey, WithExpiry(time.Hour), WithIssuer("issuer"), WithSubject("subject"), WithAudience([]string{"audience1", "audience2"}))
		require.NoError(t, err)
		tokenString, err := token.SignedString(privateKey)
		require.NoError(t, err)
		claims, err := VerifyRequestJWT(tokenString, &privateKey.PublicKey, req)
		require.NoError(t, err)

		expectedDigest, err := req.Digest()
		require.NoError(t, err)
		require.Equal(t, "0x"+expectedDigest, claims.Digest)

		require.True(t, claims.ExpiresAt.After(time.Now()))
		require.True(t, claims.IssuedAt.Before(time.Now().Add(time.Second)))
		require.Equal(t, "issuer", claims.Issuer)
		require.Equal(t, "subject", claims.Subject)
		require.Len(t, claims.Audience, 2)
		require.Equal(t, "audience1", claims.Audience[0])
		require.Equal(t, "audience2", claims.Audience[1])
	})

	t.Run("different request types", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		testCases := []struct {
			name string
			req  interface{}
		}{
			{
				name: "string params",
				req: func() jsonrpc.Request[string] {
					param := "simple string parameter"
					return jsonrpc.Request[string]{
						Version: jsonrpc.JsonRpcVersion,
						ID:      "string-test",
						Method:  "test.string",
						Params:  &param,
					}
				}(),
			},
			{
				name: "int params",
				req: func() jsonrpc.Request[int] {
					param := 42
					return jsonrpc.Request[int]{
						Version: jsonrpc.JsonRpcVersion,
						ID:      "int-test",
						Method:  "test.int",
						Params:  &param,
					}
				}(),
			},
			{
				name: "slice params",
				req: func() jsonrpc.Request[[]string] {
					param := []string{"item1", "item2", "item3"}
					return jsonrpc.Request[[]string]{
						Version: jsonrpc.JsonRpcVersion,
						ID:      "slice-test",
						Method:  "test.slice",
						Params:  &param,
					}
				}(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var token *jwt.Token
				var err error
				var claims *JWTClaims
				var signedToken string

				switch req := tc.req.(type) {
				case jsonrpc.Request[string]:
					token, err = CreateRequestJWT(req, &privateKey.PublicKey)
					require.NoError(t, err)
					signedToken, err = token.SignedString(privateKey)
					require.NoError(t, err)
					claims, err = VerifyRequestJWT(signedToken, &privateKey.PublicKey, req)
				case jsonrpc.Request[int]:
					token, err = CreateRequestJWT(req, &privateKey.PublicKey)
					require.NoError(t, err)
					signedToken, err = token.SignedString(privateKey)
					require.NoError(t, err)
					claims, err = VerifyRequestJWT(signedToken, &privateKey.PublicKey, req)
				case jsonrpc.Request[[]string]:
					token, err = CreateRequestJWT(req, &privateKey.PublicKey)
					require.NoError(t, err)
					signedToken, err = token.SignedString(privateKey)
					require.NoError(t, err)
					claims, err = VerifyRequestJWT(signedToken, &privateKey.PublicKey, req)
				}

				require.NoError(t, err)
				require.NotNil(t, claims)
				require.NotEmpty(t, claims.Digest)
			})
		}
	})
}

func TestJWTOptions_EmptyAudience(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
	require.NoError(t, err)

	req := jsonrpc.Request[string]{
		Version: jsonrpc.JsonRpcVersion,
		ID:      "test-id",
		Method:  "test.method",
		Params:  func() *string { s := "test-param"; return &s }(),
	}

	token, err := CreateRequestJWT(req, &privateKey.PublicKey, WithAudience([]string{}))
	require.NoError(t, err)

	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	parsedToken, err := jwt.ParseWithClaims(signedToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	claims := parsedToken.Claims.(*JWTClaims)
	require.Empty(t, claims.Audience)
}

func TestSigningMethodSecp256k1(t *testing.T) {
	method := &SigningMethodSecp256k1{}

	t.Run("Alg method returns ES256K", func(t *testing.T) {
		require.Equal(t, "ES256K", method.Alg())
	})

	t.Run("Sign with invalid key type", func(t *testing.T) {
		invalidKey := "not-a-private-key"
		_, err := method.Sign("test-string", invalidKey)
		require.Error(t, err)
		require.Equal(t, jwt.ErrInvalidKeyType, err)
	})

	t.Run("Verify with invalid key type", func(t *testing.T) {
		invalidKey := "not-a-public-key"
		err := method.Verify("test-string", []byte("signature"), invalidKey)
		require.Error(t, err)
		require.Equal(t, jwt.ErrInvalidKeyType, err)
	})

	t.Run("Verify with invalid signature format", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		invalidSignature := []byte("invalid-signature-format")
		err = method.Verify("test-string", invalidSignature, &privateKey.PublicKey)
		require.Error(t, err)
	})

	t.Run("Verify with wrong signature", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		// Sign one string
		signature, err := method.Sign("original-string", privateKey)
		require.NoError(t, err)

		// Try to verify with a different string
		err = method.Verify("different-string", signature, &privateKey.PublicKey)
		require.Error(t, err)
		require.Equal(t, jwt.ErrSignatureInvalid, err)
	})

	t.Run("Sign and verify success", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
		require.NoError(t, err)

		testString := "test-signing-string"
		signature, err := method.Sign(testString, privateKey)
		require.NoError(t, err)
		require.NotEmpty(t, signature)

		err = method.Verify(testString, signature, &privateKey.PublicKey)
		require.NoError(t, err)
	})
}
