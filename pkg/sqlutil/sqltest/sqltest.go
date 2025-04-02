package sqltest

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
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

// CreateOrReplace creates a new database with the given name (optionally from template), and schedules it to be dropped
// after test completion.
func CreateOrReplace(t testing.TB, u url.URL, dbName string, template string) url.URL {
	if u.Path == "" {
		t.Fatal("path missing from database URL")
	}

	if l := len(dbName); l > 63 {
		t.Fatalf("dbName %v too long (%d), max is 63 bytes", dbName, l)
	}
	// Cannot drop test database if we are connected to it, so we must connect
	// to a different one. 'postgres' should be present on all postgres installations
	u.Path = "/postgres"
	db, err := sql.Open(pg.DriverPostgres, u.String())
	if err != nil {
		t.Fatalf("in order to drop the test database, we need to connect to a separate database"+
			" called 'postgres'. But we are unable to open 'postgres' database: %+v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		t.Fatalf("unable to drop postgres migrations test database: %v", err)
	}
	if template != "" {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", dbName, template))
	} else {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	}
	if err != nil {
		t.Fatalf("unable to create postgres test database with name '%s': %v", dbName, err)
	}
	u.Path = fmt.Sprintf("/%s", dbName)
	t.Cleanup(func() { assert.NoError(t, drop(u)) })
	return u
}

// drop drops the database at the given URL.
func drop(dbURL url.URL) error {
	if dbURL.Path == "" {
		return errors.New("path missing from database URL")
	}
	dbname := strings.TrimPrefix(dbURL.Path, "/")

	// Cannot drop test database if we are connected to it, so we must connect
	// to a different one. 'postgres' should be present on all postgres installations
	dbURL.Path = "/postgres"
	db, err := sql.Open(pg.DriverPostgres, dbURL.String())
	if err != nil {
		return fmt.Errorf("in order to drop the test database, we need to connect to a separate database"+
			" called 'postgres'. But we are unable to open 'postgres' database: %+v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
	if err != nil {
		return fmt.Errorf("unable to drop postgres migrations test database: %v", err)
	}
	return nil
}
