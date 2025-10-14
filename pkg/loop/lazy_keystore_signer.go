package loop

import (
	"context"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/types/keystore"
)

type LazyKeystoreSigner interface {
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SetKeystore(ks keystore.Keystore)
	HasKeystore() bool
}

// lazyKeystoreSigner is a thread-safe wrapper that allows the keystore
// to be set after the signer is created. This enables beholder to start
// with rotating auth configured, but the actual keystore can be injected later.
type lazyKeystoreSigner struct {
	mu       sync.RWMutex
	keystore keystore.Keystore
}

func NewLazyKeystoreSigner() LazyKeystoreSigner {
	return &lazyKeystoreSigner{mu: sync.RWMutex{}}
}

// Sign implements the beholder.Signer interface
func (l *lazyKeystoreSigner) Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error) {
	l.mu.RLock()
	ks := l.keystore
	l.mu.RUnlock()

	if ks == nil {
		return nil, fmt.Errorf("keystore not yet available for signing")
	}

	return ks.Sign(ctx, keyID, data)
}

// SetKeystore updates the underlying keystore. This is thread-safe and can be
// called at any time, even after beholder has been initialized.
func (l *lazyKeystoreSigner) SetKeystore(ks keystore.Keystore) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.keystore = ks
}

// HasKeystore returns true if a keystore has been set
func (l *lazyKeystoreSigner) HasKeystore() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.keystore != nil
}
