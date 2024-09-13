package ccipocr3

type RMNReport struct {
	ReportVersion               string // e.g. "RMN_V1_6_ANY2EVM_REPORT".
	DestChainID                 BigInt // If applies, a chain specific id, e.g. evm chain id otherwise empty.
	DestChainSelector           ChainSelector
	RmnRemoteContractAddress    Bytes
	OfframpAddress              Bytes
	RmnHomeContractConfigDigest Bytes32
	LaneUpdates                 []RMNLaneUpdate
}

type RMNLaneUpdate struct {
	SourceChainSelector ChainSelector
	OnRampAddress       Bytes // (for EVM should be abi-encoded)
	MinSeqNr            SeqNum
	MaxSeqNr            SeqNum
	MerkleRoot          Bytes32
}

type RMNECDSASignature struct {
	R Bytes32 `json:"r"`
	S Bytes32 `json:"s"`
}
