package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/scylladb/go-reflectx"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	// need to make sure pgx driver is registered before opening connection
	_ "github.com/jackc/pgx/v4/stdlib"
)

// NOTE: This is the default level in Postgres anyway, we just make it explicit here.
const defaultIsolation = sql.LevelReadCommitted

const (
	// DriverPostgres represents the postgres dialect.
	DriverPostgres = "pgx"
	// DriverTxWrappedPostgres is useful for tests.
	// When the connection is opened, it starts a transaction and all
	// operations performed on the DB will be within that transaction.
	DriverTxWrappedPostgres = "txdb"
	// DriverInMemoryPostgres uses an in memory database, github.com/marcboeker/go-duckdb.
	DriverInMemoryPostgres = "duckdb"
)

var (
	otelOpts = []otelsql.Option{otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
			OmitConnQuery:        false,
		})}
)

type DBConfig struct {
	IdleInTxSessionTimeout     time.Duration // idle_in_transaction_session_timeout
	LockTimeout                time.Duration // lock_timeout
	MaxOpenConns, MaxIdleConns int
}

func (config DBConfig) New(ctx context.Context, uri string, driverName string) (*sqlx.DB, error) {
	if driverName == DriverTxWrappedPostgres {
		// txdb uses the uri as a unique identifier for each transaction. Each ORM should
		// be encapsulated in its own transaction, and thus needs its own unique id.
		//
		// We can happily throw away the original uri here because if we are using txdb
		// it should have already been set at the point where we called [txdb.Register].
		return config.open(ctx, uuid.NewString(), driverName)
	}
	return config.openDB(ctx, uri, driverName)
}

func (config DBConfig) open(ctx context.Context, uri string, driverName string) (*sqlx.DB, error) {
	sqlDB, err := otelsql.Open(driverName, uri, otelOpts...)
	if err != nil {
		return nil, err
	}
	if _, err := sqlDB.ExecContext(ctx, config.initSQL()); err != nil {
		return nil, err
	}

	return config.newDB(ctx, sqlDB, driverName)
}

func (config DBConfig) openDB(ctx context.Context, uri string, driverName string) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("database: failed to parse config: %w", err)
	}

	initSQL := config.initSQL()
	connector := stdlib.GetConnector(*connConfig, stdlib.OptionAfterConnect(func(ctx context.Context, c *pgx.Conn) (err error) {
		_, err = c.Exec(ctx, initSQL)
		return
	}))

	sqlDB := otelsql.OpenDB(connector, otelOpts...)
	return config.newDB(ctx, sqlDB, driverName)
}

func (config DBConfig) initSQL() string {
	return fmt.Sprintf(`SET TIME ZONE 'UTC'; SET lock_timeout = %d; SET idle_in_transaction_session_timeout = %d; SET default_transaction_isolation = %q`,
		config.LockTimeout.Milliseconds(), config.IdleInTxSessionTimeout.Milliseconds(), defaultIsolation)
}

func (config DBConfig) newDB(ctx context.Context, sqldb *sql.DB, driverName string) (*sqlx.DB, error) {
	db := sqlx.NewDb(sqldb, driverName)
	db.MapperFunc(reflectx.CamelToSnakeASCII)
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	return db, disallowReplica(ctx, db)
}

func disallowReplica(ctx context.Context, db *sqlx.DB) error {
	var val string
	err := db.GetContext(ctx, &val, "SHOW session_replication_role")
	if err != nil {
		return err
	}

	if val == "replica" {
		return fmt.Errorf("invalid `session_replication_role`: %s. Refusing to connect to replica database. Writing to a replica will corrupt the database", val)
	}

	return nil
}

// Deprecated: Use Driver* constants
type DialectName string

const (
	// Deprecated: Use DriverPostgres
	Postgres DialectName = DriverPostgres
	// Deprecated: Use DriverTxWrappedPostgres
	TransactionWrappedPostgres DialectName = DriverTxWrappedPostgres
	// Deprecated: Use DriverTxWrappedPostgres DriverInMemoryPostgres
	InMemoryPostgres DialectName = DriverInMemoryPostgres
)
