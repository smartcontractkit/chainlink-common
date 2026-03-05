package starkkey

import (
	"testing"

	commonkeystore "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestStarkNetKeys_ExportImport(t *testing.T) {
	corekeys.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys.KeyType, error) {
	key, err := New()
	return TestWrapped{key}, err
}

func decryptKey(keyJSON []byte, password string) (corekeys.KeyType, error) {
	key, err := FromEncryptedJSON(keyJSON, password)
	return TestWrapped{key}, err
}

// wrap key to conform to desired test interface
type TestWrapped struct {
	Key
}

func (w TestWrapped) ToEncryptedJSON(password string, scryptParams commonkeystore.ScryptParams) ([]byte, error) {
	return ToEncryptedJSON(w.Key, password, scryptParams)
}
