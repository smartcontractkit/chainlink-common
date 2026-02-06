// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeCSA = "csa"

func (ks *Store) GenerateEncryptedCSAKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeCSA, password)
}

func FromEncryptedCSAKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeCSA, password)
}
