package ccipocr3

import (
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	libocrtypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

type CommitPluginConfig struct {
	// DestChain is the ccip destination chain configured for the commit plugin DON.
	DestChain ChainSelector `json:"destChain"`

	// PricedTokens is a list of tokens that we want to submit price updates for.
	PricedTokens []types.Account `json:"pricedTokens"`

	// TokenPricesObserver indicates that the node can observe token prices.
	TokenPricesObserver bool `json:"tokenPricesObserver"`

	// NewMsgScanBatchSize is the number of max new messages to scan, typically set to 256.
	NewMsgScanBatchSize int `json:"newMsgScanBatchSize"`
}

func (c CommitPluginConfig) Validate() error {
	if c.DestChain == ChainSelector(0) {
		return fmt.Errorf("destChain not set")
	}

	if len(c.PricedTokens) == 0 {
		return fmt.Errorf("priced tokens not set, at least one priced token is required")
	}

	if c.NewMsgScanBatchSize == 0 {
		return fmt.Errorf("newMsgScanBatchSize not set")
	}

	return nil
}

type ExecutePluginConfig struct {
	// DestChain is the ccip destination chain configured for the execute DON.
	DestChain ChainSelector `json:"destChain"`

	// ObserverInfo is a map of oracle IDs to ObserverInfo.
	ObserverInfo map[commontypes.OracleID]ObserverInfo `json:"observerInfo"`

	// MessageVisibilityInterval is the time interval for which the messages are visible by the plugin.
	MessageVisibilityInterval time.Duration `json:"messageVisibilityInterval"`

	// FChain defines the FChain value for each chain. FChain is used while forming consensus based on the observations.
	FChain map[ChainSelector]int `json:"fChain"`
}

type ObserverInfo struct {
	// Writer indicates that the node can contribute by sending reports to the destination chain.
	// Being a Writer guarantees that the node can also read from the destination chain.
	Writer bool `json:"writer"`

	// Reads define the chains that the current node can read from.
	Reads []ChainSelector `json:"reads"`
}

// ChainConfig will live on the home chain and will be used to update chain configuration like F value and supported nodes dynamically.
type ChainConfig struct {
	// FChain defines the FChain value for the chain. FChain is used while forming consensus based on the observations.
	FChain int `json:"fChain"`
	// SupportedNodes is a map of PeerIDs to SupportedChains.
	SupportedNodes mapset.Set[libocrtypes.PeerID] `json:"supportedNodes"`
}
