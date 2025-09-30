package evm_test

import (
	"math/big"
	"testing"

	"context"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestTxKey(t *testing.T) {
	storage := storage.NewMemoryStorage()
	ks, err := keystore.NewKeystore(storage, "test-password")
	require.NoError(t, err)
	myKey, err := evm.CreateTxKey(ks, "test-tx-key")
	require.NoError(t, err)
	resp, err := myKey.SignTx(context.Background(), evm.SignTxRequest{
		ChainID: big.NewInt(1),
		Tx:      &gethtypes.Transaction{},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Tx)
	// We still do the administration with the full name.
	_, err = ks.DeleteKeys(context.Background(), keystore.DeleteKeysRequest{
		Names: []string{myKey.FullName},
	})
	require.NoError(t, err)
}
