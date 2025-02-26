package core

import (
	"context"
	"crypto"
	"io"
)

type Keystore interface {
	Accounts(ctx context.Context) (accounts []string, err error)
	// Sign returns data signed by account.
	// nil data can be used as a no-op to check for account existence.
	Sign(ctx context.Context, account string, data []byte) (signed []byte, err error)
}

var _ crypto.Signer = &CryptoSigner{}

type CryptoSigner struct {
	ctx     func() context.Context
	account string
	signFn  func(ctx context.Context, account string, data []byte) (signed []byte, err error)
}

// NewCryptoSigner returns a new crypto.Signer backed by signFn, which is usually a Keystore.Sign method.
func NewCryptoSigner(ctx context.Context, account string, signFn func(ctx context.Context, account string, data []byte) (signed []byte, err error)) *CryptoSigner {
	return &CryptoSigner{
		ctx:     func() context.Context { return ctx },
		account: account,
		signFn:  signFn,
	}
}

func (c *CryptoSigner) Public() crypto.PublicKey { return c.account }

func (c *CryptoSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	//TODO sanity check rand & opts?
	return c.signFn(c.ctx(), c.account, digest)
}
