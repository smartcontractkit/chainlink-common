package pgstore

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

var _ keystore.Storage = &Storage{}

// Storage implements Storage using a Postgres database
type Storage struct {
	ds   sqlutil.DataSource
	name string
}

func NewStorage(ds sqlutil.DataSource, name string) *Storage {
	return &Storage{
		ds:   ds,
		name: name,
	}
}

func (p *Storage) GetEncryptedKeystore(ctx context.Context) (res []byte, err error) {
	err = p.ds.GetContext(ctx, &res, `SELECT encrypted_data FROM encrypted_keystore WHERE name = $1`, p.name)
	return
}

func (p *Storage) PutEncryptedKeystore(ctx context.Context, encryptedKeystore []byte) (err error) {
	_, err = p.ds.ExecContext(ctx,
		`INSERT INTO encrypted_keystore (name, encrypted_data, updated_at) VALUES ($1, $2, NOW())
				ON CONFLICT(name)
				DO UPDATE SET encrypted_data = EXCLUDED.encrypted_data, updated_at = EXCLUDED.updated_at`,
		p.name, encryptedKeystore)
	return
}
