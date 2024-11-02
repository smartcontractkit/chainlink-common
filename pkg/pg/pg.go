package pg

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/scylladb/go-reflectx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func NewSqlxDB(t testing.TB, dbURL string) *sqlx.DB {
	tests.SkipShortDB(t)
	err := RegisterTxDb(dbURL)
	if err != nil {
		t.Errorf("failed to register txdb dialect: %s", err.Error())
		return nil
	}
	db, err := sqlx.Open(string(TransactionWrappedPostgres), uuid.New().String())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	db.MapperFunc(reflectx.CamelToSnakeASCII)

	return db
}
