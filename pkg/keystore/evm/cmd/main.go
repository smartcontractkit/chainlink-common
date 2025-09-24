package main

import (
	"context"
	"fmt"
	"log"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	evmks "github.com/smartcontractkit/chainlink-common/pkg/keystore/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
)

const (
	name = "myscript"
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
	evm := evmks.NewEVM(keystore)
	keyInfo, err := evm.CreateKey(context.Background(), evmks.EVMCreateKeyRequest{Name: name})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(keyInfo)
	blah, err := evm.SignTx(context.Background(), evmks.EVMSignTxRequest{
		Name: name, Tx: &gethtypes.Transaction{}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(blah)
}
