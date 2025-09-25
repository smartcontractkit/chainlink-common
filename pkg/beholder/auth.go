package beholder

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// authHeaderKey is the name of the header that the node authenticator will use to send the auth token
var authHeaderKey = "X-Beholder-Node-Auth-Token"

// authHeaderVersion is the version of the auth header format
var authHeaderVersion = "1"
var authHeaderV2 = "2"

type HeaderProvider interface {
	Headers(ctx context.Context) (map[string]string, error)
}

type Signer interface {
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
}

type staticAuthHeaderProvider struct {
	headers map[string]string
}

func (p *staticAuthHeaderProvider) Headers(_ context.Context) (map[string]string, error) {
	return p.headers, nil
}

func NewStaticAuthHeaderProvider(headers map[string]string) HeaderProvider {
	return &staticAuthHeaderProvider{headers: headers}
}

type rotatingAuthHeaderProvider struct {
	csaPubKey   ed25519.PublicKey
	signer      Signer
	headers     map[string]string
	ttl         time.Duration
	lastUpdated time.Time
}

func NewRotatingAuthHeaderProvider(csaPubKey ed25519.PublicKey, signer Signer, ttl time.Duration) HeaderProvider {
	return &rotatingAuthHeaderProvider{
		csaPubKey:   csaPubKey,
		signer:      signer,
		ttl:         ttl,
		log:         log,
		headers:     make(map[string]string),
		lastUpdated: time.Unix(0, 0),
	}
}

func (r *rotatingAuthHeaderProvider) Headers(ctx context.Context) (map[string]string, error) {
	if time.Since(r.lastUpdated) > r.ttl {
		// Append timestamp bytes to the public key bytes
		timestamp := time.Now().UnixMilli()
		timestampBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
		messageBytes := append(r.csaPubKey, timestampBytes...)
		// Sign(public key bytes + timestamp bytes)
		signature, err := r.signer.Sign(ctx, r.csaPubKey, messageBytes)
		if err != nil {
			return nil, fmt.Errorf("beholder: failed to sign auth header: %w", err)
		}

		r.headers[authHeaderKey] = fmt.Sprintf("%s:%x:%d:%x", authHeaderV2, r.csaPubKey, timestamp, signature)
		r.lastUpdated = time.Now()
	}
	return r.headers, nil
}

// BuildAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
// Deprecated: use NewAuthHeaders
func BuildAuthHeaders(privKey ed25519.PrivateKey) map[string]string {
	messageBytes := privKey.Public().(ed25519.PublicKey)
	signature := ed25519.Sign(privKey, messageBytes)
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}
}

// NewAuthHeaders creates the auth header value to be included on requests.
// The current format for the header is:
//
// <version>:<public_key_hex>:<signature_hex>
//
// where the byte value of <public_key_hex> is what's being signed
func NewAuthHeaders(ed25519Signer crypto.Signer) (map[string]string, error) {
	messageBytes := ed25519Signer.Public().(ed25519.PublicKey)
	signature, err := ed25519Signer.Sign(rand.Reader, messageBytes, crypto.Hash(0))
	if err != nil {
		return nil, fmt.Errorf("ed25519: failed to sign message: %w", err)
	}
	headerValue := fmt.Sprintf("%s:%x:%x", authHeaderVersion, messageBytes, signature)

	return map[string]string{authHeaderKey: headerValue}, nil
}
