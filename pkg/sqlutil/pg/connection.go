package pg

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/scylladb/go-reflectx"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	// need to make sure pgx driver is registered before opening connection
	_ "github.com/jackc/pgx/v4/stdlib"
)

// NOTE: This is the default level in Postgres anyway, we just make it
// explicit here
const defaultIsolation = sql.LevelReadCommitted

// DialectName is a compiler enforced type used that maps to database dialect names
type DialectName string

const (
	// DialectPostgres represents the postgres dialect.
	DialectPostgres DialectName = "pgx"
	// DialectTransactionWrappedPostgres is useful for tests.
	// When the connection is opened, it starts a transaction and all
	// operations performed on the DB will be within that transaction.
	DialectTransactionWrappedPostgres DialectName = "txdb"
)

type ConnectionConfig struct {
	DefaultIdleInTxSessionTimeout time.Duration
	DefaultLockTimeout            time.Duration
	MaxOpenConns                  int
	MaxIdleConns                  int
}

func (config ConnectionConfig) NewDB(uri string, dialect DialectName) (db *sqlx.DB, err error) {
	if dialect == DialectTransactionWrappedPostgres {
		// Dbtx uses the uri as a unique identifier for each transaction. Each ORM
		// should be encapsulated in it's own transaction, and thus needs its own
		// unique id.
		//
		// We can happily throw away the original uri here because if we are using
		// txdb it should have already been set at the point where we called
		// txdb.Register
		uri = uuid.New().String()
	}

	// Initialize sql/sqlx
	sqldb, err := otelsql.Open(string(dialect), uri,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
			OmitConnQuery:        false,
		}),
	)
	if err != nil {
		return nil, err
	}
	db = sqlx.NewDb(sqldb, string(dialect))
	db.MapperFunc(reflectx.CamelToSnakeASCII)

	// Set default connection options
	lockTimeout := config.DefaultLockTimeout.Milliseconds()
	idleInTxSessionTimeout := config.DefaultIdleInTxSessionTimeout.Milliseconds()
	stmt := fmt.Sprintf(`SET TIME ZONE 'UTC'; SET lock_timeout = %d; SET idle_in_transaction_session_timeout = %d; SET default_transaction_isolation = %q`,
		lockTimeout, idleInTxSessionTimeout, defaultIsolation.String())
	if _, err = db.Exec(stmt); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	return db, disallowReplica(db)
}

func disallowReplica(db *sqlx.DB) error {
	var val string
	err := db.Get(&val, "SHOW session_replication_role")
	if err != nil {
		return err
	}

	if val == "replica" {
		return fmt.Errorf("invalid `session_replication_role`: %s. Refusing to connect to replica database. Writing to a replica will corrupt the database", val)
	}

	return nil
}
