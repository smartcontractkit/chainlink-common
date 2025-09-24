package keystore

import (
	"context"
	"fmt"
)

// EVM specific keys are managed through the EVM keystore
// which builds on top of the generic keystore interface.
// It prefixes the key name with "evm_" to distinguish it from other keys
// and restricts key types to secp256k1 only.
// Satisfying the Signer/Reader interfaces are optional
// depends on what the wrapping code wants to expose, perhaps it wants
// to encapsulate what can be signed.
var _ Signer = EVM{}
var _ Reader = EVM{}

const (
	evmKeyType    = "evm"
	evmKeyTypeKey = Secp256k1
)

type EVM struct {
	ks Keystore
}

func NewEVM(ks Keystore) *EVM {
	return &EVM{ks: ks}
}

// Insert a family prefix.
func (e EVM) buildKeyName(name string) string {
	return fmt.Sprintf("%s_%s", evmKeyType, name)
}

func (e EVM) Sign(ctx context.Context, name string, data []byte) ([]byte, error) {
	return e.ks.Sign(ctx, e.buildKeyName(name), data)
}

func (e EVM) Verify(ctx context.Context, name string, data []byte, signature []byte) (bool, error) {
	return e.ks.Verify(ctx, e.buildKeyName(name), data, signature)
}

func (e EVM) ListKeys(ctx context.Context) ([]KeyInfo, error) {
	// TODO: filter by evmKeyType
	return e.ks.ListKeys(ctx)
}

func (e EVM) GetKey(ctx context.Context, name string) (KeyInfo, error) {
	return e.ks.GetKey(ctx, e.buildKeyName(name))
}

func PubkeyToAddress(pubkey []byte) (string, error) {
	// TODO: return the evm address
	return "", nil
}

// Below are more restricted methods which ensure the right key type is used.
// EVM key is always secp256k1
func (e EVM) CreateKey(ctx context.Context, name string) (KeyInfo, error) {
	return e.ks.CreateKey(ctx, e.buildKeyName(name), evmKeyType)
}

func (e EVM) DeleteKey(ctx context.Context, name string) error {
	return e.ks.DeleteKey(ctx, e.buildKeyName(name))
}

func (e EVM) ImportKey(ctx context.Context, name string, data []byte) ([]byte, error) {
	return e.ks.ImportKey(ctx, e.buildKeyName(name), evmKeyType, data)
}
