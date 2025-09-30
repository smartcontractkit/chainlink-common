package evm

import (
	"context"
	"fmt"

	"math/big"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
)

type TxKey struct {
	ks keystore.Keystore
	// Fully qualified name in keystore. Use for administration.
	FullName string
	Name     string
}

type SignTxRequest struct {
	ChainID *big.Int
	Tx      *gethtypes.Transaction
}

type SignTxResponse struct {
	Tx *gethtypes.Transaction
}

func (k *TxKey) SignTx(ctx context.Context, req SignTxRequest) (SignTxResponse, error) {
	signer := gethtypes.LatestSignerForChainID(req.ChainID)
	h := signer.Hash(req.Tx)
	signReq := keystore.SignRequest{
		Name: k.Name,
		Data: h[:],
	}
	signResp, err := k.ks.Sign(ctx, signReq)
	if err != nil {
		return SignTxResponse{}, err
	}
	req.Tx, err = req.Tx.WithSignature(signer, signResp.Signature)
	if err != nil {
		return SignTxResponse{}, err
	}
	return SignTxResponse{Tx: req.Tx}, nil
}

func CreateTxKey(ks keystore.Keystore, localName string) (*TxKey, error) {
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				Name:    GetTxKeystoreName(localName),
				KeyType: keystore.Secp256k1,
			},
		},
	}
	resp, err := ks.CreateKeys(context.Background(), createReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Keys) == 0 {
		return nil, fmt.Errorf("no keys created")
	}
	return &TxKey{
		ks:       ks,
		Name:     localName,
		FullName: GetTxKeystoreName(localName),
	}, nil
}

func GetTxKeys(ctx context.Context, ks keystore.Keystore, names []string) ([]*TxKey, error) {
	var fullNames []string
	for _, name := range names {
		fullNames = append(fullNames, GetTxKeystoreName(name))
	}
	resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{Names: fullNames})
	if err != nil {
		return nil, err
	}

	var keys []*TxKey
	for i, key := range resp.Keys {
		keys = append(keys, &TxKey{
			ks:       ks,
			FullName: key.Name,
			Name:     names[i],
		})
	}
	return keys, nil
}
