package pgstore_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/pgstore"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

func TestPgStorage(t *testing.T) {
	t.Parallel()
	sqltest.SkipInMemory(t)
	db := sqltest.NewDB(t, sqltest.TestURL(t))
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	storage := pgstore.NewStorage(db, "test-name")
	_, err := storage.GetEncryptedKeystore(t.Context())
	require.ErrorIs(t, err, sql.ErrNoRows)
	// Test insert
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("aaa")))
	got, err := storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("aaa"), got)
	// Test update
	require.NoError(t, storage.PutEncryptedKeystore(t.Context(), []byte("bbb")))
	got, err = storage.GetEncryptedKeystore(t.Context())
	require.NoError(t, err)
	require.Equal(t, []byte("bbb"), got)
}
