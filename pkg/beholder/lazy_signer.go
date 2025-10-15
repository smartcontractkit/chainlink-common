package beholder

import (
	"context"
	"fmt"
	"sync/atomic"
)

// lazySigner is a thread-safe wrapper that allows the keystore
// to be set after the signer is created. This enables beholder to start
// with rotating auth configured, but the actual keystore can be injected later.
// The zero value is usable.
type lazySigner struct {
	signer atomic.Pointer[Signer]
}

// Sign implements the beholder.Signer interface
func (l *lazySigner) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
	s := l.signer.Load()
	if s == nil {
		return nil, fmt.Errorf("keystore not yet available for signing")
	}
	return (*s).Sign(ctx, keyID, data)
}

// Set updates the underlying keystore. This is thread-safe and can be
// called at any time, even after beholder has been initialized.
func (l *lazySigner) Set(signer Signer) {
	l.signer.Store(&signer)
}

// IsSet returns true if a keystore has been set
func (l *lazySigner) IsSet() bool {
	return l.signer.Load() != nil
}
