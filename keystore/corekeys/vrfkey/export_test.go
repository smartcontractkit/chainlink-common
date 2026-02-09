package vrfkey

import (
	"testing"
)

func TestVRFKeys_ExportImport(t *testing.T) {
	corekeys2.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys2.KeyType, error) {
	return NewV2()
}

func decryptKey(keyJSON []byte, password string) (corekeys2.KeyType, error) {
	return FromEncryptedJSON(keyJSON, password)
}
