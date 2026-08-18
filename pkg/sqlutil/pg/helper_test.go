package pg

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func DisallowReplica(ctx context.Context, db *sqlx.DB) error {
	return disallowReplica(ctx, db)
}
