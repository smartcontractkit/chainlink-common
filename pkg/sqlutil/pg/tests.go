package pg

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

// Deprecated: Use sqltest.NewDB
func NewTestDB(t testing.TB, dbURL string) *sqlx.DB {
	t.Fatal("Deprecated: Use sqltest.NewTestDB")
	return nil
}

// Deprecated: Use sqltest.TestURL
func TestURL(t testing.TB) string {
	t.Fatal("Deprecated: Use sqltest.TestURL instead")
	return ""
}
