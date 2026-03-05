package solkey

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestSolanaKeys_ExportImport(t *testing.T) {
	corekeys.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys.KeyType, error) {
	return New()
}

func decryptKey(keyJSON []byte, password string) (corekeys.KeyType, error) {
	return FromEncryptedJSON(keyJSON, password)
}
