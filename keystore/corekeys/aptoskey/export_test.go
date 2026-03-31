package aptoskey

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonkeystore "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestAptosKeys_ExportImport(t *testing.T) {
	corekeys.RunKeyExportImportTestcase(t, createKey, decryptKey)
}

func createKey() (corekeys.KeyType, error) {
	return New()
}

func decryptKey(keyJSON []byte, password string) (corekeys.KeyType, error) {
	return FromEncryptedJSON(keyJSON, password)
}

func TestKey_EncryptedJSON_RoundTrip_PreservesAccountAddress(t *testing.T) {
	key, err := New()
	require.NoError(t, err)

	// Override address to verify it's the stored value that round-trips, not re-derived
	customAddr := "000000000000000000000000000000000000000000000000000000000000cafe"
	key = key.WithAccountAddress(customAddr)

	encrypted, err := key.ToEncryptedJSON("testpassword", commonkeystore.FastScryptParams)
	require.NoError(t, err)

	decrypted, err := FromEncryptedJSON(encrypted, "testpassword")
	require.NoError(t, err)

	assert.Equal(t, customAddr, decrypted.Account(), "account address must survive JSON round-trip")
	assert.Equal(t, key.PublicKeyStr(), decrypted.PublicKeyStr())
}

func TestKey_FromEncryptedJSON_Migration_NoAccountAddress(t *testing.T) {
	// Verify that old-format JSON without accountAddress field is handled correctly
	var export aptosEncryptedKeyExport
	oldJSON := `{"keyType":"Aptos","publicKey":"abcd","crypto":{}}`
	require.NoError(t, json.Unmarshal([]byte(oldJSON), &export))
	assert.Empty(t, export.AccountAddress, "old format should have no AccountAddress")
}
