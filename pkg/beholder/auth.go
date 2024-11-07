package beholder

import (
	"crypto/ed25519"
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
func BuildAuthHeaders(privKey ed25519.PrivateKey, pubKey ed25519.PublicKey) map[string]string {
	messageBytes := pubKey
	signature := ed25519.Sign(privKey, messageBytes)
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}
}
