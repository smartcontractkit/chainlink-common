package ccipocr3

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// ChainFeeComponents redeclares the ChainFeeComponents type from the chainlink-common/pkg/types to avoid
// a cyclic dependency caused by provider_ccip_ocr3.go importing this package.
type ChainFeeComponents struct {
	// The cost of executing transaction in the chain's EVM (or the L2 environment).
	ExecutionFee *big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityFee *big.Int
}

// ChainAccessor for all direct chain access. The accessor is responsible for
// in addition to direct access to the chain, this interface also translates
// onchain representations of data to the plugin representation.
type ChainAccessor interface {
	AllAccessors
	SourceAccessor
	DestinationAccessor
	USDCMessageReader
	PriceReader
}

// AllAccessors contains functionality that is available to all types of accessors.
type AllAccessors interface {
	// GetContractAddress returns the contract address that is registered for the provided contract name and chain.
	// WARNING: This function will fail if the oracle does not support the requested chain.
	//
	// TODO(NONEVM-1865): do we want to mark this as deprecated in favor of Metadata()?
	GetContractAddress(contractName string) ([]byte, error)

	// GetAllConfig is the next iteration of GetAllConfigsLegacy(). Instead of returning a large snapshot
	// struct, it will ideally return a ChainConfigInterface that can be used to selectively fetch individual configs
	// depending on that particular chain's needs.
	/*
		GetAllConfig(
			ctx context.Context,
		) (ChainConfigInterface, error) // TBD...
	*/

	// GetAllConfigsLegacy returns a snapshot of all chain configurations for this chain using the legacy
	// config structs.
	//
	// destChainSelector is used to determine whether or not destination chain specific configs should be fetched.
	// sourceChainSelectors is used to determine which source chain configs should be fetched.
	//
	// This includes the following contracts:
	// - Router
	// - OffRamp
	// - OnRamp
	// - FeeQuoter
	// - RMNProxy
	// - RMNRemote
	// - CurseInfo
	//
	// Access Type: Method(many, see code)
	// Contract: Many
	// Confidence: Unconfirmed
	GetAllConfigsLegacy(
		ctx context.Context,
		destChainSelector ChainSelector,
		sourceChainSelectors []ChainSelector,
	) (ChainConfigSnapshot, map[ChainSelector]SourceChainConfig, error)

	// GetChainFeeComponents Returns all fee components for given chains if corresponding
	// chain writer is available.
	//
	// Access Type: ChainWriter
	// Contract: N/A
	// Confidence: N/A
	GetChainFeeComponents(ctx context.Context) (ChainFeeComponents, error)

	// Sync can be used to perform frequent syncing operations inside the reader implementation.
	// Returns an error if the sync operation failed.
	Sync(ctx context.Context, contractName string, contractAddress UnknownAddress) error
}

// DestinationAccessor contains all functions typically associated by the destination chain.
type DestinationAccessor interface {

	// CommitReportsGTETimestamp reads CommitReportAccepted events starting from a given timestamp.
	// The number of results are limited according to limit.
	//
	// Access Type: Event(CommitReportAccepted)
	// Contract: OffRamp
	// Confidence: Unconfirmed, Finalized
	CommitReportsGTETimestamp(
		ctx context.Context,
		ts time.Time,
		confidence primitives.ConfidenceLevel,
		limit int,
	) ([]CommitPluginReportWithMeta, error)

	// ExecutedMessages looks for ExecutionStateChanged events for each sequence
	// in a given range. The presence of these events indicates that an attempt to
	// execute the message has been made, which the system considers "executed".
	// A slice of all executed sequence numbers is returned.
	//
	// Access Type: Event(ExecutionStateChanged)
	// Contract: OffRamp
	// Confidence: Unconfirmed, Finalized
	ExecutedMessages(
		ctx context.Context,
		ranges map[ChainSelector][]SeqNumRange,
		confidence primitives.ConfidenceLevel,
	) (map[ChainSelector][]SeqNum, error)

	// NextSeqNum reads the source chain config to get the next expected
	// sequence number for the given source chains.
	//
	// Access Type: Method(NextSeqNum)
	// Contract: OffRamp
	// Confidence: Unconfirmed
	NextSeqNum(ctx context.Context, sources []ChainSelector) (map[ChainSelector]SeqNum, error)

	// Nonces for all provided selector/address pairs. Addresses must be encoded
	// according to the source chain requirements by using the AddressCodec.
	//
	// Access Type: Method(GetInboundNonce)
	// Contract: NonceManager
	// Confidence: Unconfirmed
	Nonces(ctx context.Context, addresses map[ChainSelector][]UnknownEncodedAddress) (map[ChainSelector]map[string]uint64, error)

	// GetChainFeePriceUpdate Gets latest chain fee price update for the provided chains.
	//
	// Access Type: Method(GetChainFeePriceUpdate)
	// Contract: FeeQuoter
	// Confidence: Unconfirmed
	GetChainFeePriceUpdate(ctx context.Context, selectors []ChainSelector) (map[ChainSelector]TimestampedUnixBig, error)

	// GetLatestPriceSeqNr returns the latest price sequence number for the destination chain.
	// Not to confuse with the sequence number of the messages. This is the OCR sequence number.
	//
	// Access Type: Method(GetLatestPriceSequenceNumber)
	// Contract: OffRamp
	// Confidence: Unconfirmed
	GetLatestPriceSeqNr(ctx context.Context) (SeqNum, error)
}

type SourceAccessor interface {
	// MsgsBetweenSeqNums returns all messages being sent to the provided dest
	// chain, between the provided sequence numbers. Messages are sorted ascending
	// based on their timestamp.
	//
	// Access Type: Event(CCIPMessageSent)
	// Contract: OnRamp
	// Confidence: Finalized
	MsgsBetweenSeqNums(ctx context.Context, dest ChainSelector, seqNumRange SeqNumRange) ([]Message, error)

	// LatestMessageTo returns the sequence number associated with the most
	// recent message being sent to a given destination.
	//
	// Access Type: Event(CCIPMessageSent)
	// Contract: OnRamp
	// Confidence: Finalized
	LatestMessageTo(ctx context.Context, dest ChainSelector) (SeqNum, error)

	// GetExpectedNextSequenceNumber returns the expected next sequence number
	// messages being sent to the provided destination.
	//
	// Access Type: Method(GetExpectedNextSequenceNumber)
	// Contract: OnRamp
	// Confidence: Unconfirmed
	GetExpectedNextSequenceNumber(ctx context.Context, dest ChainSelector) (SeqNum, error)

	// GetTokenPriceUSD looks up a token price in USD. The address value should
	// be retrieved from a configuration cache (i.e. ConfigPoller).
	//
	// Access Type: Method(GetTokenPrice)
	// Contract: FeeQuoter
	// Confidence: Unconfirmed
	//
	// Notes: This function is new and serves as a general price interface for
	//        both LinkPriceUSD and GetWrappedNativeTokenPriceUSD.
	//        See Design Doc (Combined Token Price Helper) for notes.
	GetTokenPriceUSD(ctx context.Context, address UnknownAddress) (TimestampedUnixBig, error)

	// GetFeeQuoterDestChainConfig returns the fee quoter destination chain config.
	//
	// Access Type: Method(GetDestChainConfig)
	// Contract: FeeQuoter
	// Confidence: Unconfirmed
	//
	// Notes: This is a new general purpose function needed to implement
	//        GetMedianDataAvailabilityGasConfig.
	GetFeeQuoterDestChainConfig(ctx context.Context, dest ChainSelector) (FeeQuoterDestChainConfig, error)
}

// USDCMessageReader retrieves each of the CCTPv1 MessageSent event created
// when a ccipSend is made with USDC token transfer. The events are created
// when the USDC Token pool calls the 3rd party MessageTransmitter contract.
type USDCMessageReader interface {
	MessagesByTokenID(ctx context.Context,
		source, dest ChainSelector,
		tokens map[MessageTokenID]RampTokenAmount,
	) (map[MessageTokenID]Bytes, error)
}

type PriceReader interface {
	// GetFeedPricesUSD returns the prices of the provided tokens in USD normalized to e18.
	//	1 USDC = 1.00 USD per full token, each full token is 1e6 units -> 1 * 1e18 * 1e18 / 1e6 = 1e30
	//	1 ETH = 2,000 USD per full token, each full token is 1e18 units -> 2000 * 1e18 * 1e18 / 1e18 = 2_000e18
	//	1 LINK = 5.00 USD per full token, each full token is 1e18 units -> 5 * 1e18 * 1e18 / 1e18 = 5e18
	// The order of the returned prices corresponds to the order of the provided tokens.
	GetFeedPricesUSD(ctx context.Context,
		tokens []UnknownEncodedAddress) (TokenPriceMap, error)

	// GetFeeQuoterTokenUpdates returns the latest token prices from the FeeQuoter on the specified chain
	GetFeeQuoterTokenUpdates(
		ctx context.Context,
		tokens []UnknownEncodedAddress,
		chain ChainSelector,
	) (map[UnknownEncodedAddress]TimestampedBig, error)
}
