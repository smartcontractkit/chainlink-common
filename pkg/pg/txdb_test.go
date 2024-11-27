package pg

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestTxDBDriver(t *testing.T) {
	t.Parallel()
	dbURL, ok := os.LookupEnv("CL_DATABASE_URL")
	if !ok {
		t.Log("CL_DATABASE_URL not set--falling back to testing txdb backed by an in-memory db")
		dbURL = string(InMemoryPostgres)
	}
	db := NewSqlxDB(t, dbURL)
	dropTable := func() error {
		_, err := db.Exec(`DROP TABLE IF EXISTS txdb_test`)
		return err
	}
	// clean up, if previous tests failed
	err := dropTable()
	assert.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE txdb_test (id TEXT NOT NULL)`)
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = dropTable()
	})
	_, err = db.Exec(`INSERT INTO txdb_test VALUES ($1)`, uuid.New().String())
	assert.NoError(t, err)
	ensureValuesPresent := func(t *testing.T, db *sqlx.DB) {
		var ids []string
		err = db.Select(&ids, `SELECT id from txdb_test`)
		assert.NoError(t, err)
		assert.Len(t, ids, 1)
	}

	ensureValuesPresent(t, db)
	t.Run("Cancel of tx's context does not trigger rollback of driver's tx", func(t *testing.T) {
		ctx, cancel := context.WithCancel(tests.Context(t))
		_, err := db.BeginTx(ctx, nil)
		assert.NoError(t, err)
		cancel()
		// BeginTx spawns separate goroutine that rollbacks the tx and tries to close underlying connection, unless
		// db driver says that connection is still active.
		// This approach is not ideal, but there is no better way to wait for independent goroutine to complete
		time.Sleep(time.Second * 10)
		ensureValuesPresent(t, db)
	})

	t.Run("Make sure calling sql.Register() can be called twice", func(t *testing.T) {
		require.NoError(t, RegisterTxDb("foo"))
		require.NoError(t, RegisterTxDb("bar"))
		drivers := sql.Drivers()
		assert.Contains(t, drivers, "txdb")
	})
	t.Run("Make sure sql.Register() can be called concurrently without racing", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			go RegisterTxDb(strconv.Itoa(i))
		}
	})
}
