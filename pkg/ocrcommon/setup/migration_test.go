package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

func TestCreateMigration_Naming(t *testing.T) {
	const table = "my_announcements"

	t.Run("no existing migrations creates 0001", func(t *testing.T) {
		dir := t.TempDir()

		path, err := CreateMigration(dir, table)
		require.NoError(t, err)

		assert.Equal(t, "0001_create_my_announcements.sql", filepath.Base(path))
	})

	t.Run("N existing migrations creates N+1", func(t *testing.T) {
		dir := t.TempDir()

		// Seed an arbitrary number of pre-existing migrations.
		existing := []string{
			"0001_initial.sql",
			"0002_something.sql",
			"0003_another.sql",
		}
		for _, name := range existing {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("-- +goose Up\n-- +goose Down\n"), 0o644))
		}
		// A non-migration file must be ignored when choosing the next number.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore me"), 0o644))

		path, err := CreateMigration(dir, table)
		require.NoError(t, err)

		assert.Equal(t, "0004_create_my_announcements.sql", filepath.Base(path))
	})

	t.Run("rejects invalid table names", func(t *testing.T) {
		dir := t.TempDir()
		_, err := CreateMigration(dir, "bad; DROP TABLE users")
		require.Error(t, err)
	})
}

func TestCreateMigration_UpDown(t *testing.T) {
	const table = "updown_announcements"

	sqltest.SkipInMemory(t)
	db := sqltest.NewDB(t, sqltest.TestURL(t))
	ctx := t.Context()

	dir := t.TempDir()
	_, err := CreateMigration(dir, table)
	require.NoError(t, err)

	provider, err := goose.NewProvider(goose.DialectPostgres, db.DB, os.DirFS(dir))
	require.NoError(t, err)

	t.Run("table exists after up", func(t *testing.T) {
		_, err := provider.Up(ctx)
		require.NoError(t, err)
		assert.True(t, tableExists(t, ctx, db, table))
	})

	t.Run("table removed after down", func(t *testing.T) {
		_, err := provider.Down(ctx)
		require.NoError(t, err)
		assert.False(t, tableExists(t, ctx, db, table))
	})
}

func tableExists(t *testing.T, ctx context.Context, db *sqlx.DB, table string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists))
	return exists
}
