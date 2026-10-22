package sqlutil

import (
	"database/sql"
	"errors"

	// Register the pgx database/sql driver under the name "pgx".
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config is the configuration for connecting to the database. Shaped for chainlink-common's
// x/config/cli, so a binary can bind it as flags, env vars and config-file keys directly.
type Config struct {
	// URL is the database url, such as postgresql://user:password@localhost:5432/chainlink?sslmode=disable.
	URL string
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
