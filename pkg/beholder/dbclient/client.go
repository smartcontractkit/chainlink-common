package dbclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type Client struct {
	db *sqlx.DB
}

var _ beholder.Emitter = (*Client)(nil)

func NewClient(db *sqlx.DB) *Client {
	return &Client{
		db: db,
	}
}

func (client *Client) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	_, err := client.BatchEmit(ctx, []beholder.Message{
		beholder.NewMessage(body, attrKVs...),
	})
	return err
}

func (client *Client) BatchEmit(ctx context.Context, messages []beholder.Message, options ...beholder.BatchEmitOption) ([]*chipingress.PublishResult, error) {
	tx, err := client.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() // no-op if transaction has been committed

	for _, msg := range messages {
		attrs, err := json.Marshal(msg.Attrs)
		if err != nil {
			return nil, fmt.Errorf("marshaling attributes: %w", err)
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO queue (payload, attributes) VALUES ($1, $2)", msg.Body, attrs)
		if err != nil {
			return nil, fmt.Errorf("inserting message: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return nil, nil
}

func (client *Client) Close() error {
	return nil
}
