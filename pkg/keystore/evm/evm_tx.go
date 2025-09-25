package evm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

type EVMTxKeystore interface {
	SignTx(ctx context.Context, req EVMSignTxRequest) (EVMSignTxResponse, error)
	ListKeys(ctx context.Context, req EVMListKeysRequest) (EVMListKeysResponse, error)
	GetKey(ctx context.Context, req EVMGetKeyRequest) (EVMGetKeyResponse, error)
	CreateKey(ctx context.Context, req EVMCreateKeyRequest) (EVMCreateKeyResponse, error)
	DeleteKey(ctx context.Context, req EVMDeleteKeyRequest) (EVMDeleteKeyResponse, error)
	ImportKey(ctx context.Context, req EVMImportKeyRequest) (EVMImportKeyResponse, error)
	ExportKey(ctx context.Context, req EVMExportKeyRequest) (EVMExportKeyResponse, error)
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
	getReq := keystore.GetKeysRequest{
		Names: []string{e.buildKeyName(req.Name)},
	}
	getResp, err := e.ks.GetKeys(ctx, getReq)
	if err != nil {
		return EVMGetKeyResponse{}, err
	}
	if len(getResp.Keys) == 0 {
		return EVMGetKeyResponse{}, fmt.Errorf("key not found: %s", req.Name)
	}
	return EVMGetKeyResponse{
		KeyInfo: getResp.Keys[0].KeyInfo,
	}, nil
}

func PubkeyToAddress(pubkey []byte) (string, error) {
	// Convert public key to Ethereum address
	ecdsaPubkey, err := crypto.UnmarshalPubkey(pubkey)
	if err != nil {
		return "", err
	}
	address := crypto.PubkeyToAddress(*ecdsaPubkey)
	return address.Hex(), nil
}

func (e EVM) CreateKey(ctx context.Context, req EVMCreateKeyRequest) (EVMCreateKeyResponse, error) {
	// Only EVM key types created.
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				Name:    e.buildKeyName(req.Name),
				KeyType: evmKeyType,
			},
		},
	}
	createResp, err := e.ks.CreateKeys(ctx, createReq)
	if err != nil {
		return EVMCreateKeyResponse{}, err
	}
	if len(createResp.Keys) == 0 {
		return EVMCreateKeyResponse{}, fmt.Errorf("no keys created")
	}
	address, err := PubkeyToAddress(createResp.Keys[0].KeyInfo.PublicKey)
	if err != nil {
		return EVMCreateKeyResponse{}, err
	}
	return EVMCreateKeyResponse{
		EVMKeyInfo: EVMKeyInfo{
			KeyInfo: createResp.Keys[0].KeyInfo,
			Address: address,
		},
	}, nil
}

func (e EVM) DeleteKey(ctx context.Context, req EVMDeleteKeyRequest) (EVMDeleteKeyResponse, error) {
	deleteReq := keystore.DeleteKeysRequest{
		Names: []string{e.buildKeyName(req.Name)},
	}
	_, err := e.ks.DeleteKeys(ctx, deleteReq)
	if err != nil {
		return EVMDeleteKeyResponse{}, err
	}
	return EVMDeleteKeyResponse{}, nil
}

func (e EVM) ImportKey(ctx context.Context, req EVMImportKeyRequest) (EVMImportKeyResponse, error) {
	importReq := keystore.ImportKeysRequest{
		Keys: []keystore.ImportKeyRequest{
			{
				Name:    e.buildKeyName(req.Name),
				KeyType: evmKeyType,
				Data:    req.Data,
			},
		},
	}
	importResp, err := e.ks.ImportKeys(ctx, importReq)
	if err != nil {
		return EVMImportKeyResponse{}, err
	}
	if len(importResp.Keys) == 0 {
		return EVMImportKeyResponse{}, fmt.Errorf("no keys imported")
	}
	return EVMImportKeyResponse{
		PublicKey: importResp.Keys[0].PublicKey,
	}, nil
}

func (e EVM) ExportKey(ctx context.Context, req EVMExportKeyRequest) (EVMExportKeyResponse, error) {
	exportReq := keystore.ExportKeysRequest{
		Names: []string{e.buildKeyName(req.Name)},
	}
	exportResp, err := e.ks.ExportKeys(ctx, exportReq)
	if err != nil {
		return EVMExportKeyResponse{}, err
	}
	if len(exportResp.Keys) == 0 {
		return EVMExportKeyResponse{}, fmt.Errorf("key not found: %s", req.Name)
	}
	address, err := PubkeyToAddress(exportResp.Keys[0].KeyInfo.PublicKey)
	if err != nil {
		return EVMExportKeyResponse{}, err
	}
	return EVMExportKeyResponse{
		Name:    exportResp.Keys[0].KeyInfo.Name,
		Address: address,
		Data:    exportResp.Keys[0].Data,
	}, nil
}
