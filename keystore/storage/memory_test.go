package storage_test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	storage := storage.NewMemoryStorage()
	require.NoError(t, storage.PutEncryptedKeystore(context.Background(), []byte("test")))
	got, err := storage.GetEncryptedKeystore(context.Background())
	require.NoError(t, err)
	require.Equal(t, []byte("test"), got)
}
