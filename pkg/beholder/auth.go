package beholder

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"time"
)

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion1 = "1"
var authHeaderVersion2 = "2"

// AuthHeaderConfig is a configuration struct for the BuildAuthHeadersV2 function
type AuthHeaderConfig struct {
	timestamp int64
	version   string
}

// BuildAuthHeaders creates the auth header value to be included on requests.
// There are two formats for the header. Version `1` is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
// and <signature_hex> is the signature of the public key.
// The version `2` is:
//
// <version>:<public_key_hex>:<timestamp>:<signature_hex>
//
// where the byte value of <public_key_hex> and <timestamp> are what's being signed
func BuildAuthHeaders(privKey ed25519.PrivateKey) map[string]string {
	pubKey := privKey.Public().(ed25519.PublicKey)
	messageBytes := pubKey
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%x", authHeaderVersion1, messageBytes, signature)}
}

// BuildAuthHeadersV2 creates the auth header value to be included on requests.
// See documentation on BuildAuthHeaders for more info. 
func BuildAuthHeadersV2(privKey ed25519.PrivateKey, config *AuthHeaderConfig) map[string]string {
	if config == nil {
		config = defaultAuthHeaderConfig()
	}
	if config.version == "" {
		config.version = authHeaderVersion2
	}
	// If timestamp is not set, use the current time
	if config.timestamp == 0 {
		config.timestamp = time.Now().UnixMilli()
	}
	// If timestamp is negative, set it to 0. negative values cause overflow on conversion to uint64
	// 0 timestamps will be rejected by the server as being too old
	if config.timestamp < 0 {
		config.timestamp = 0
	}

	pubKey := privKey.Public().(ed25519.PublicKey)

	timestampUnixMsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampUnixMsBytes, uint64(config.timestamp))

	messageBytes := append(pubKey, timestampUnixMsBytes...)
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%d:%x", config.version, pubKey, config.timestamp, signature)}
}

func defaultAuthHeaderConfig() *AuthHeaderConfig {
	return &AuthHeaderConfig{
		version:   authHeaderVersion2,
		timestamp: time.Now().UnixMilli(),
	}
}
