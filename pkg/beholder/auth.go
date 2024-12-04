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
var authHeaderVersion = "1"

// BuildAuthHeadersOpt are used to modify the behavior of the BuildAuthHeaders function
type BuildAuthHeadersOpt func(*authHeaderConfig)

type authHeaderConfig struct {
	privKey   ed25519.PrivateKey
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
func BuildAuthHeaders(privKey ed25519.PrivateKey, opts ...BuildAuthHeadersOpt) map[string]string {
	// Defaults
	cfg := &authHeaderConfig{
		privKey:   privKey,
		timestamp: time.Now().UnixMilli(),
		version:   authHeaderVersion,
	}
	// Apply config options
	for _, opt := range opts {
		opt(cfg)
	}
	// Use the new version of the auth header with timestamps
	if cfg.version == "2" {
		return buildAuthHeadersV2(cfg.privKey, cfg.timestamp, cfg.version)
	}
	// Use the old version of the auth header as default
	return buildAuthHeaderV1(cfg.privKey)
}

func buildAuthHeaderV1(privKey ed25519.PrivateKey) map[string]string {
	pubKey := privKey.Public().(ed25519.PublicKey)
	messageBytes := pubKey
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)}
}

func buildAuthHeadersV2(privKey ed25519.PrivateKey, timestamp int64, version string) map[string]string {
	pubKey := privKey.Public().(ed25519.PublicKey)

	timestampUnixMsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampUnixMsBytes, uint64(timestamp))

	messageBytes := append(pubKey, timestampUnixMsBytes...)
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%d:%x", version, pubKey, timestamp, signature)}
}

// WithAuthHeaderTimestamp is an option to set the timestamp to be used in auth headers ~ default is time.Now().UnixMilli()
func WithAuthHeaderTimestamp(timestamp int64) BuildAuthHeadersOpt {
	return func(cfg *authHeaderConfig) {
		cfg.timestamp = timestamp
	}
}

// WithAuthHeaderV2 is an option to use version 2 of auth headers that uses timestamps
func WithAuthHeaderV2() BuildAuthHeadersOpt {
	return func(cfg *authHeaderConfig) {
		cfg.version = "2"
	}
}
