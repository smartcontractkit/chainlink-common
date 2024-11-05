package sqlutil

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/scylladb/go-reflectx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTxDBDriver(t *testing.T) {
	duckdb, err := sql.Open("duckdb", "")
	const driverName = "txdb_test"
	sql.Register(driverName, &txDriver{
		db:    duckdb,
		conns: make(map[string]*conn),
	})
	sqlx.BindDriver(driverName, sqlx.DOLLAR)

	db, err := sqlx.Open(driverName, uuid.New().String())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	db.MapperFunc(reflectx.CamelToSnakeASCII)

	dropTable := func() error {
		_, err := db.Exec(`DROP TABLE IF EXISTS txdb_test`)
		return err
	}
	// clean up, if previous tests failed
	err = dropTable()
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
		ctx, cancel := context.WithCancel(context.Background())
		_, err := db.BeginTx(ctx, nil)
		assert.NoError(t, err)
		cancel()

		// BeginTx spawns separate goroutine that rollbacks the tx and tries to close underlying connection, unless
		// db driver says that connection is still active.
		// This approach is not ideal, but there is no better way to wait for independent goroutine to complete
		time.Sleep(time.Second * 2)
		ensureValuesPresent(t, db)
	})

	t.Run("Test statement", func(t *testing.T) {
		stmt, err := db.Prepare("SELECT id FROM txdb_test")
		defer stmt.Close()
		assert.NoError(t, err)
		rows, err := stmt.Query()
		assert.True(t, rows.Next())
		var id string
		rows.Scan(id)
		assert.False(t, rows.Next())

		_, err = stmt.Exec()
		assert.NoError(t, err)

		_, err = stmt.ExecContext(context.Background())
		assert.NoError(t, err)

		_, err = stmt.QueryContext(context.Background())
		assert.NoError(t, err)
	})
}
