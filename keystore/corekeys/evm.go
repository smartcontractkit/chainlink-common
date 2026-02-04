// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeEVM = "evm"

func (ks *Store) GenerateEncryptedEVMKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeEVM, password)
}

func FromEncryptedEVMKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeEVM, password)
}
