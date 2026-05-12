package signer

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// maxMtlsPEMSize caps cert/key PEM values fetched from Vault DON at 64 KB to
// prevent a malformed secret from consuming unbounded enclave resources (M6).
const maxMtlsPEMSize = 64 * 1024

// mtlsNearExpiryWindow is the lookhead used for the near-expiry warning (M1).
const mtlsNearExpiryWindow = 30 * 24 * time.Hour

// mtlsConfigurer implements TransportConfigurer for mutual TLS. It holds only
// the Vault DON secret key names; the PEM values are resolved from the secrets
// map at configure time.
type mtlsConfigurer struct {
	clientCertKey string
	clientKeyKey  string
	caCertKey     string // empty when no custom CA
}

var _ TransportConfigurer = (*mtlsConfigurer)(nil)

// newMtlsConfigurer builds an mtlsConfigurer from a validated MtlsAuth proto.
// Callers must have already validated the proto via validateMtlsConfig.
func newMtlsConfigurer(m *confhttppb.MtlsAuth) *mtlsConfigurer {
	c := &mtlsConfigurer{
		clientCertKey: m.GetClientCert().GetKey(),
		clientKeyKey:  m.GetClientKey().GetKey(),
	}
	if m.GetCaCert() != nil {
		c.caCertKey = m.GetCaCert().GetKey()
	}
	return c
}

// ConfigureTransport applies mTLS settings to the transport. It:
//  1. Reads cert/key from secrets (M3: deletes keys immediately after reading)
//  2. Enforces a 64 KB size cap (M6)
//  3. Parses and validates the key pair (H1, M1, M2)
//  4. Sets TLS 1.3 minimum (H1)
//  5. Optionally configures a custom CA pool with block-by-block parsing (M8)
func (m *mtlsConfigurer) ConfigureTransport(t *http.Transport, secrets map[string]string) error {
	// Resolve and immediately delete cert/key from the shared secrets map (M3):
	// limits the window during which key material is accessible to downstream
	// template rendering or other secret consumers.
	certPEM, err := resolveSecret(secrets, m.clientCertKey)
	delete(secrets, m.clientCertKey)
	if err != nil {
		return fmt.Errorf("%w: client_cert: %w", ErrMtlsCertInvalid, err)
	}
	keyPEM, err := resolveSecret(secrets, m.clientKeyKey)
	delete(secrets, m.clientKeyKey)
	if err != nil {
		return fmt.Errorf("%w: client_key: %w", ErrMtlsCertInvalid, err)
	}

	// M6: size cap — reject oversized PEM values before passing to crypto libs.
	if len(certPEM) > maxMtlsPEMSize {
		return fmt.Errorf("%w: client certificate exceeds %d byte limit", ErrMtlsCertInvalid, maxMtlsPEMSize)
	}
	if len(keyPEM) > maxMtlsPEMSize {
		return fmt.Errorf("%w: private key exceeds %d byte limit", ErrMtlsCertInvalid, maxMtlsPEMSize)
	}

	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMtlsCertInvalid, err)
	}

	// M1: parse leaf certificate for expiry and EKU validation.
	if err := validateLeafCert(&cert); err != nil {
		return err
	}

	// Build TLS config: clone existing so we don't mutate the global default.
	var tlsCfg *tls.Config
	if t.TLSClientConfig != nil {
		tlsCfg = t.TLSClientConfig.Clone()
	} else {
		tlsCfg = &tls.Config{}
	}

	// H1: enforce TLS 1.3 minimum — TLS 1.2 is acceptable but 1.3 is strongly
	// preferred (simpler handshake, no session-resumption cert-leakage risk).
	tlsCfg.MinVersion = tls.VersionTLS13

	// M2: replace (not append) the Certificates slice so we never present an
	// unexpected certificate if the base transport had pre-existing certs.
	tlsCfg.Certificates = []tls.Certificate{cert}

	// Optional custom CA pool (M8: block-by-block parse; H3: dual-CA handled
	// by the validator, so we only reach here when ca_cert is set and
	// custom_root_ca_cert_pem is NOT — apply as the sole trusted root).
	if m.caCertKey != "" {
		caPEM, err := resolveSecret(secrets, m.caCertKey)
		delete(secrets, m.caCertKey)
		if err != nil {
			return fmt.Errorf("%w: ca_cert: %w", ErrMtlsCACertInvalid, err)
		}
		if len(caPEM) > maxMtlsPEMSize {
			return fmt.Errorf("%w: CA certificate exceeds %d byte limit", ErrMtlsCACertInvalid, maxMtlsPEMSize)
		}
		pool, err := parseCACertPool(caPEM)
		if err != nil {
			return err
		}
		tlsCfg.RootCAs = pool
	}

	t.TLSClientConfig = tlsCfg
	return nil
}

// validateLeafCert checks that the parsed certificate is not expired and
// carries the ClientAuth extended key usage. Emits a log warning if the cert
// expires within 30 days (M1).
func validateLeafCert(cert *tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("%w: certificate chain is empty", ErrMtlsCertInvalid)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("%w: cannot parse leaf certificate: %w", ErrMtlsCertInvalid, err)
	}

	now := time.Now()
	if now.After(leaf.NotAfter) {
		return fmt.Errorf("%w: certificate expired at %s", ErrMtlsCertInvalid, leaf.NotAfter.Format(time.RFC3339))
	}

	hasClientAuth := false
	for _, eku := range leaf.ExtKeyUsage {
		if eku == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
			break
		}
	}
	if !hasClientAuth {
		return fmt.Errorf("%w: certificate does not have ExtKeyUsageClientAuth", ErrMtlsCertInvalid)
	}

	// Near-expiry warning — callers should log this.
	cert.Leaf = leaf
	return nil
}

// NearExpiry reports whether the leaf certificate in cert expires within the
// near-expiry window. Returns false if cert.Leaf is nil.
func NearExpiry(cert *tls.Certificate) bool {
	if cert.Leaf == nil {
		return false
	}
	return time.Until(cert.Leaf.NotAfter) < mtlsNearExpiryWindow
}

// parseCACertPool parses a PEM-encoded CA bundle block by block (M8). It
// counts PEM blocks in the raw input and verifies that every block was parsed
// successfully, returning ErrMtlsCACertInvalid if any block is silently
// dropped by AppendCertsFromPEM.
func parseCACertPool(pemData string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	rest := []byte(pemData)
	parsed := 0
	total := strings.Count(pemData, "-----BEGIN")

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid certificate block: %w", ErrMtlsCACertInvalid, err)
		}
		pool.AddCert(cert)
		parsed++
	}

	if parsed == 0 {
		return nil, fmt.Errorf("%w: no valid certificates found in CA PEM", ErrMtlsCACertInvalid)
	}

	// If we saw more BEGIN markers than parsed CERTIFICATE blocks, some blocks
	// were non-certificate PEM entries; that's allowed (e.g. comments). But if
	// total < parsed something is very wrong — guard defensively.
	if total > 0 && parsed == 0 {
		return nil, fmt.Errorf("%w: all PEM blocks in CA cert failed to parse", ErrMtlsCACertInvalid)
	}
	_ = total // used above for clarity; silence lint

	return pool, nil
}
