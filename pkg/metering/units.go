package metering

import (
	"fmt"

	chainselectors "github.com/smartcontractkit/chain-selectors"
	billing "github.com/smartcontractkit/chainlink-protos/billing/go"
)

var (
	PayloadUnit = unit{
		Name: billing.ResourceType_RESOURCE_TYPE_NETWORK.String(),
		Unit: billing.MeasurementUnit_MEASUREMENT_UNIT_BYTES.String()}


	ComputeUnit = unit{
		Name: billing.ResourceType_RESOURCE_TYPE_COMPUTE.String(),
		Unit: billing.MeasurementUnit_MEASUREMENT_UNIT_MILLISECONDS.String()}
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