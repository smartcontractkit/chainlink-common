package beholder

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/credentials"
)

var (
	// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
	authHeaderKey = "X-Beholder-Node-Auth-Token"
	// authHeaderVersion is the version of the auth header format
	authHeaderVersion1 = "1"
	authHeaderVersion2 = "2"
)

const DefaultAuthHeaderTTL = 1 * time.Minute

type AuthHeaderProvider interface {
	Credentials() credentials.PerRPCCredentials
}

// authHeaderPerRPCredentials is a PerRPCCredentials implementation that provides the auth headers
type authHeaderPerRPCredentials struct {
	privKey                  ed25519.PrivateKey
	lastUpdated              time.Time
	headerTTL                time.Duration
	requireTransportSecurity bool
	headers                  map[string]string
	version                  string
	mu                       sync.Mutex
}

// AuthHeaderProviderConfig configures AuthHeaderProvider
type AuthHeaderProviderConfig struct {
	HeaderTTL                time.Duration
	Version                  string
	RequireTransportSecurity bool
}

func NewAuthHeaderProvider(privKey ed25519.PrivateKey, config *AuthHeaderProviderConfig) AuthHeaderProvider {
	if config == nil {
		config = &AuthHeaderProviderConfig{}
	}
	if config.HeaderTTL <= 0 {
		config.HeaderTTL = DefaultAuthHeaderTTL
	}
	if config.Version == "" {
		config.Version = authHeaderVersion2
	}

	creds := &authHeaderPerRPCredentials{
		privKey:                  privKey,
		headerTTL:                config.HeaderTTL,
		version:                  config.Version,
		requireTransportSecurity: config.RequireTransportSecurity,
	}
	// Initialize the headers ~ lastUpdated is 0 so the headers are generated on the first call
	creds.refresh()
	return creds
}

func (a *authHeaderPerRPCredentials) Credentials() credentials.PerRPCCredentials {
	return a
}

func (a *authHeaderPerRPCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return a.getHeaders(), nil
}

func (a *authHeaderPerRPCredentials) RequireTransportSecurity() bool {
	return a.requireTransportSecurity
}

// getHeaders returns the auth headers, refreshing them if they are expired
func (a *authHeaderPerRPCredentials) getHeaders() map[string]string {
	if time.Since(a.lastUpdated) > a.headerTTL {
		a.refresh()
	}
	return a.headers
}

// refresh creates a new signed auth header token and sets the lastUpdated time to now
func (a *authHeaderPerRPCredentials) refresh() {
	a.mu.Lock()
	defer a.mu.Unlock()

	timeNow := time.Now()

	switch a.version {
	// refresh doesn't actually do anything for version 1 since we are only signing the public key
	// this for backwards compatibility and smooth transition to version 2
	case authHeaderVersion1:
		a.headers = BuildAuthHeaders(a.privKey)
	case authHeaderVersion2:
		a.headers = buildAuthHeadersV2(a.privKey, &AuthHeaderConfig{timestamp: timeNow.UnixMilli()})
	default:
		a.headers = buildAuthHeadersV2(a.privKey, &AuthHeaderConfig{timestamp: timeNow.UnixMilli()})
	}
	// Set the lastUpdated time to now
	a.lastUpdated = timeNow
}

// AuthHeaderConfig configures buildAuthHeadersV2
type AuthHeaderConfig struct {
	timestamp int64
	version   string
}

// BuildAuthHeaders creates the auth headers to be included on requests.
// There are two formats for the header. Version `1` is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
// and <signature_hex> is the signature of the public key.
func BuildAuthHeaders(privKey ed25519.PrivateKey) map[string]string {
	pubKey := privKey.Public().(ed25519.PublicKey)
	messageBytes := pubKey
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%x", authHeaderVersion1, messageBytes, signature)}
}

// buildAuthHeadersV2 creates the auth headers to be included on requests.
// Version `2` is:
//
// <version>:<public_key_hex>:<timestamp>:<signature_hex>
//
// where the concatenated byte value of <public_key_hex> & <timestamp> is what's being signed
func buildAuthHeadersV2(privKey ed25519.PrivateKey, config *AuthHeaderConfig) map[string]string {
	if config == nil {
		config = &AuthHeaderConfig{}
	}
	if config.version == "" {
		config.version = authHeaderVersion2
	}
	// If timestamp is negative or 0, set it to current timestamp.
	// negative values cause overflow on conversion to uint64
	if config.timestamp <= 0 {
		config.timestamp = time.Now().UnixMilli()
	}

	pubKey := privKey.Public().(ed25519.PublicKey)

	timestampUnixMsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampUnixMsBytes, uint64(config.timestamp))

	messageBytes := append(pubKey, timestampUnixMsBytes...)
	signature := ed25519.Sign(privKey, messageBytes)

	return map[string]string{authHeaderKey: fmt.Sprintf("%s:%x:%d:%x", config.version, pubKey, config.timestamp, signature)}
}
