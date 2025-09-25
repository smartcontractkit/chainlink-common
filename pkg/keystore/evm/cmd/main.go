package main

import (
	"context"
	"fmt"
	"log"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	evmks "github.com/smartcontractkit/chainlink-common/pkg/keystore/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/smartcontractkit/libocr/offchainreporting2plus"
)

const (
	name     = "evm-tx-key"
	ocr2Name = "ocr2-key-bundle"
)

func main() {
	storage, err := storage.NewFileStorage("test_keystore.json")
	if err != nil {
		log.Fatal(err)
	}
	keystore, err := keystore.NewKeystore(storage, "test_password")
	if err != nil {
		log.Fatal(err)
	}
	evmTxStore := evmks.NewEVM(keystore)
	keyInfo, err := evmTxStore.CreateKey(context.Background(), evmks.EVMCreateKeyRequest{Name: name})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(keyInfo)
	blah, err := evmTxStore.SignTx(context.Background(), evmks.EVMSignTxRequest{
		Name: name, Tx: &gethtypes.Transaction{}})
	if err != nil {
		log.Fatal(err)
	}
	_, err = evmTxStore.DeleteKey(context.Background(), evmks.EVMDeleteKeyRequest{Name: name})
	if err != nil {
		log.Fatal(err)
	}

	evmOCROnchainStore := evmks.NewOCR2OnchainKeyringStore(keystore)
	onchainKeyringResp, err := evmOCROnchainStore.CreateKeyring(context.Background(), evmks.OCR2OnchainKeyringCreateRequest{Name: ocr2Name})
	if err != nil {
		log.Fatal(err)
	}
	evmOCROffchainStore := evmks.NewOCR2OffchainKeyringStore(keystore)
	offchainKeyringResp, err := evmOCROffchainStore.CreateKeyring(context.Background(), evmks.OCR2OffchainKeyringCreateRequest{Name: ocr2Name})
	if err != nil {
		log.Fatal(err)
	}
	_, err = offchainreporting2plus.NewOracle(offchainreporting2plus.OCR2OracleArgs{
		OnchainKeyring:  onchainKeyringResp.Keyring,
		OffchainKeyring: offchainKeyringResp.Keyring,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(blah)
}
