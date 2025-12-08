package keystore_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestFileStorage(t *testing.T) {
	t.Parallel()
	storage := keystore.NewFileStorage(filepath.Join(t.TempDir(), "out.txt"))
	_, err := storage.GetEncryptedKeystore(t.Context())
	require.ErrorContains(t, err, "no such file or directory")
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("test")))
	got, err := storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("test"), got)
}
