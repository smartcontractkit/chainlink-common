package sql

import (
	"github.com/99designs/keyring"
	"github.com/smartcontractkit/sqlx"
)

var (
	_ keyring.Keyring = (*backend)(nil)
)

type backend struct {
	db *sqlx.DB
}

type encryptedKey struct {
	Name  string `db:"key_name"`
	Value string `db:"key_value"`
}

func NewKeyring(db *sqlx.DB) *backend {
	// TODO: db migrations
	return &backend{db: db}
}

func (db *backend) Get(key string) (keyring.Item, error) {
	// TODO: decrypt after loading
	return keyring.Item{}, nil
}

func (db *backend) GetMetadata(_ string) (keyring.Metadata, error) {
	return keyring.Metadata{}, nil
}

func (db *backend) Set(item keyring.Item) error {
	// TODO: encrypt after saving
	return nil
}

func (db *backend) Remove(key string) error {
	return nil
}

func (db *backend) Keys() ([]string, error) {
	return nil, nil
}
