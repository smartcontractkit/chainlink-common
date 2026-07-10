package durableemitter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

const chipDurableEventsTable = "cre.chip_durable_events"

// PgDurableEventStore is a Postgres-backed implementation of DurableEventStore.
// Tests live in chainlink/core/services/durableemitter as they require DB migrations.
type PgDurableEventStore struct {
	ds sqlutil.DataSource
}

var (
	_ DurableEventStore    = (*PgDurableEventStore)(nil)
	_ DurableQueueObserver = (*PgDurableEventStore)(nil)
	_ BatchInserter        = (*PgDurableEventStore)(nil)
)

func NewPgDurableEventStore(ds sqlutil.DataSource) *PgDurableEventStore {
	return &PgDurableEventStore{ds: ds}
}

func (s *PgDurableEventStore) Insert(ctx context.Context, payload []byte) (int64, error) {
	const q = `INSERT INTO ` + chipDurableEventsTable + ` (payload) VALUES ($1) RETURNING id`
	var id int64
	if err := s.ds.GetContext(ctx, &id, q, payload); err != nil {
		return 0, fmt.Errorf("failed to insert chip durable event: %w", err)
	}
	return id, nil
}

func (s *PgDurableEventStore) InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error) {
	if len(payloads) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(payloads))
	args := make([]interface{}, len(payloads))
	for i, p := range payloads {
		placeholders[i] = fmt.Sprintf("($%d)", i+1)
		args[i] = p
	}
	q := fmt.Sprintf(
		"INSERT INTO %s (payload) VALUES %s RETURNING id",
		chipDurableEventsTable,
		strings.Join(placeholders, ","),
	)

	var ids []int64
	if err := s.ds.SelectContext(ctx, &ids, q, args...); err != nil {
		return nil, fmt.Errorf("failed to batch insert chip durable events: %w", err)
	}
	return ids, nil
}

func (s *PgDurableEventStore) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM ` + chipDurableEventsTable + ` WHERE id = $1`
	if _, err := s.ds.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("failed to delete chip durable event id=%d: %w", id, err)
	}
	return nil
}

func (s *PgDurableEventStore) BatchDelete(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	const q = `
DELETE FROM ` + chipDurableEventsTable + ` WHERE id IN (
    SELECT id FROM ` + chipDurableEventsTable + `
    WHERE id = ANY($1)
    FOR UPDATE SKIP LOCKED
)`
	// Delete durability is not required under delete-on-delivery: a delete commit
	// lost to a crash just leaves the row to be re-delivered (idempotent) or reclaimed
	// by the expiry loop. We can relax synchronous_commit for the delete transaction,
	// which removes the commit WAL-flush wait (IO:XactSync) which dominates at high
	// delete rates.
	var deleted int64
	err := sqlutil.TransactDataSource(ctx, s.ds, nil, func(tx sqlutil.DataSource) error {
		if _, err := tx.ExecContext(ctx, "SET LOCAL synchronous_commit = off"); err != nil {
			return fmt.Errorf("failed to set synchronous_commit=off: %w", err)
		}
		res, err := tx.ExecContext(ctx, q, pq.Array(ids))
		if err != nil {
			return err
		}
		deleted, _ = res.RowsAffected()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to batch delete delivered chip durable events: %w", err)
	}
	return deleted, nil
}

func (s *PgDurableEventStore) ListPending(ctx context.Context, createdBefore, afterCreatedAt time.Time, afterID int64, limit int) ([]DurableEvent, error) {
	// The (created_at, id) > ($2, $3) row-value comparison is the paging cursor:
	// the retransmit loop advances it forward through the backlog and wraps to a
	// zero cursor at the end, so a persistently-failing ("poison") event can't
	// monopolise the head of the list and block everything behind it.
	const q = `
SELECT id, payload, created_at
FROM ` + chipDurableEventsTable + `
WHERE delivered_at IS NULL
  AND created_at < $1
  AND (created_at, id) > ($2, $3)
ORDER BY created_at ASC, id ASC
LIMIT $4`

	type row struct {
		ID        int64     `db:"id"`
		Payload   []byte    `db:"payload"`
		CreatedAt time.Time `db:"created_at"`
	}

	var rows []row
	if err := s.ds.SelectContext(ctx, &rows, q, createdBefore, afterCreatedAt, afterID, limit); err != nil {
		return nil, fmt.Errorf("failed to list pending chip durable events: %w", err)
	}

	out := make([]DurableEvent, 0, len(rows))
	for _, r := range rows {
		out = append(out, DurableEvent{
			ID:        r.ID,
			Payload:   r.Payload,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}

func (s *PgDurableEventStore) DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error) {
	const q = `
WITH deleted AS (
    DELETE FROM ` + chipDurableEventsTable + `
    WHERE created_at <= now() - $1::interval
    RETURNING id
)
SELECT count(*) FROM deleted`

	var count int64
	if err := s.ds.GetContext(ctx, &count, q, ttl.String()); err != nil {
		return 0, fmt.Errorf("failed to delete expired chip durable events: %w", err)
	}
	return count, nil
}

type chipDurableQueueAgg struct {
	Cnt        int64      `db:"cnt"`
	Total      int64      `db:"total"`
	PayloadSum int64      `db:"payload_sum"`
	MinCreated *time.Time `db:"min_created"`
}

// ObserveDurableQueue implements DurableQueueObserver for queue depth / age gauges.
// cnt/payload_sum/min_created describe only the undelivered backlog, while total
// is the authoritative physical row count (incl. delivered-but-not-purged rows)
// used as the queue-depth gauge.
func (s *PgDurableEventStore) ObserveDurableQueue(ctx context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error) {
	const qAgg = `
SELECT
	count(*) FILTER (WHERE delivered_at IS NULL)::bigint AS cnt,
	count(*)::bigint AS total,
	coalesce(sum(octet_length(payload)) FILTER (WHERE delivered_at IS NULL), 0)::bigint AS payload_sum,
	min(created_at) FILTER (WHERE delivered_at IS NULL) AS min_created
FROM ` + chipDurableEventsTable

	var row chipDurableQueueAgg
	if err := s.ds.GetContext(ctx, &row, qAgg); err != nil {
		return DurableQueueStats{}, fmt.Errorf("durable queue aggregate: %w", err)
	}
	var st DurableQueueStats
	st.Depth = row.Cnt
	st.TotalRows = row.Total
	st.PayloadBytes = row.PayloadSum
	if row.MinCreated != nil {
		st.OldestPendingAge = time.Since(*row.MinCreated)
	}
	if eventTTL > 0 && nearExpiryLead > 0 && nearExpiryLead < eventTTL {
		ttlSec := int64(eventTTL.Round(time.Second) / time.Second)
		leadSec := int64(nearExpiryLead.Round(time.Second) / time.Second)
		const qNear = `
SELECT count(*)::bigint
FROM ` + chipDurableEventsTable + `
WHERE delivered_at IS NULL
  AND created_at >= now() - ($1::bigint * interval '1 second')
  AND created_at < now() - (($1::bigint - $2::bigint) * interval '1 second')`
		if err := s.ds.GetContext(ctx, &st.NearTTLCount, qNear, ttlSec, leadSec); err != nil {
			return DurableQueueStats{}, fmt.Errorf("durable queue near-ttl: %w", err)
		}
	}
	return st, nil
}
