package sqltest

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/scylladb/go-reflectx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
)

const envVar = "CL_DATABASE_URL"

// NewDB returns a new test database for the given URL.
func NewDB(t testing.TB, dbURL string) *sqlx.DB {
	err := RegisterTxDB(dbURL)
	if err != nil {
		t.Fatalf("failed to register txdb dialect: %s", err.Error())
		return nil
	}
	db, err := sqlx.Open(pg.DriverTxWrappedPostgres, uuid.New().String())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	db.MapperFunc(reflectx.CamelToSnakeASCII)

	return db
}

// TestURL returns a database URL for this test.
// It may be external (if configured) or in-memory.
func TestURL(t testing.TB) string {
	dbURL, ok := os.LookupEnv(envVar)
	if ok {
		return dbURL
	}
	t.Logf("%s not set--falling back to testing txdb backed by an in-memory db", envVar)
	return pg.DriverInMemoryPostgres
}

// SkipInMemory skips the test when using an in-memory database.
func SkipInMemory(t *testing.T) {
	if _, ok := os.LookupEnv(envVar); !ok {
		t.Skip("skipping test due to in-memory db")
	}
}
