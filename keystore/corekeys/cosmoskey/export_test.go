package cosmoskey

import (
	"testing"
)

func TestCosmosKeys_ExportImport(t *testing.T) {
	corekeys2.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys2.KeyType, error) {
	return New(), nil
}

func decryptKey(keyJSON []byte, password string) (corekeys2.KeyType, error) {
	return FromEncryptedJSON(keyJSON, password)
}
