package beholder

import (
	"context"
	"fmt"
	"sync"
)

type LazySigner interface {
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SetSigner(signer Signer)
	HasSigner() bool
}

// lazyKeystoreSigner is a thread-safe wrapper that allows the keystore
// to be set after the signer is created. This enables beholder to start
// with rotating auth configured, but the actual keystore can be injected later.
type lazySigner struct {
	mu sync.RWMutex
	Signer
}

func NewLazySigner() LazySigner {
	return &lazySigner{mu: sync.RWMutex{}, Signer: nil}
}

// Sign implements the beholder.Signer interface
func (l *lazySigner) Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.Signer == nil {
		return nil, fmt.Errorf("keystore not yet available for signing")
	}

	return l.Signer.Sign(ctx, keyID, data)
}

// SetKeystore updates the underlying keystore. This is thread-safe and can be
// called at any time, even after beholder has been initialized.
func (l *lazySigner) SetSigner(signer Signer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Signer = signer
}

// HasKeystore returns true if a keystore has been set
func (l *lazySigner) HasSigner() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Signer != nil
}
