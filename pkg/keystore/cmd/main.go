package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/file"
)

const (
	name = "myscript"
)

func main() {
	fileKeystore, err := file.NewFileKeystore("test_password", "test_keystore.json")
	if err != nil {
		log.Fatal(err)
	}
	evm := keystore.NewEVM(fileKeystore)
	keyInfo, err := evm.CreateKey(context.Background(), name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(keyInfo)
	blah, err := evm.Sign(context.Background(), name, []byte("hello world"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(blah)
}
