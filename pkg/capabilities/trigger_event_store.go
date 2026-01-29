package capabilities

/*
import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type pgTriggerEventStore struct {
	db        *sql.DB
	tableName string
}

func NewPostgresEventStore(ctx context.Context, db *sql.DB, tableName string) (EventStore, error) {
	s := &pgTriggerEventStore{db: db, tableName: tableName}
	if err := s.ensureSchema(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *pgTriggerEventStore) ensureSchema(ctx context.Context) error {
	ddl := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  trigger_id   TEXT NOT NULL,
  event_id     TEXT NOT NULL,
  any_type_url TEXT NOT NULL,
  payload      BYTEA NOT NULL,
  first_at     TIMESTAMPTZ NOT NULL,
  last_sent_at TIMESTAMPTZ,
  attempts     INT NOT NULL DEFAULT 0,
  PRIMARY KEY (trigger_id, workflow_id, event_id)
);
CREATE INDEX IF NOT EXISTS %s_firstat_idx ON %s (first_at);
`, s.tableName, s.tableName, s.tableName)

	_, err := s.db.ExecContext(ctx, ddl)
	return err
}

func (s *pgTriggerEventStore) Insert(ctx context.Context, rec PendingEvent) error {
	q := fmt.Sprintf(`
INSERT INTO %s (trigger_id, event_id, any_type_url, payload, first_at, last_sent_at, attempts)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (trigger_id, event_id)
DO UPDATE SET
  any_type_url = EXCLUDED.any_type_url,
  payload      = EXCLUDED.payload,
  first_at     = LEAST(%s.first_at, EXCLUDED.first_at),
  last_sent_at = EXCLUDED.last_sent_at,
  attempts     = EXCLUDED.attempts;`, s.tableName, s.tableName)

	_, err := s.db.ExecContext(
		ctx, q,
		rec.TriggerId, rec.EventId,
		rec.AnyTypeURL, rec.Payload,
		rec.FirstAt, nullTime(rec.LastSentAt), rec.Attempts,
	)
	return err
}

func (s *pgTriggerEventStore) Delete(ctx context.Context, triggerId, eventId string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE trigger_id=$1 AND event_id=$2;`, s.tableName)
	_, err := s.db.ExecContext(ctx, q, triggerId, eventId)
	return err
}

func (s *pgTriggerEventStore) List(ctx context.Context) ([]PendingEvent, error) {
	q := fmt.Sprintf(`
SELECT trigger_id, event_id, any_type_url, payload, first_at, last_sent_at, attempts
FROM %s
ORDER BY first_at ASC;`, s.tableName)

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PendingEvent
	for rows.Next() {
		var rec PendingEvent
		var lastSent sql.NullTime
		if err := rows.Scan(
			&rec.TriggerId, &rec.EventId,
			&rec.AnyTypeURL, &rec.Payload,
			&rec.FirstAt, &lastSent, &rec.Attempts,
		); err != nil {
			return nil, err
		}
		if lastSent.Valid {
			rec.LastSentAt = lastSent.Time
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}
*/
