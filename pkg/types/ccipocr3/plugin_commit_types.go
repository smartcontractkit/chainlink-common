package ccipocr3

import "math/big"

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
	MerkleRoots  []MerkleRoot `json:"merkleRoots"`
	PriceUpdates PriceUpdates `json:"priceUpdates"`
	// RMNSignatures are the ECDSA signatures from the RMN signing nodes on the RMNReport structure.
	// For more details see the contract here: https://github.com/smartcontractkit/chainlink/blob/7ba0f37134a618375542079ff1805fe2224d7916/contracts/src/v0.8/ccip/interfaces/IRMNV2.sol#L8-L12
	RMNSignatures []RMNECDSASignature `json:"rmnSignatures"`
	RmnRawVs      *big.Int            `json:"rmnRawVs"`
}

// IsEmpty returns true if the CommitPluginReport is empty
func (r CommitPluginReport) IsEmpty() bool {
	return len(r.MerkleRoots) == 0 &&
		len(r.PriceUpdates.TokenPriceUpdates) == 0 &&
		len(r.PriceUpdates.GasPriceUpdates) == 0 &&
		len(r.RMNSignatures) == 0
}

// Mirroring https://github.com/smartcontractkit/chainlink/blob/cd5c78959575f593b27fd83d8766086d0c678487/contracts/src/v0.8/ccip/libraries/Internal.sol#L356-L362
type MerkleRoot struct {
	SourceChainSelector ChainSelector `json:"sourceChainSelector"`
	OnRampAddress       Bytes         `json:"onRampAddress"`
	MinSeqNr            SeqNum        `json:"minSeqNr"`
	MaxSeqNr            SeqNum        `json:"maxSeqNr"`
	MerkleRoot          Bytes32       `json:"merkleRoot"`
}

func NewMerkleRootChain(
	sourceChainSel ChainSelector,
	onRampAddress Bytes,
	minSeqNr SeqNum,
	maxSeqNr SeqNum,
	merkleRoot Bytes32,
) MerkleRoot {
	return MerkleRoot{
		SourceChainSelector: sourceChainSel,
		OnRampAddress:       onRampAddress,
		MinSeqNr:            minSeqNr,
		MaxSeqNr:            maxSeqNr,
		MerkleRoot:          merkleRoot,
	}
}

type PriceUpdates struct {
	TokenPriceUpdates []TokenPrice    `json:"tokenPriceUpdates"`
	GasPriceUpdates   []GasPriceChain `json:"gasPriceUpdates"`
}
