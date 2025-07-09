package metering

import (
	"fmt"

	chainselectors "github.com/smartcontractkit/chain-selectors"
)

var (
	PayloadUnit = unit{Name: "payload", Unit: "bytes"}

	// ComputeUnit is an example.
	ComputeUnit = unit{Name: "compute", Unit: "ms"}
)

// unit provides exported Name and unit fields for
// capability devs to consume when implementing
// metering. Do not export.
type unit struct {
	// Name of the Metering Unit, i.e. payload, compute, storage
	Name string

	// Unit of the Metering Unit, i.e. bytes, seconds
	Unit string
}

func GasUnitForChain(chainID uint64) (string, error) {
	  // Getting ChainId based on ChainSelector
	  selector, err := chainselectors.SelectorFromChainId(chainID)
	  if err != nil {
		return "", err
	  }

	  return fmt.Sprintf("GAS.%d",selector), nil
}