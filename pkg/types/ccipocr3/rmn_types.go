package ccipocr3

import "sort"

// RMNReport is the payload that is signed by the RMN nodes, transmitted and verified onchain.
type RMNReport struct {
	ReportVersionDigest         Bytes32 // e.g. keccak256("RMN_V1_6_ANY2EVM_REPORT")
	DestChainID                 BigInt  // If applies, a chain specific id, e.g. evm chain id otherwise empty.
	DestChainSelector           ChainSelector
	RmnRemoteContractAddress    UnknownAddress
	OfframpAddress              UnknownAddress
	RmnHomeContractConfigDigest Bytes32
	LaneUpdates                 []RMNLaneUpdate
}

func NewRMNReport(
	reportVersionDigest Bytes32,
	destChainID BigInt,
	destChainSelector ChainSelector,
	rmnRemoteContractAddress UnknownAddress,
	offRampAddress UnknownAddress,
	rmnHomeContractConfigDigest Bytes32,
	laneUpdates []RMNLaneUpdate,
) RMNReport {
	return RMNReport{
		ReportVersionDigest:         reportVersionDigest,
		DestChainID:                 destChainID,
		DestChainSelector:           destChainSelector,
		RmnRemoteContractAddress:    rmnRemoteContractAddress,
		OfframpAddress:              offRampAddress,
		RmnHomeContractConfigDigest: rmnHomeContractConfigDigest,
		LaneUpdates:                 laneUpdates,
	}
}

// RMNLaneUpdate represents an interval that has been observed by an RMN node.
// It is part of the payload that is signed and transmitted onchain.
type RMNLaneUpdate struct {
	SourceChainSelector ChainSelector
	OnRampAddress       UnknownAddress // (for EVM should be abi-encoded)
	MinSeqNr            SeqNum
	MaxSeqNr            SeqNum
	MerkleRoot          Bytes32
}

// RMNECDSASignature represents the signature provided by RMN on the RMNReport structure.
// The V value of the signature is included in the top-level commit report as part of a
// bitmap.
type RMNECDSASignature struct {
	R Bytes32 `json:"r"`
	S Bytes32 `json:"s"`
}

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
