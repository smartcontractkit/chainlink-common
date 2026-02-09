package core

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

type AccountsFilter struct {
	// WithDeterminism set ensures that the accounts list is returned in a fixed
	// order across all node restarts.
	WithDeterminism bool
}

// Implementations of this interface should embed the UnimplementedKeystore struct,
// as to ensure forward compatibility with changes to the Keystore interface.
type Keystore interface {
	Accounts(ctx context.Context) (accounts []string, err error)
	ListAccounts(context.Context, *AccountsFilter) ([]string, error)
	// Sign returns data signed by account.
	// nil data can be used as a no-op to check for account existence.
	Sign(ctx context.Context, account string, data []byte) (signed []byte, err error)

	Decrypt(ctx context.Context, account string, encrypted []byte) (decrypted []byte, err error)
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

// Notice: when using the `StandardCapabilityAccount`, signing payloads must be prefixed using the
// `peeridhelper.MakePeerIDSignatureDomainSeparatedPayload` function.
var StandardCapabilityAccount = "STANDARD_CAPABILITY_ACCOUNT"

// singleAccountSigner implements Keystore for a single account.
type singleAccountSigner struct {
	UnimplementedKeystore
	account *string
	signer  crypto.Signer
}

var _ Keystore = &singleAccountSigner{}

func NewSingleAccountSigner(account *string, signer crypto.Signer) (*singleAccountSigner, error) {
	return &singleAccountSigner{account: account, signer: signer}, nil
}

func (c *singleAccountSigner) Accounts(ctx context.Context) (accounts []string, err error) {
	if c.account == nil {
		return nil, fmt.Errorf("account is nil")
	}

	return []string{*c.account}, nil
}

func (c *singleAccountSigner) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	if c.account != nil && *c.account == account {
		return c.signer.Sign(rand.Reader, data, crypto.Hash(0))
	}
	return nil, fmt.Errorf("account not found: %s", account)
}

func (c *singleAccountSigner) Decrypt(ctx context.Context, account string, encrypted []byte) (decrypted []byte, err error) {
	return nil, fmt.Errorf("decrypt not supported for single account signer")
}

type Decrypter interface {
	Decrypt(encrypted []byte) (decrypted []byte, err error)
}

// signerDecrypter implements Keystore for a single sign account and decrypt account.
type signerDecrypter struct {
	UnimplementedKeystore
	account   string
	signer    crypto.Signer
	decrypter Decrypter
}

var _ Keystore = &signerDecrypter{}

func NewSignerDecrypter(account string, signer crypto.Signer, decrypter Decrypter) (*signerDecrypter, error) {
	return &signerDecrypter{account: account, signer: signer, decrypter: decrypter}, nil
}

func (c *signerDecrypter) Accounts(ctx context.Context) ([]string, error) {
	return []string{c.account}, nil
}

var genericPrefix = peeridhelper.MakePeerIDSignatureDomainSeparatedPayload("", []byte{})

func (c *signerDecrypter) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	if c.account != account {
		return nil, fmt.Errorf("account not found: %s", account)
	}
	if c.signer == nil {
		return nil, fmt.Errorf("signer is nil")
	}

	// For the `StandardCapabilityAccount`, assert that the user is passing in correctly domain separated data.
	// The first 97 bytes of any domain separated payload will match the generic prefix.
	// Implicitly, this also requires the message length to be <= 1024 bytes.
	if account == StandardCapabilityAccount {
		if !bytes.HasPrefix(data, genericPrefix[:97]) {
			return nil, fmt.Errorf("data does not have expected prefix")
		}
	}

	return c.signer.Sign(rand.Reader, data, crypto.Hash(0))
}

func (c *signerDecrypter) Decrypt(ctx context.Context, account string, encrypted []byte) (decrypted []byte, err error) {
	if c.account != account {
		return nil, fmt.Errorf("account not found: %s", account)
	}
	if c.decrypter == nil {
		return nil, fmt.Errorf("decrypter is nil")
	}
	return c.decrypter.Decrypt(encrypted)
}

var _ Keystore = &UnimplementedKeystore{}

type UnimplementedKeystore struct{}

func (u *UnimplementedKeystore) Accounts(ctx context.Context) (accounts []string, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method Accounts not implemented")
}

func (u *UnimplementedKeystore) ListAccounts(ctx context.Context, filter *AccountsFilter) ([]string, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}

func (u *UnimplementedKeystore) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sign not implemented")
}

func (u *UnimplementedKeystore) Decrypt(ctx context.Context, account string, encrypted []byte) (decrypted []byte, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method Decrypt not implemented")
}
