package beholder

import (
	"errors"
	"fmt"
	"maps"
)

type Authenticator struct {
	headers map[string]string
}

func NewAuthenticator(config Config) (*Authenticator, error) {
	if err := validateAuthConfig(config); err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	if config.AuthenticatorSigner != nil {
		headers = deriveAuthHeaders(config)
	}
	if len(config.AuthenticatorHeaders) != 0 {
		headers = config.AuthenticatorHeaders
	}

	return &Authenticator{headers}, nil
}

func (a *Authenticator) GetHeaders() map[string]string {
	return maps.Clone(a.headers)
}

func validateAuthConfig(c Config) error {
	if c.AuthenticatorSigner != nil && len(c.AuthenticatorHeaders) > 0 {
		return errors.New("cannot configure both authenticator signer and authenticator header value")
	}

	if c.AuthenticatorSigner != nil && len(c.AuthenticatorPublicKey) == 0 {
		return errors.New("authenticator public key must be configured when authenticator signer is configured")
	}

	return nil
}

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion = "1"

// deriveAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
func deriveAuthHeaders(config Config) map[string]string {
	messageBytes := config.AuthenticatorPublicKey
	signature := config.AuthenticatorSigner(messageBytes)
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}
}
