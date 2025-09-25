package chipingress

import (
	"context"
	"testing"
)

func TestBasicAuth(t *testing.T) {

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	t.Run("AuthHeaderIsReturnedCorrectly", func(t *testing.T) {
		credentials := newBasicAuthCredentials("user", "pass", false)
		meta, err := credentials.GetRequestMetadata(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := "Basic dXNlcjpwYXNz"
		if meta["authorization"] != expected {
			t.Errorf("got %q, want %q", meta["authorization"], expected)
		}
	})

	t.Run("TransportSecurityIsRequired", func(t *testing.T) {
		credentials := newBasicAuthCredentials("user", "pass", true)
		if !credentials.RequireTransportSecurity() {
			t.Errorf("expected transport security to be required")
		}
	})

	t.Run("TransportSecurityIsNotRequired", func(t *testing.T) {
		credentials := newBasicAuthCredentials("user", "pass", false)
		if credentials.RequireTransportSecurity() {
			t.Errorf("expected transport security not to be required")
		}
	})

	t.Run("HandlesEmptyCredentials", func(t *testing.T) {
		credentials := newBasicAuthCredentials("", "", false)
		meta, err := credentials.GetRequestMetadata(t.Context())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := "Basic Og=="
		if meta["authorization"] != expected {
			t.Errorf("got %q, want %q", meta["authorization"], expected)
		}
	})
}

type testHeaderProvider struct {
	headers map[string]string
}

func (p *testHeaderProvider) Headers(ctx context.Context) map[string]string {
	return p.headers
}

func TestTokenAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	t.Run("AuthTokenIsReturnedCorrectly", func(t *testing.T) {
		provider := &testHeaderProvider{headers: map[string]string{"X-Auth-Token": "my-token"}}
		credentials := newTokenAuthCredentials(provider, false)
		meta, err := credentials.GetRequestMetadata(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if meta["X-Auth-Token"] != "my-token" {
			t.Errorf("got %q, want %q", meta["X-Auth-Token"], "my-token")
		}
	})

	t.Run("TransportSecurityIsRequired", func(t *testing.T) {
		provider := &testHeaderProvider{headers: map[string]string{"X-Auth-Token": "my-token"}}
		credentials := newTokenAuthCredentials(provider, true)
		if !credentials.RequireTransportSecurity() {
			t.Errorf("expected transport security to be required")
		}
	})

	t.Run("TransportSecurityIsNotRequired", func(t *testing.T) {
		provider := &testHeaderProvider{headers: map[string]string{"X-Auth-Token": "my-token"}}
		credentials := newTokenAuthCredentials(provider, false)
		if credentials.RequireTransportSecurity() {
			t.Errorf("expected transport security not to be required")
		}
	})
}
