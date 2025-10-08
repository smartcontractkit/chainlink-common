package storage_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	storage := storage.NewMemoryStorage()
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("test")))
	got, err := storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("test"), got)
}
