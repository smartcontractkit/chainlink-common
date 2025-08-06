package pg_test

import (
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

func Test_disallowReplica(t *testing.T) {
	sqltest.SkipInMemory(t)
	ctx := t.Context()

	db := sqltest.NewDB(t, sqltest.TestURL(t))
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err := db.Exec("SET session_replication_role= 'origin'")
	require.NoError(t, err)
	err = pg.DisallowReplica(ctx, db)
	require.NoError(t, err)

	_, err = db.Exec("SET session_replication_role= 'replica'")
	require.NoError(t, err)
	err = pg.DisallowReplica(ctx, db)
	require.Error(t, err, "replica role should be disallowed")

	_, err = db.Exec("SET session_replication_role= 'not_valid_role'")
	require.Error(t, err)
}

func TestDBConfig_WithTracing(t *testing.T) {
	t.Parallel()
	cfg := pg.DBConfig{}
	require.False(t, cfg.EnableTracing)

	cfgWithTrace := cfg.WithTracing(true)
	require.True(t, cfgWithTrace.EnableTracing)
	require.False(t, cfg.EnableTracing) // ensure original is not modified

	cfgNoTrace := cfgWithTrace.WithTracing(false)
	require.False(t, cfgNoTrace.EnableTracing)
	require.True(t, cfgWithTrace.EnableTracing) // ensure original is not modified
}
