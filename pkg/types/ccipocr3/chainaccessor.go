package ccipocr3

import (
	"context"
	"math/big"
	"sort"
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
	RMNAccessor
}

// AllAccessors contains functionality that is available to all types of accessors.
type AllAccessors interface {
	// GetContractAddress returns the contract address that is registered for the provided contract name and chain.
	// WARNING: This function will fail if the oracle does not support the requested chain.
	//
	// TODO(NONEVM-1865): do we want to mark this as deprecated in favor of Metadata()?
	GetContractAddress(contractName string) ([]byte, error)

	// GetAllConfig is the next iteration of GetAllConfigLegacySnapshot(). Instead of returning a large snapshot
	// struct, it will ideally return a ChainConfigInterface that can be used to selectively fetch individual configs
	// depending on that particular chain's needs.
	/*
		GetAllConfig(
			ctx context.Context,
		) (ChainConfigInterface, error) // TBD...
	*/

	// GetAllConfigLegacySnapshot returns the existing ChainConfigSnapshot struct. This function replaces
	// prepareBatchConfigRequests and is a temporary mechanism to support the mirgation to CAL until we can
	// build out GetAllConfig() above.
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
	GetAllConfigLegacySnapshot(ctx context.Context) (ChainConfigSnapshot, error)

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
	GetChainFeePriceUpdate(ctx context.Context, selectors []ChainSelector) map[ChainSelector]TimestampedBig

	// GetLatestPriceSeqNr returns the latest price sequence number for the destination chain.
	// Not to confuse with the sequence number of the messages. This is the OCR sequence number.
	//
	// Access Type: Method(GetLatestPriceSequenceNumber)
	// Contract: OffRamp
	// Confidence: Unconfirmed
	GetLatestPriceSeqNr(ctx context.Context) (uint64, error)
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

type RMNAccessor interface {
	// GetRMNCurseInfo returns rmn curse/pausing information about the provided chains
	// from the destination chain RMN remote contract. Caller should be able to access destination.
	GetRMNCurseInfo(ctx context.Context) (CurseInfo, error)
}

////////////////////////////////////////////////////////////////
// TODO: Find a better location for the types below this line //
//       For the purpose of designing these interfaces, the   //
//       location is not critical.                            //
////////////////////////////////////////////////////////////////

// Random types. These are defined here mainly to bring focus to types which should
// probably be removed or replaced.

type TimestampedBig struct {
	Timestamp time.Time `json:"timestamp"`
	Value     BigInt    `json:"value"`
}

// TimestampedUnixBig Maps to on-chain struct
// https://github.com/smartcontractkit/chainlink/blob/37f3132362ec90b0b1c12fb1b69b9c16c46b399d/contracts/src/v0.8/ccip/libraries/Internal.sol#L43-L47
//
//nolint:lll //url
type TimestampedUnixBig struct {
	// Value in uint224, can contain several packed fields
	Value *big.Int `json:"value"`
	// Timestamp in seconds since epoch of most recent update
	Timestamp uint32 `json:"timestamp"`
}

func NewTimestampedBig(value int64, timestamp time.Time) TimestampedBig {
	return TimestampedBig{
		Value:     BigInt{Int: big.NewInt(value)},
		Timestamp: timestamp,
	}
}

func TimeStampedBigFromUnix(input TimestampedUnixBig) TimestampedBig {
	return TimestampedBig{
		Value:     NewBigInt(input.Value),
		Timestamp: time.Unix(int64(input.Timestamp), 0),
	}
}

type CommitPluginReportWithMeta struct {
	Report    CommitPluginReport `json:"report"`
	Timestamp time.Time          `json:"timestamp"`
	BlockNum  uint64             `json:"blockNum"`
}

type CommitReportsByConfidenceLevel struct {
	Finalized   []CommitPluginReportWithMeta `json:"finalized"`
	Unfinalized []CommitPluginReportWithMeta `json:"unfinalized"`
}

// ContractAddresses is a map of contract names across all chain selectors and their address.
// Currently only one contract per chain per name is supported.
type ContractAddresses map[string]map[ChainSelector]UnknownAddress

// CurseInfo contains cursing information that are fetched from the rmn remote contract.
type CurseInfo struct {
	// CursedSourceChains contains the cursed source chains.
	CursedSourceChains map[ChainSelector]bool
	// CursedDestination indicates that the destination chain is cursed.
	CursedDestination bool
	// GlobalCurse indicates that all chains are cursed.
	GlobalCurse bool
}

func (ci CurseInfo) NonCursedSourceChains(inputChains []ChainSelector) []ChainSelector {
	if ci.GlobalCurse {
		return nil
	}

	sourceChains := make([]ChainSelector, 0, len(inputChains))
	for _, ch := range inputChains {
		if !ci.CursedSourceChains[ch] {
			sourceChains = append(sourceChains, ch)
		}
	}
	sort.Slice(sourceChains, func(i, j int) bool { return sourceChains[i] < sourceChains[j] })

	return sourceChains
}

// GlobalCurseSubject Defined as a const in RMNRemote.sol
// Docs of RMNRemote:
// An active curse on this subject will cause isCursed() and isCursed(bytes16) to return true. Use this subject
// for issues affecting all of CCIP chains, or pertaining to the chain that this contract is deployed on, instead of
// using the local chain selector as a subject.
var GlobalCurseSubject = [16]byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
}

// RemoteConfig contains the configuration fetched from the RMNRemote contract.
type RemoteConfig struct {
	ContractAddress UnknownAddress     `json:"contractAddress"`
	ConfigDigest    Bytes32            `json:"configDigest"`
	Signers         []RemoteSignerInfo `json:"signers"`
	// F defines the max number of faulty RMN nodes; F+1 signers are required to verify a report.
	FSign            uint64  `json:"fSign"` // previously: MinSigners
	ConfigVersion    uint32  `json:"configVersion"`
	RmnReportVersion Bytes32 `json:"rmnReportVersion"` // e.g., keccak256("RMN_V1_6_ANY2EVM_REPORT")
}

func (r RemoteConfig) IsEmpty() bool {
	// NOTE: contract address will always be present, since the code auto populates it
	return r.ConfigDigest == (Bytes32{}) &&
		len(r.Signers) == 0 &&
		r.FSign == 0 &&
		r.ConfigVersion == 0 &&
		r.RmnReportVersion == (Bytes32{})
}

// RemoteSignerInfo contains information about a signer from the RMNRemote contract.
type RemoteSignerInfo struct {
	// The signer's onchain address, used to verify report signature
	OnchainPublicKey UnknownAddress `json:"onchainPublicKey"`
	// The index of the node in the RMN config
	NodeIndex uint64 `json:"nodeIndex"`
}
