package core

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

const PluginSecureMintName = "securemint"

type PluginSecureMint interface {
	services.Service
	// NewSecureMintFactory returns a ReportingPluginFactory for the secure mint plugin.
	NewSecureMintFactory(ctx context.Context, lggr logger.Logger, externalAdapter ExternalAdapter) (ReportingPluginFactory[ChainSelector], error)
	// TODO(gg): is it correct that NewSecureMintFactory gets a param with a type specified in the por repo? --> maybe we have to alias it in cl-common?
}

// ExternalAdapter is an alias for por.ExternalAdapter
// TODO(gg): maybe move all por types that are used by the client (core node) to cl-common?
// TODO(gg): Otherwise add godoc to all below types.
// TODO(gg): should we put this under core/secure_mint? Or rename to SecureMintExternalAdapter? `core.ExternalAdapter` seems too broad for what it's doing.
type ExternalAdapter interface {
	GetPayload(ctx context.Context, blocks Blocks) (ExternalAdapterPayload, error)
}

type BlockNumber uint64

type ChainSelector uint64

type Blocks map[ChainSelector]BlockNumber

type BlockMintablePair struct {
	Block    BlockNumber
	Mintable *big.Int
}

type Mintables map[ChainSelector]BlockMintablePair

type ReserveInfo struct {
	ReserveAmount *big.Int
	Timestamp     time.Time
}

type ExternalAdapterPayload struct {
	Mintables   Mintables   // The mintable amounts for each chain and its block.
	ReserveInfo ReserveInfo // The latest reserve amount and timestamp used to calculate the minting allowance above.

	LatestBlocks Blocks // The latest blocks for each chain.
}

// ReportingPluginFactory wraps ocr3types.ReportingPluginFactory[RI] to add a Service to it.
type ReportingPluginFactory[RI any] interface {
	services.Service
	ocr3types.ReportingPluginFactory[RI]
}
