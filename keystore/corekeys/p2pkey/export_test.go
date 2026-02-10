package p2pkey

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestP2PKeys_ExportImport(t *testing.T) {
	corekeys.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys.KeyType, error) {
	return NewV2()
}

func decryptKey(keyJSON []byte, password string) (corekeys.KeyType, error) {
	return FromEncryptedJSON(keyJSON, password)
}
