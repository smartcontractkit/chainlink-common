package securemint

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

// Report is the report that's created by the secure mint plugin.
// It contains a mintable token amount at a certain block number for a specific chain.
type Report struct {
	ConfigDigest types.ConfigDigest
	SeqNr        uint64
	Block        BlockNumber
	Mintable     *big.Int

	// The following fields might be useful in the future, but are not currently used
	// ReserveAmount *big.Int
	// ReserveTimestamp time.Time
}

// ExternalAdapter is the component used by the secure mint plugin to request various secure mint related data points.
type ExternalAdapter interface {
	GetPayload(ctx context.Context, blocks Blocks) (ExternalAdapterPayload, error)
}

// BlockNumber is a block number.
type BlockNumber uint64

// ChainSelector is a way of uniquely identifying a chain, see https://github.com/smartcontractkit/chain-selectors.
type ChainSelector uint64

// Blocks contains the latest blocks per chain.
type Blocks map[ChainSelector]BlockNumber

// BlockMintablePair is a mintable amount of a specific token at a certain block number.
type BlockMintablePair struct {
	Block    BlockNumber
	Mintable *big.Int
}

// Mintables contains the mintable amounts of a specific token per chain.
type Mintables map[ChainSelector]BlockMintablePair

// ReserveInfo is a reserve amount of a specific token at a certain timestamp.
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
