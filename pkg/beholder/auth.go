package beholder

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
)

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion = "1"

type HeaderProvider interface {
	Headers(ctx context.Context) map[string]string
}

type staticAuthHeaderProvider struct {
	headers map[string]string
}

func (p *staticAuthHeaderProvider) Headers(_ context.Context) map[string]string {
	return p.headers
}

func NewStaticAuthHeaderProvider(headers map[string]string) HeaderProvider {
	return &staticAuthHeaderProvider{headers: headers}
}

// BuildAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
// Deprecated: use NewAuthHeaders
func BuildAuthHeaders(privKey ed25519.PrivateKey) map[string]string {
	messageBytes := privKey.Public().(ed25519.PublicKey)
	signature := ed25519.Sign(privKey, messageBytes)
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}
}

// NewAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
func NewAuthHeaders(ed25519Signer crypto.Signer) (map[string]string, error) {
	messageBytes := ed25519Signer.Public().(ed25519.PublicKey)
	signature, err := ed25519Signer.Sign(rand.Reader, messageBytes, crypto.Hash(0))
	if err != nil {
		return nil, fmt.Errorf("ed25519: failed to sign message: %w", err)
	}
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}, nil
}
