package types

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type ChainReceipt[TX_HASH, BLOCK_HASH types.Hashable] interface {
	GetStatus() uint64
	GetTxHash() TX_HASH
	GetBlockNumber() *big.Int
	IsZero() bool
	IsUnmined() bool
	GetFeeUsed() uint64
	GetTransactionIndex() uint
	GetBlockHash() BLOCK_HASH
}
