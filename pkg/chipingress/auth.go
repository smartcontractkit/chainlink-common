package chipingress

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc/credentials"
)

var _ credentials.PerRPCCredentials = basicAuthCredentials{}
var _ credentials.PerRPCCredentials = tokenAuthCredentials{}

type HeaderProvider interface {
	Headers(ctx context.Context) (map[string]string, error)
}

type headerProviderFunc func(ctx context.Context) (map[string]string, error)

func (f headerProviderFunc) Headers(ctx context.Context) (map[string]string, error) {
	return f(ctx)
}

// Basic-Auth authentication for Chip Ingress
type basicAuthCredentials struct {
	authHeader map[string]string
	requireTLS bool
}

func (b basicAuthCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {

	return b.authHeader, nil
}

func (b basicAuthCredentials) RequireTransportSecurity() bool {
	return b.requireTLS
}

func newBasicAuthCredentials(userName, password string, requireTLS bool) basicAuthCredentials {

	auth := userName + ":" + password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	header := map[string]string{
		"authorization": "Basic " + encoded,
	}

	return basicAuthCredentials{header, requireTLS}
}

// CSA-Key based authentication for Chip Ingress
type tokenAuthCredentials struct {
	authTokenProvider HeaderProvider
	requireTLS        bool
}

// implement PerRPCCredentials interface
func (c tokenAuthCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	if c.authTokenProvider == nil {
		return nil, nil
	}
	return c.authTokenProvider.Headers(ctx)
}

func (c tokenAuthCredentials) RequireTransportSecurity() bool {
	return c.requireTLS
}

func newTokenAuthCredentials(provider HeaderProvider, requireTLS bool) tokenAuthCredentials {
	return tokenAuthCredentials{
		authTokenProvider: provider,
		requireTLS:        requireTLS,
	}
}
