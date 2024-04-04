package pg_test

import (
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func Test_disallowReplica(t *testing.T) {
	sqltest.SkipInMemory(t)
	ctx := tests.Context(t)

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
