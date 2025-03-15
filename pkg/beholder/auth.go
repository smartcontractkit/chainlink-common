package beholder

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
)

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion = "1"

// BuildAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
func BuildAuthHeaders(ed25519Signer crypto.Signer) (map[string]string, error) {
	pubKey := ed25519Signer.Public().(ed25519.PublicKey)
	messageBytes := pubKey
	signature, err := ed25519Signer.Sign(rand.Reader, messageBytes, crypto.Hash(0))
	if err != nil {
		return nil, fmt.Errorf("ed25519: failed to sign message: %w", err)
	}
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}, nil
}
