package keystore_test

import (
	"bytes"
	"context"
	"fmt"
)

type StaticKeystoreConfig struct {
	Account string
	Encoded []byte
	Signed  []byte
}

type StaticKeystore struct {
	StaticKeystoreConfig
}

func (s StaticKeystore) Accounts(ctx context.Context) (accounts []string, err error) {
	return []string{string(s.Account)}, nil
}

func (s StaticKeystore) Sign(ctx context.Context, id string, data []byte) ([]byte, error) {
	if string(s.Account) != id {
		return nil, fmt.Errorf("expected id %q but got %q", s.Account, id)
	}
	if !bytes.Equal(s.Encoded, data) {
		return nil, fmt.Errorf("expected encoded data %x but got %x", s.Encoded, data)
	}
	return s.Signed, nil
}
