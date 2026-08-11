package sqlutil

import (
	"database/sql"
	"errors"

	// Register the pgx database/sql driver under the name "pgx".
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config is the configuration for connecting to the database. Tagged for
// chainlink-common's pkg/config/flags, so a binary can bind it as flags, env vars and
// config-file keys directly.
type Config struct {
	URL string `usage:"database url" example:"'postgresql://user:password@localhost:5432/chainlink?sslmode=disable'"`
}

// OpenDB opens the database at c.URL with the pgx driver.
//
// As with [sql.Open], no connection is established until the returned DB is used, so a bad
// host or credentials surface on first query rather than here. The caller owns the DB and must
// Close it.
func OpenDB(c Config) (*sql.DB, error) {
	if c.URL == "" {
		return nil, errors.New("database url is required")
	}
	return sql.Open("pgx", c.URL)
}
