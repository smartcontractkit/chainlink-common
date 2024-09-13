package ccipocr3

type RMNReport struct {
	ReportVersion               string // e.g. "RMN_V1_6_ANY2EVM_REPORT".
	DestChainID                 BigInt // If applies, a chain specific id, e.g. evm chain id otherwise empty.
	DestChainSelector           ChainSelector
	RmnRemoteContractAddress    []byte
	OfframpAddress              []byte
	RmnHomeContractConfigDigest Bytes32
	LaneUpdates                 []RMNLaneUpdate
}

type RMNLaneUpdate struct {
	SourceChainSelector ChainSelector
	OnRampAddress       []byte
	MinSeqNr            SeqNum
	MaxSeqNr            SeqNum
	MerkleRoot          Bytes32
}

type ECDSASignature struct {
	R Bytes
	S Bytes
}
