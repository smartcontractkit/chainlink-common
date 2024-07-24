package ccipocr3

type CommitPluginReport struct {
	MerkleRoots  []MerkleRootChain `json:"merkleRoots"`
	PriceUpdates PriceUpdates      `json:"priceUpdates"`
}

func NewCommitPluginReport(merkleRoots []MerkleRootChain, tokenPriceUpdates []TokenPrice, gasPriceUpdate []GasPriceChain) CommitPluginReport {
	return CommitPluginReport{
		MerkleRoots:  merkleRoots,
		PriceUpdates: PriceUpdates{TokenPriceUpdates: tokenPriceUpdates, GasPriceUpdates: gasPriceUpdate},
	}
}

// IsEmpty returns true if the CommitPluginReport is empty
func (r CommitPluginReport) IsEmpty() bool {
	return len(r.MerkleRoots) == 0 &&
		len(r.PriceUpdates.TokenPriceUpdates) == 0 &&
		len(r.PriceUpdates.GasPriceUpdates) == 0
}

type MerkleRootChain struct {
	SourceChainSelector ChainSelector
	Interval            SeqNumRange
	MerkleRoot          Bytes32
}

func NewMerkleRootChain(
	chainSel ChainSelector,
	seqNumsRange SeqNumRange,
	merkleRoot Bytes32,
) MerkleRootChain {
	return MerkleRootChain{
		SourceChainSelector: chainSel,
		Interval:            seqNumsRange,
		MerkleRoot:          merkleRoot,
	}
}

type PriceUpdates struct {
	TokenPriceUpdates []TokenPrice    `json:"tokenPriceUpdates"`
	GasPriceUpdates   []GasPriceChain `json:"gasPriceUpdates"`
}
