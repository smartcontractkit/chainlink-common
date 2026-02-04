// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeP2P = "p2p"

func (ks *Store) GenerateEncryptedP2PKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeP2P, password)
}

func FromEncryptedP2PKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeP2P, password)
}
