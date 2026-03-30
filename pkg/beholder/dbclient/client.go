package dbclient

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

var (
	queueTable         = pgx.Identifier{"queue"}
	queueInsertColumns = []string{"payload", "attributes"}
)

type Client struct {
	db *pgxpool.Pool
}

func NewClient(db *pgxpool.Pool) *Client {
	return &Client{
		db: db,
	}
}

func (client *Client) Emit(ctx context.Context, messages []beholder.Message, options ...beholder.BatchEmitOption) error {
	tx, err := client.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if transaction has been commited

	rows := make([][]any, len(messages))
	for i, msg := range messages {
		attrs := make([]any, 0, len(msg.Attrs)*2)
		for k, v := range msg.Attrs {
			attrs = append(attrs, k, v)
		}
		rows[i] = []any{msg.Body, msg.Attrs}
	}

	n, err := tx.CopyFrom(ctx, queueTable, queueInsertColumns, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copying batch: %w", err)
	}
	if n != int64(len(rows)) {
		return fmt.Errorf("copied %d rows, expected %d", n, len(rows))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (client *Client) Close() error {
	return nil
}
