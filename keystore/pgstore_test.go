package keystore_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

func TestPgStorage(t *testing.T) {
	t.Parallel()
	sqltest.SkipInMemory(t)
	db := sqltest.NewDB(t, sqltest.TestURL(t))
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	storage := keystore.NewPgStorage(db, "test")
	_, err := storage.GetEncryptedKeystore(t.Context())
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("test")))
	got, err := storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("test"), got)
}
