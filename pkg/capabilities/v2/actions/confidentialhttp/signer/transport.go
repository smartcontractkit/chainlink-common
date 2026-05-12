package signer

import "net/http"

// TransportConfigurer mutates a *http.Transport before the TCP connection is
// dialled. It is the correct interface for mTLS: TLS configuration must be set
// before the handshake, so it cannot be handled by a Signer (which operates on
// an already-connected request).
//
// Implementations must be safe for concurrent use; a single configurer may be
// called from multiple goroutines.
type TransportConfigurer interface {
	ConfigureTransport(t *http.Transport, secrets map[string]string) error
}
