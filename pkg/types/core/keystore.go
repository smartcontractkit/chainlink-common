package core

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type Keystore interface {
	Accounts(ctx context.Context) (accounts []string, err error)
	// Sign returns data signed by account.
	// nil data can be used as a no-op to check for account existence.
	Sign(ctx context.Context, account string, data []byte) (signed []byte, err error)
}

var _ crypto.Signer = &Ed25519Signer{}

type SignFn func(ctx context.Context, account string, data []byte) (signed []byte, err error)

// Ed25519Signer implements crypto.Signer and services.StartClose.
type Ed25519Signer struct {
	stopCh  services.StopChan
	account string
	pubKey  crypto.PublicKey
	signFn  func(ctx context.Context, account string, data []byte) (signed []byte, err error)
}

// NewEd25519Signer returns a new Ed25519Signer backed by signFn, which is usually a Keystore.Sign method.
func NewEd25519Signer(account string, signFn SignFn) (*Ed25519Signer, error) {
	account = strings.TrimPrefix(account, "0x")
	b, err := hex.DecodeString(account)
	if err != nil {
		return nil, fmt.Errorf("failed to decode account as hex: %w", err)
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid ed25519 public key size: %d", len(b))
	}
	return &Ed25519Signer{
		stopCh:  make(services.StopChan),
		account: account,
		pubKey:  ed25519.PublicKey(b),
		signFn:  signFn,
	}, nil
}

func (s *Ed25519Signer) Start(ctx context.Context) error { return nil }

func (s *Ed25519Signer) Close() error {
	close(s.stopCh)
	return nil
}

func (s *Ed25519Signer) Public() crypto.PublicKey { return s.pubKey }

func (s *Ed25519Signer) Sign(r io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	ctx, cancel := s.stopCh.NewCtx()
	defer cancel()
	if r != rand.Reader {
		return nil, fmt.Errorf("invalid reader: only crypto/rand.Reader is supported")
	}
	if opts != crypto.Hash(0) { // x509.PureEd25519
		return nil, fmt.Errorf("invalid opts, only crypto.Hash(0) is supported: %v", opts)
	}
	return s.signFn(ctx, s.account, digest)
}

var P2PAccountKey = "P2P_SIGNER"

// singleAccountSigner implements Keystore for a single account.
type singleAccountSigner struct {
	account string
	signer  crypto.Signer
}

var _ Keystore = &singleAccountSigner{}

func NewSingleAccountSigner(account string, signer crypto.Signer) (*singleAccountSigner, error) {
	return &singleAccountSigner{account: account, signer: signer}, nil
}
func (c *singleAccountSigner) Accounts(ctx context.Context) (accounts []string, err error) {
	return []string{c.account}, nil
}
func (c *singleAccountSigner) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	if c.account == account {
		return c.signer.Sign(rand.Reader, data, crypto.Hash(0))
	}
	return nil, fmt.Errorf("account not found: %s", account)
}
