package sql

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/assert"
)

func TestNewKeystore(t *testing.T) {
	kr := NewKeyring(nil)
	opts := func(options *keyring.Options) {
		// enable some options
	}

	// TODO: add custom cdc (2. param)
	ks := keyring.NewKeystore(kr, nil, opts)
	assert.NotEmpty(t, ks)
}
