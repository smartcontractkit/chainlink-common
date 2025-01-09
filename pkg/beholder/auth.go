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

const (
	// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
	authHeaderKey = "X-Beholder-Node-Auth-Token"
	// authHeaderVersion is the version of the auth header format
	authHeaderVersion1 = "1"
	authHeaderVersion2 = "2"
	// defaultAuthHeaderTTL is the default time before the auth header is refreshed
	defaultAuthHeaderTTL = 1 * time.Minute
)

type AuthHeaderProvider interface {
	// Credentials returns the PerRPCCredentials implementation
	Credentials() credentials.PerRPCCredentials
	// SetRequireTransportSecurity sets the value of requireTransportSecurity
	SetRequireTransportSecurity(bool)
}

// AuthHeaderProviderConfig configures AuthHeaderProvider
type AuthHeaderProviderConfig struct {
	HeaderTTL                time.Duration
	Version                  string
	RequireTransportSecurity bool
}

// authHeaderPerRPCredentials is a PerRPCCredentials implementation that provides the auth headers
type authHeaderPerRPCCredentials struct {
	privKey                  ed25519.PrivateKey
	lastUpdated              time.Time
	headerTTL                time.Duration
	refreshFunc              func() (map[string]string, error)
	requireTransportSecurity bool
	headers                  map[string]string
	version                  string
	mu                       sync.Mutex
}

func (config AuthHeaderProviderConfig) New(refreshFunc func() (map[string]string, error)) AuthHeaderProvider {
	if config.HeaderTTL <= 0 {
		config.HeaderTTL = defaultAuthHeaderTTL
	}
	if config.Version == "" {
		config.Version = authHeaderVersion2
	}
	if refreshFunc == nil {
		refreshFunc = func() (map[string]string, error) { return make(map[string]string), nil }
	}

	creds := &authHeaderPerRPCCredentials{
		headerTTL:                config.HeaderTTL,
		version:                  config.Version,
		requireTransportSecurity: config.RequireTransportSecurity,
		refreshFunc:              refreshFunc,
	}
	return creds
}

func NewAuthHeaderProvider(refreshFunc func() (map[string]string, error)) AuthHeaderProvider {
	return AuthHeaderProviderConfig{}.New(refreshFunc)
}

func (a *authHeaderPerRPCCredentials) Credentials() credentials.PerRPCCredentials {
	return a
}

func (a *authHeaderPerRPCCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return a.getHeaders()
}

func (a *authHeaderPerRPCCredentials) RequireTransportSecurity() bool {
	return a.requireTransportSecurity
}

// SetRequireTransportSecurity sets the value of requireTransportSecurity
// This is to safeguard against inconsistent values between the PerRPCCredentials and the AuthHeaderProvider
func (a *authHeaderPerRPCCredentials) SetRequireTransportSecurity(newValue bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.requireTransportSecurity = newValue
}

// refresh updates the auth headers with configurable refresh mechanism
// This allows for LOOPPs to refresh the auth headers by calling gRPC and core node to refresh the auth headers directly with its private key
func (a *authHeaderPerRPCCredentials) refresh() (map[string]string, error) {
	return a.refreshFunc()
}

// getHeaders returns the auth headers, refreshing them if they are expired
func (a *authHeaderPerRPCCredentials) getHeaders() (map[string]string, error) {
	if time.Since(a.lastUpdated) > a.headerTTL {

		a.mu.Lock()
		defer a.mu.Unlock()

		headers, err := a.refresh()
		if err != nil {
			return nil, err
		}

		a.headers = headers
		a.lastUpdated = time.Now()
	}

	return a.headers, nil
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

// BuildAuthHeadersV2 creates the auth headers to be included on requests.
// Version `2` is:
//
// <version>:<public_key_hex>:<timestamp>:<signature_hex>
//
// where the concatenated byte value of <public_key_hex> & <timestamp> is what's being signed
func BuildAuthHeadersV2(privKey ed25519.PrivateKey, config *AuthHeaderConfig) map[string]string {
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
