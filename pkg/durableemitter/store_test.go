package durableemitter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

// TestPgDurableEventStore_ListPendingByteBudget runs against CL_DATABASE_URL and
// skips without it.
func TestPgDurableEventStore_ListPendingByteBudget(t *testing.T) {
	sqltest.SkipInMemory(t)
	db := sqltest.NewDB(t, sqltest.TestURL(t))
	ctx := t.Context()
	for _, q := range []string{
		`CREATE SCHEMA IF NOT EXISTS cre`,
		`CREATE TABLE IF NOT EXISTS cre.chip_durable_events (
			id BIGSERIAL PRIMARY KEY,
			payload BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			delivered_at TIMESTAMPTZ NULL)`,
		`DELETE FROM cre.chip_durable_events`,
	} {
		_, err := db.ExecContext(ctx, q)
		require.NoError(t, err)
	}

	store := NewPgDurableEventStore(db)
	sizes := []int{10, 10, 10, 50, 10}
	payloads := make([][]byte, len(sizes))
	for i, n := range sizes {
		payloads[i] = make([]byte, n)
	}
	ids, err := store.InsertBatch(ctx, payloads)
	require.NoError(t, err)
	before := time.Now().Add(time.Hour)

	page, err := store.ListPending(ctx, before, time.Time{}, 0, 10, 25)
	require.NoError(t, err)
	assert.Equal(t, ids[:2], eventIDs(page), "stops before the row that crosses the budget")
	assert.Len(t, page[0].Payload, 10)

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 10, 30)
	require.NoError(t, err)
	assert.Equal(t, ids[:3], eventIDs(page), "a row landing exactly on the budget is kept")

	last := page[len(page)-1]
	page, err = store.ListPending(ctx, before, last.CreatedAt, last.ID, 10, 25)
	require.NoError(t, err)
	assert.Equal(t, ids[3:4], eventIDs(page), "a single row over the budget is still returned")

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, ids, eventIDs(page), "zero budget is unbounded")

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 2, 1000)
	require.NoError(t, err)
	assert.Equal(t, ids[:2], eventIDs(page), "row limit still applies")

	page, err = store.ListPending(ctx, time.Now().Add(-time.Hour), time.Time{}, 0, 10, 25)
	require.NoError(t, err)
	assert.Empty(t, page)
}
