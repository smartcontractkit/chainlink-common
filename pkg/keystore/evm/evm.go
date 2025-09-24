package evm

import (
	"context"
	"fmt"

	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
)

const (
	evmKeyTag = "evm"
)

var evmKeyType = keystore.Secp256k1

type EVMCreateKeyRequest struct {
	Name string
}

type EVMKeyInfo struct {
	KeyInfo keystore.KeyInfo
	Address string
}
type EVMCreateKeyResponse struct {
	EVMKeyInfo
}

type EVMDeleteKeyRequest struct {
	Name string
}

type EVMDeleteKeyResponse struct{}

type EVMImportKeyRequest struct {
	Name    string
	Data    []byte
	Address string // Sanity check
}

type EVMImportKeyResponse struct {
	PublicKey []byte
}

type EVMListKeysRequest struct{}

type EVMListKeysResponse struct {
	Keys []EVMKeyInfo
}

type EVMGetKeyRequest struct {
	Name string
	// Could allow for searching by address too.
}

type EVMGetKeyResponse struct {
	KeyInfo keystore.KeyInfo
}
type EVMSignRequest struct {
	Name string
	Data []byte
}
type EVMSignResponse struct {
	Signature []byte
}

type EVMVerifyRequest struct {
	Name      string
	Data      []byte
	Signature []byte
}
type EVMVerifyResponse struct {
	Valid bool
}

type EVM struct {
	ks keystore.Keystore
}

func NewEVM(ks keystore.Keystore) *EVM {
	return &EVM{ks: ks}
}

func (e EVM) buildKeyName(name string) string {
	return fmt.Sprintf("%s_%s", evmKeyTag, name)
}

func (e EVM) isEVMKey(name string) bool {
	return strings.HasPrefix(name, evmKeyTag)
}

func (e EVM) Sign(ctx context.Context, req EVMSignRequest) (EVMSignResponse, error) {
	signReq := keystore.SignRequest{
		Name: e.buildKeyName(req.Name),
		Data: req.Data,
	}
	signResp, err := e.ks.Sign(ctx, signReq)
	if err != nil {
		return EVMSignResponse{}, err
	}
	return EVMSignResponse{
		Signature: signResp.Signature,
	}, nil
}

func (e EVM) Verify(ctx context.Context, req EVMVerifyRequest) (EVMVerifyResponse, error) {
	verifyReq := keystore.VerifyRequest{
		Name:      e.buildKeyName(req.Name),
		Data:      req.Data,
		Signature: req.Signature,
	}
	verifyResp, err := e.ks.Verify(ctx, verifyReq)
	if err != nil {
		return EVMVerifyResponse{}, err
	}
	return EVMVerifyResponse{
		Valid: verifyResp.Valid,
	}, nil
}

func (e EVM) ListKeys(ctx context.Context, req EVMListKeysRequest) (EVMListKeysResponse, error) {
	listReq := keystore.ListKeysRequest{}
	listResp, err := e.ks.ListKeys(ctx, listReq)
	if err != nil {
		return EVMListKeysResponse{}, err
	}
	var keys []EVMKeyInfo
	for _, key := range listResp.Keys {
		if key.KeyType != evmKeyType {
			continue
		}
		if !e.isEVMKey(key.Name) {
			continue
		}
		address, err := PubkeyToAddress(key.PublicKey)
		if err != nil {
			return EVMListKeysResponse{}, err
		}
		keys = append(keys, EVMKeyInfo{KeyInfo: key, Address: address})
	}

	return EVMListKeysResponse{
		Keys: keys,
	}, nil
}

func (e EVM) GetKey(ctx context.Context, req EVMGetKeyRequest) (EVMGetKeyResponse, error) {
	getReq := keystore.GetKeyRequest{
		Name: e.buildKeyName(req.Name),
	}
	getResp, err := e.ks.GetKey(ctx, getReq)
	if err != nil {
		return EVMGetKeyResponse{}, err
	}
	return EVMGetKeyResponse{
		KeyInfo: getResp.KeyInfo,
	}, nil
}

func PubkeyToAddress(pubkey []byte) (string, error) {
	// TODO: return the evm address
	return "", nil
}

// Below are more restricted methods which ensure the right key type is used.
// EVM key is always secp256k1
func (e EVM) CreateKey(ctx context.Context, req EVMCreateKeyRequest) (EVMCreateKeyResponse, error) {
	createReq := keystore.CreateKeyRequest{
		Name:    e.buildKeyName(req.Name),
		KeyType: evmKeyType,
	}
	createResp, err := e.ks.CreateKey(ctx, createReq)
	if err != nil {
		return EVMCreateKeyResponse{}, err
	}
	address, err := PubkeyToAddress(createResp.KeyInfo.PublicKey)
	if err != nil {
		return EVMCreateKeyResponse{}, err
	}
	return EVMCreateKeyResponse{
		EVMKeyInfo: EVMKeyInfo{
			KeyInfo: createResp.KeyInfo,
			Address: address,
		},
	}, nil
}

func (e EVM) DeleteKey(ctx context.Context, req EVMDeleteKeyRequest) (EVMDeleteKeyResponse, error) {
	deleteReq := keystore.DeleteKeyRequest{
		Name: e.buildKeyName(req.Name),
	}
	_, err := e.ks.DeleteKey(ctx, deleteReq)
	if err != nil {
		return EVMDeleteKeyResponse{}, err
	}
	return EVMDeleteKeyResponse{}, nil
}

func (e EVM) ImportKey(ctx context.Context, req EVMImportKeyRequest) (EVMImportKeyResponse, error) {
	importReq := keystore.ImportKeyRequest{
		Name:    e.buildKeyName(req.Name),
		KeyType: evmKeyType,
		Data:    req.Data,
	}
	importResp, err := e.ks.ImportKey(ctx, importReq)
	if err != nil {
		return EVMImportKeyResponse{}, err
	}
	return EVMImportKeyResponse{
		PublicKey: importResp.PublicKey,
	}, nil
}
