package securemint

import (
	"context"
	"math/big"
	"time"
)

// ExternalAdapter is an alias for por.ExternalAdapter
// TODO(gg): maybe move all por types that are used by the client (core node) to cl-common?

// ExternalAdapter is the component used by the secure mint plugin to request various secure mint related data points.
type ExternalAdapter interface {
	GetPayload(ctx context.Context, blocks Blocks) (ExternalAdapterPayload, error)
}

type BlockNumber uint64

type ChainSelector uint64

// Blocks contains the latest blocks per chain selector.
type Blocks map[ChainSelector]BlockNumber

// BlockMintablePair is a mintable amount at a certain block number.
type BlockMintablePair struct {
	Block    BlockNumber
	Mintable *big.Int
}

type Mintables map[ChainSelector]BlockMintablePair

// ReserveInfo is a reserve amount at a certain timestamp.
type ReserveInfo struct {
	ReserveAmount *big.Int
	Timestamp     time.Time
}

// ExternalAdapterPayload is the response from an EA containing various secure mint related data points.
type ExternalAdapterPayload struct {
	Mintables   Mintables   // The mintable amounts for each chain and its block.
	ReserveInfo ReserveInfo // The latest reserve amount and timestamp used to calculate the minting allowance above.

	LatestBlocks Blocks // The latest blocks for each chain.
}
