package ethkey

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestEthKeys_ExportImport(t *testing.T) {
	corekeys.RunKeyExportImportTestcase(t, createKey, func(keyJSON []byte, password string) (kt corekeys.KeyType, err error) {
		t.SkipNow()
		return kt, err
	})
}

func createKey() (corekeys.KeyType, error) {
	return NewV2()
}
