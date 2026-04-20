package signer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// oauth2TokenResponseMaxBytes caps how much of the IdP response we'll read.
// IdPs typically return <4KB; a rogue IdP should not DoS us.
const oauth2TokenResponseMaxBytes = 16 * 1024

// clientAuthMethodBasic and clientAuthMethodBody are the two ways OAuth2
// callers pass client_id/client_secret to the token endpoint.
const (
	clientAuthMethodBasic = "basic_auth"
	clientAuthMethodBody  = "request_body"
)

func newOAuth2Signer(cfg *confhttppb.OAuth2Auth, httpClient *http.Client, cache *oauth2Cache) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("oauth2 auth config is nil")
	}
	switch v := cfg.GetVariant().(type) {
	case *confhttppb.OAuth2Auth_ClientCredentials:
		return newOAuth2ClientCredsSigner(v.ClientCredentials, httpClient, cache)
	case *confhttppb.OAuth2Auth_RefreshToken:
		return newOAuth2RefreshSigner(v.RefreshToken, httpClient, cache)
	case nil:
		return nil, errors.New("oauth2 variant not set")
	default:
		return nil, fmt.Errorf("unsupported oauth2 variant %T", v)
	}
}

// tokenResponse is the minimal shape of a successful OAuth2 token exchange.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// postTokenRequest sends form-encoded data to tokenURL and returns the
// parsed response. Classifies failures into the signer error sentinels so
// the framework can map them to the right caperrors code.
func postTokenRequest(
	ctx context.Context,
	httpClient *http.Client,
	tokenURL string,
	form url.Values,
	basicAuth *struct{ user, pass string },
) (*tokenResponse, error) {
	if !strings.HasPrefix(tokenURL, "https://") {
		return nil, fmt.Errorf("%w: %q", ErrOAuth2TokenURLInvalid, tokenURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuth2TokenEndpointUnreachable, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if basicAuth != nil {
		req.SetBasicAuth(basicAuth.user, basicAuth.pass)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuth2TokenEndpointUnreachable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, oauth2TokenResponseMaxBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", ErrOAuth2TokenResponseInvalid, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Intentionally do NOT include the response body — some IdPs echo
		// back client_id or include the invalid token, which would leak
		// into workflow error messages.
		return nil, fmt.Errorf("%w: status %d", ErrOAuth2TokenEndpointHTTPError, resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuth2TokenResponseInvalid, err)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("%w: access_token missing", ErrOAuth2TokenResponseInvalid)
	}
	return &tr, nil
}
