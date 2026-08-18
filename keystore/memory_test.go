package keystore_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestMemoryStorage(t *testing.T) {
	t.Parallel()
	storage := keystore.NewMemoryStorage()
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("test")))
	got, err := storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("test"), got)
}
