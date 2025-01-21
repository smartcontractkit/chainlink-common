package sqltest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

// NewInMemoryDataSource returns a new in-memory DataSource
func NewInMemoryDataSource(t *testing.T) sqlutil.DataSource {
	db, err := sqlx.Open("duckdb", "")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

// NewNoOpDataSource returns an empty DataSource type which will satisfy the interface
func NewNoOpDataSource() sqlutil.DataSource {
	return &noOpDataSource{}
}

type noOpDataSource struct{}

func (ds *noOpDataSource) BindNamed(s string, _ interface{}) (string, []interface{}, error) {
	return "", nil, nil
}

func (ds *noOpDataSource) DriverName() string {
	return ""
}

func (ds *noOpDataSource) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (ds *noOpDataSource) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (ds *noOpDataSource) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return nil, nil
}

func (ds *noOpDataSource) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}

func (_m *noOpDataSource) PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	return nil, nil
}

func (ds *noOpDataSource) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (ds *noOpDataSource) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return nil
}

func (ds *noOpDataSource) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return nil, nil
}

func (ds *noOpDataSource) Rebind(s string) string {
	return ""
}

func (ds *noOpDataSource) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
