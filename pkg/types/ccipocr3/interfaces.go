package ccipocr3

import (
	"context"
	"math/big"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/services"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	libocrtypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

type CCIPReader interface {
	// CommitReportsGTETimestamp reads the requested chain starting at a given timestamp
	// and finds all ReportAccepted up to the provided limit.
	CommitReportsGTETimestamp(ctx context.Context, dest ChainSelector, ts time.Time, limit int) ([]CommitPluginReportWithMeta, error)

	// ExecutedMessageRanges reads the destination chain and finds which messages are executed.
	// A slice of sequence number ranges is returned to express which messages are executed.
	ExecutedMessageRanges(ctx context.Context, source, dest ChainSelector, seqNumRange SeqNumRange) ([]SeqNumRange, error)

	// MsgsBetweenSeqNums reads the provided chains.
	// Finds and returns ccip messages submitted between the provided sequence numbers.
	// Messages are sorted ascending based on their timestamp and limited up to the provided limit.
	MsgsBetweenSeqNums(ctx context.Context, chain ChainSelector, seqNumRange SeqNumRange) ([]CCIPMsg, error)

	// NextSeqNum reads the destination chain.
	// Returns the next expected sequence number for each one of the provided chains.
	// TODO: if destination was a parameter, this could be a capability reused across plugin instances.
	NextSeqNum(ctx context.Context, chains []ChainSelector) (seqNum []SeqNum, err error)

	// GasPrices reads the provided chains gas prices.
	GasPrices(ctx context.Context, chains []ChainSelector) ([]BigInt, error)

	// Close closes any open resources.
	Close(ctx context.Context) error
}

type TokenPricesReader interface {
	// GetTokenPricesUSD returns the prices of the provided tokens in USD.
	// The order of the returned prices corresponds to the order of the provided tokens.
	GetTokenPricesUSD(ctx context.Context, tokens []types.Account) ([]*big.Int, error)
}

type CommitPluginCodec interface {
	Encode(context.Context, CommitPluginReport) ([]byte, error)
	Decode(context.Context, []byte) (CommitPluginReport, error)
}

type ExecutePluginCodec interface {
	Encode(context.Context, ExecutePluginReport) ([]byte, error)
	Decode(context.Context, []byte) (ExecutePluginReport, error)
}

type MessageHasher interface {
	Hash(context.Context, CCIPMsg) (Bytes32, error)
}

type HomeChainPoller interface {
	GetChainConfig(chainSelector ChainSelector) (ChainConfig, error)
	GetAllChainConfigs() (map[ChainSelector]ChainConfig, error)
	// GetSupportedChainsForPeer Gets all chain selectors that the peerID can read/write from/to
	GetSupportedChainsForPeer(id libocrtypes.PeerID) (mapset.Set[ChainSelector], error)
	// GetKnownCCIPChains Gets all chain selectors that are known to CCIP
	GetKnownCCIPChains() (mapset.Set[ChainSelector], error)
	// GetFChain Gets the FChain value for each chain
	GetFChain() (map[ChainSelector]int, error)
	services.Service
}
