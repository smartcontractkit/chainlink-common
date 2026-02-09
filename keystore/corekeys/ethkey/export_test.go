package ethkey

import (
	"testing"
)

func TestEthKeys_ExportImport(t *testing.T) {
	corekeys2.RunKeyExportImportTestcase(t, createKey, func(keyJSON []byte, password string) (kt corekeys2.KeyType, err error) {
		t.SkipNow()
		return kt, err
	})
}

func createKey() (corekeys2.KeyType, error) {
	return NewV2()
}
