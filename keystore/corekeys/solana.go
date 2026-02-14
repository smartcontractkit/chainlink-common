// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeSolana = "solana"

func (ks *Store) GenerateEncryptedSolanaKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeSolana, password)
}

func FromEncryptedSolanaKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeSolana, password)
}
