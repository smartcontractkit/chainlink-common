package evm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
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
	Address string // sanity check
}

type EVMImportKeyResponse struct {
	PublicKey []byte
}

type EVMExportKeyRequest struct {
	Name string
}
type EVMExportKeyResponse struct {
	Name    string
	Address string
	Data    []byte
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
type EVMSignTxRequest struct {
	Name        string
	FromAddress string
	ChainID     *big.Int
	Tx          *gethtypes.Transaction
}
type EVMSignTxResponse struct {
	Tx *gethtypes.Transaction
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

// Sign an EVM transaction.
func (e EVM) SignTx(ctx context.Context, req EVMSignTxRequest) (EVMSignTxResponse, error) {
	signer := gethtypes.LatestSignerForChainID(req.ChainID)
	h := signer.Hash(req.Tx)
	signReq := keystore.SignRequest{
		Name: e.buildKeyName(req.Name),
		Data: h[:],
	}
	signResp, err := e.ks.Sign(ctx, signReq)
	if err != nil {
		return EVMSignTxResponse{}, err
	}
	req.Tx, err = req.Tx.WithSignature(signer, signResp.Signature)
	if err != nil {
		return EVMSignTxResponse{}, err
	}
	return EVMSignTxResponse{
		Tx: req.Tx,
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
	// TODO:
	return "", nil
}

func (e EVM) CreateKey(ctx context.Context, req EVMCreateKeyRequest) (EVMCreateKeyResponse, error) {
	// Only EVM key types created.
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

func (e EVM) ExportKey(ctx context.Context, req EVMExportKeyRequest) (EVMExportKeyResponse, error) {
	exportReq := keystore.ExportKeyRequest{
		Name: e.buildKeyName(req.Name),
	}
	exportResp, err := e.ks.ExportKey(ctx, exportReq)
	if err != nil {
		return EVMExportKeyResponse{}, err
	}
	address, err := PubkeyToAddress(exportResp.KeyInfo.PublicKey)
	if err != nil {
		return EVMExportKeyResponse{}, err
	}
	return EVMExportKeyResponse{
		Name:    exportResp.KeyInfo.Name,
		Address: address,
		Data:    exportResp.Data,
	}, nil
}
