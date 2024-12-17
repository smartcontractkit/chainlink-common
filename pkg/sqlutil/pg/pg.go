package pg

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/scylladb/go-reflectx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewTestDB(t testing.TB, dbURL string) *sqlx.DB {
	err := RegisterTxDb(dbURL)
	if err != nil {
		t.Fatalf("failed to register txdb dialect: %s", err.Error())
		return nil
	}
	db, err := sqlx.Open(string(TransactionWrappedPostgres), uuid.New().String())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	db.MapperFunc(reflectx.CamelToSnakeASCII)

	return db
}

func TestURL(t testing.TB) string {
	dbURL, ok := os.LookupEnv("CL_DATABASE_URL")
	if ok {
		return dbURL
	}
	t.Log("CL_DATABASE_URL not set--falling back to testing txdb backed by an in-memory db")
	return string(InMemoryPostgres)
}
