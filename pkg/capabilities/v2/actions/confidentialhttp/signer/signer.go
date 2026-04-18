// Package signer implements request-signing for the confidentialHTTP capability.
//
// Each Signer takes a ready-to-send *http.Request, the values of the Vault-DON
// secrets that the workflow declared, and mutates the request in place to
// attach authentication (header, signature, etc.).
//
// Signers are pure logic: they never talk to Vault DON. Secret resolution
// happens upstream in confidential-compute's framework executor before Sign
// is invoked. The signer receives the already-decrypted values via the
// secrets map.
//
// The OAuth2 signers are the one exception: they make outbound HTTP calls to
// the IdP's token endpoint to exchange credentials for an access token. That
// still does not involve Vault DON — the credentials used for the exchange
// come from the secrets map.
package signer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// Signer mutates httpReq in place to attach authentication material.
// secrets holds secret values keyed by the name the workflow declared in
// ConfidentialHTTPRequest.vault_don_secrets.
type Signer interface {
	Sign(ctx context.Context, httpReq *http.Request, secrets map[string]string) error
}

// Builder constructs a Signer from an AuthConfig. A single Builder instance
// may be reused across many requests; it owns shared state such as the OAuth2
// token cache.
type Builder struct {
	httpClient *http.Client
	oauthCache *oauth2Cache
}

// NewBuilder returns a Builder configured with the supplied http.Client.
// The client is used by OAuth2 signers to reach the IdP's token endpoint.
// If nil, http.DefaultClient is used.
func NewBuilder(client *http.Client) *Builder {
	if client == nil {
		client = http.DefaultClient
	}
	return &Builder{
		httpClient: client,
		oauthCache: newOAuth2Cache(),
	}
}

// Build selects the appropriate Signer for the given AuthConfig.
// Returns (nil, nil) when auth is nil — callers should treat that as "no
// signing, send the request as-is".
func (b *Builder) Build(auth *confhttppb.AuthConfig) (Signer, error) {
	if auth == nil {
		return nil, nil
	}
	switch m := auth.GetMethod().(type) {
	case *confhttppb.AuthConfig_ApiKey:
		return newAPIKeySigner(m.ApiKey)
	case *confhttppb.AuthConfig_Basic:
		return newBasicSigner(m.Basic)
	case *confhttppb.AuthConfig_Bearer:
		return newBearerSigner(m.Bearer)
	case *confhttppb.AuthConfig_Hmac:
		return newHmacSigner(m.Hmac)
	case *confhttppb.AuthConfig_Oauth2:
		return newOAuth2Signer(m.Oauth2, b.httpClient, b.oauthCache)
	case nil:
		return nil, errors.New("auth method not set")
	default:
		return nil, fmt.Errorf("unsupported auth method %T", m)
	}
}

// resolveSecret returns the utf8 secret value for name or an error if the
// workflow did not provide it. The confidentialhttp capability validator
// enforces that every name referenced by AuthConfig appears in
// vault_don_secrets, so missing values here should be treated as an
// internal error in most cases.
func resolveSecret(secrets map[string]string, name string) (string, error) {
	if name == "" {
		return "", ErrSecretNameEmpty
	}
	v, ok := secrets[name]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrSecretNotProvided, name)
	}
	if v == "" {
		return "", fmt.Errorf("%w: %q", ErrSecretEmpty, name)
	}
	return v, nil
}
