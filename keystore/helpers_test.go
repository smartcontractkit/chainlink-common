package keystore_test

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"

	"github.com/stretchr/testify/require"
)

type Key struct {
	KeyType   keystore.KeyType
	PublicKey []byte
}

type KeystoreTH struct {
	mu         sync.RWMutex
	Keystore   keystore.Keystore
	keysByName map[string]Key
	keysByType map[keystore.KeyType][]Key
}

func NewKeystoreTH(t *testing.T) *KeystoreTH {
	ctx := t.Context()
	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(scrypt.FastScryptParams),
		keystore.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))),
	)
	require.NoError(t, err)
	return &KeystoreTH{
		Keystore:   ks,
		keysByName: make(map[string]Key),
		keysByType: make(map[keystore.KeyType][]Key),
	}
}

func (th *KeystoreTH) KeysByName() map[string]Key {
	th.mu.RLock()
	defer th.mu.RUnlock()
	return th.keysByName
}

func (th *KeystoreTH) KeysByType() map[keystore.KeyType][]Key {
	th.mu.RLock()
	defer th.mu.RUnlock()
	return th.keysByType
}

func (th *KeystoreTH) KeyName(keyType keystore.KeyType, index int) string {
	return fmt.Sprintf("test-key-%s-%d", keyType, index)
}

// CreateTestKeys creates 2 keys of each type in the keystore.
func (th *KeystoreTH) CreateTestKeys(t *testing.T) {
	th.mu.Lock()
	defer th.mu.Unlock()
	ctx := t.Context()
	for _, keyType := range keystore.AllKeyTypes {
		keys, err := th.Keystore.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{KeyName: th.KeyName(keyType, 0), KeyType: keyType},
				{KeyName: th.KeyName(keyType, 1), KeyType: keyType},
			},
		})
		require.NoError(t, err)
		th.keysByName[keys.Keys[0].KeyInfo.Name] = Key{KeyType: keys.Keys[0].KeyInfo.KeyType, PublicKey: keys.Keys[0].KeyInfo.PublicKey}
		th.keysByType[keyType] = append(th.keysByType[keyType], Key{KeyType: keys.Keys[0].KeyInfo.KeyType, PublicKey: keys.Keys[0].KeyInfo.PublicKey})

		th.keysByName[keys.Keys[1].KeyInfo.Name] = Key{KeyType: keys.Keys[1].KeyInfo.KeyType, PublicKey: keys.Keys[1].KeyInfo.PublicKey}
		th.keysByType[keyType] = append(th.keysByType[keyType], Key{KeyType: keys.Keys[1].KeyInfo.KeyType, PublicKey: keys.Keys[1].KeyInfo.PublicKey})
	}
}
