package beholder

import (
	"context"
	"encoding/hex"
	"fmt"
)

// NodeAuthenticator implements PerRPCCredentials interface from google.golang.org/grpc/credentials
type NodeAuthenticator struct {
	staticAuthHeaders        map[string]string
	requireTransportSecurity bool
}

func NewNodeAuthenticator(config Config) (*NodeAuthenticator, error) {
	if err := validateNodeAuthConfig(config); err != nil {
		return nil, err
	}

	return &NodeAuthenticator{
		staticAuthHeaders:        deriveAuthHeaders(config),
		requireTransportSecurity: !config.InsecureConnection,
	}, nil
}

func (na *NodeAuthenticator) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	authHeaders := make(map[string]string)
	for k, v := range na.staticAuthHeaders {
		authHeaders[k] = v
	}
	return authHeaders, nil
}

func (na *NodeAuthenticator) RequireTransportSecurity() bool {
	return na.requireTransportSecurity
}

func validateNodeAuthConfig(config Config) error {
	if config.CSAAuthEnabled {
		if len(config.CSAPublicKey) == 0 {
			return fmt.Errorf("CSA auth is enabled but no CSA public key was provided")
		}
		if config.CSASigner == nil {
			return fmt.Errorf("CSA auth is enabled but no CSA signer was provided")
		}
	}
	return nil
}

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion = "1"

// deriveAuthHeaders creates the auth header map to be included on requests
func deriveAuthHeaders(config Config) map[string]string {
	authHeaders := make(map[string]string)

	if !config.CSAAuthEnabled {
		return authHeaders
	}

	authHeaders[authHeaderKey] = deriveAuthHeaderValue(config)

	return authHeaders
}

// deriveAuthHeaderValue creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<csa_public_key_hex>:<signature_hex>
//
// where the byte value of <csa_public_key_hex> is what's being signed
func deriveAuthHeaderValue(config Config) string {
	messageBytes := config.CSAPublicKey

	signature := config.CSASigner(messageBytes)

	return fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)
}
