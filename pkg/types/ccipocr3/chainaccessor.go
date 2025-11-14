package ccipocr3

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// ChainAccessor is responsible for direct access to the chain. In addition
// to direct access, this interface also translates onchain representations
// of data to the plugin representation.
type ChainAccessor interface {
	AllAccessors
	SourceAccessor
	DestinationAccessor
	USDCMessageReader
	PriceReader
}

// AllAccessors contains functionality that is available to all types of accessors (source and dest).
type AllAccessors interface {
	// GetContractAddress returns the contract address that is registered for the provided contract name
	// on the chain associated with the specific accessor implementation.
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
	// When called on a destination chain accessor:
	//   - destChainSelector should match the accessor's chain
	//   - sourceChainSelectors determines which source chain configs to fetch from the OffRamp
	//   - Returns ChainConfigSnapshot with OffRamp, FeeQuoter, RMNProxy, RMNRemote, CurseInfo configs
	//   - Returns map of SourceChainConfig for each requested source chain
	//
	// When called on a source chain accessor:
	//   - destChainSelector is used to fetch OnRamp dest chain specific configs
	//   - sourceChainSelectors will likely be empty and/or ignored by the accessor impl
	//   - Returns ChainConfigSnapshot with OnRamp, FeeQuoter, Router configs
	//   - Returns empty map for source chain configs
	//
	// Contract: Many
	// Confidence: Unconfirmed
	GetAllConfigsLegacy(
		ctx context.Context,
		destChainSelector ChainSelector,
		sourceChainSelectors []ChainSelector,
	) (ChainConfigSnapshot, map[ChainSelector]SourceChainConfig, error)

	// GetChainFeeComponents returns all fee components for the chain.
	//
	// Contract: N/A
	// Confidence: N/A
	GetChainFeeComponents(ctx context.Context) (ChainFeeComponents, error)

	// Sync binds a contract to the accessor implementation by contract name and address so the accessor
	// can use the contract addresses for subsequent calls. This is used to dynamically discover and bind
	// contracts (e.g., OnRamps, FeeQuoter, RMNRemote).
	// Returns an error if the bind operation failed.
	Sync(ctx context.Context, contractName string, contractAddress UnknownAddress) error
}

// DestinationAccessor contains all functions typically associated with the destination chain.
type DestinationAccessor interface {

	// CommitReportsGTETimestamp reads CommitReportAccepted events starting from a given timestamp.
	// The number of results is limited according to the limit parameter.
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

	// NextSeqNum reads the source chain config from the OffRamp to get the next expected
	// sequence number (MinSeqNr) for the given source chains.
	//
	// Access Type: Method(GetSourceChainConfig)
	// Contract: OffRamp
	// Confidence: Unconfirmed
	NextSeqNum(ctx context.Context, sources []ChainSelector) (map[ChainSelector]SeqNum, error)

	// Nonces returns nonces for all provided selector/address pairs. Addresses must be encoded
	// according to the source chain requirements by using the AddressCodec.
	//
	// Access Type: Method(GetInboundNonce)
	// Contract: NonceManager
	// Confidence: Unconfirmed
	Nonces(ctx context.Context, addresses map[ChainSelector][]UnknownEncodedAddress) (map[ChainSelector]map[string]uint64, error)

	// GetChainFeePriceUpdate returns the latest chain fee price update for the provided selectors. This queries
	// the FeeQuoter contract on the chain accociated with this accessor.
	//
	// Access Type: Method(GetChainFeePriceUpdate)
	// Contract: FeeQuoter
	// Confidence: Unconfirmed
	GetChainFeePriceUpdate(ctx context.Context, selectors []ChainSelector) (map[ChainSelector]TimestampedUnixBig, error)

	// GetLatestPriceSeqNr returns the latest price sequence number for the destination chain.
	// Not to be confused with the sequence number of the messages. This is the OCR sequence number.
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
	// for messages being sent to the provided destination.
	//
	// Access Type: Method(GetExpectedNextSequenceNumber)
	// Contract: OnRamp
	// Confidence: Unconfirmed
	GetExpectedNextSequenceNumber(ctx context.Context, dest ChainSelector) (SeqNum, error)

	// GetTokenPriceUSD looks up a token price in USD for the provided address.
	// Serves as a general price interface for fetching both LINK price and wrapped
	// native token price in USD.
	//
	// Access Type: Method(GetTokenPrice)
	// Contract: FeeQuoter
	// Confidence: Unconfirmed
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
	GetFeedPricesUSD(
		ctx context.Context,
		tokens []UnknownEncodedAddress,
		tokenInfo map[UnknownEncodedAddress]TokenInfo,
	) (TokenPriceMap, error)

	// GetFeeQuoterTokenUpdates returns the latest token prices from the FeeQuoter on the specified chain
	GetFeeQuoterTokenUpdates(
		ctx context.Context,
		tokensBytes []UnknownAddress,
	) (map[UnknownEncodedAddress]TimestampedUnixBig, error)
}

// ChainFeeComponents redeclares the ChainFeeComponents type from the chainlink-common/pkg/types to avoid
// a cyclic dependency caused by provider_ccip_ocr3.go importing this package.
type ChainFeeComponents struct {
	// The cost of executing transaction in the chain's EVM (or the L2 environment).
	ExecutionFee *big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityFee *big.Int
}
