// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeDKG = "dkg"

func (ks *Store) GenerateEncryptedDKGKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeDKG, password)
}

func FromEncryptedDKGKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeDKG, password)
}
