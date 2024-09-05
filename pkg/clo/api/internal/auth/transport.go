package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

// Transport is an http.RoundTripper that makes authenticated requests,
// wrapping a base RoundTripper and adding a session token header with
// the supplied token.
type transport struct {
	// token is the token used in outgoing requests' session token headers.
	token Token

	// Base is the base RoundTripper used to make HTTP requests.
	base http.RoundTripper
}

// RoundTrip authenticates the request with a session token from transport.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.token.Get()
	if err == nil {
		req.Header.Set("X-Session-Token", token)
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %w", err)
	}

	// Clear the cached token if forbidden
	if resp.StatusCode == http.StatusForbidden {
		if err = t.token.Delete(); err != nil {
			return nil, fmt.Errorf("could not delete cached auth token: %w", err)
		}
		return nil, errors.New("forbidden")
	}
	return resp, err
}

// NewHttpClient creates a *http.Client from the provided token.
func newHttpClient(t Token) *http.Client {
	return &http.Client{
		Transport: &transport{
			token: t,
			base:  http.DefaultTransport,
		},
	}
}

// NewGqlClient creates a graphql.Client to make authenticated requests to
// the provided endpoint using the provided session token.
func NewGqlClient(endpoint string, t Token) graphql.Client {
	// Set up an autheticated http client using the token.
	httpClient := newHttpClient(t)

	// Return a graphQL client that wraps the authenticated http client.
	return graphql.NewClient(endpoint, httpClient)
}
