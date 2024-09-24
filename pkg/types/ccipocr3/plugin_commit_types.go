package ccipocr3

// CommitPluginReport contains the necessary information to commit CCIP
// messages from potentially many source chains, to a single destination chain.
//
// It must consist of either:
//
//  1. a non-empty MerkleRoots array, or
//  2. a non-empty PriceUpdates array
//
// If neither of the above is provided the report is considered empty and should
// not be transmitted on-chain.
//
// In the event the MerkleRoots array is non-empty, it may also contain
// RMNSignatures, if RMN is configured for some lanes involved in the commitment.
// A report with RMN signatures but without merkle roots is invalid.
type CommitPluginReport struct {
	MerkleRoots  []MerkleRootChain `json:"merkleRoots"`
	PriceUpdates PriceUpdates      `json:"priceUpdates"`
	// RMNSignatures are the ECDSA signatures from the RMN signing nodes on the RMNReport structure.
	// For more details see the contract here: https://github.com/smartcontractkit/chainlink/blob/7ba0f37134a618375542079ff1805fe2224d7916/contracts/src/v0.8/ccip/interfaces/IRMNV2.sol#L8-L12
	RMNSignatures []RMNECDSASignature `json:"rmnSignatures"`
}

// IsEmpty returns true if the CommitPluginReport is empty
func (r CommitPluginReport) IsEmpty() bool {
	return len(r.MerkleRoots) == 0 &&
		len(r.PriceUpdates.TokenPriceUpdates) == 0 &&
		len(r.PriceUpdates.GasPriceUpdates) == 0 &&
		len(r.RMNSignatures) == 0
}

type MerkleRootChain struct {
	ChainSel     ChainSelector `json:"chain"`
	SeqNumsRange SeqNumRange   `json:"seqNumsRange"`
	MerkleRoot   Bytes32       `json:"merkleRoot"`
}

func NewMerkleRootChain(
	chainSel ChainSelector,
	seqNumsRange SeqNumRange,
	merkleRoot Bytes32,
) MerkleRootChain {
	return MerkleRootChain{
		ChainSel:     chainSel,
		SeqNumsRange: seqNumsRange,
		MerkleRoot:   merkleRoot,
	}
}

type PriceUpdates struct {
	TokenPriceUpdates []TokenPrice    `json:"tokenPriceUpdates"`
	GasPriceUpdates   []GasPriceChain `json:"gasPriceUpdates"`
}
