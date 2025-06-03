package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IReserverManager struct{ client *evm.Client }

func NewIReserveManager(chainSelector uint, address []byte, defaultGasConfig evm.GasConfig) IReserverManager {
	c := &evm.Client{}
	return IReserverManager{client: c}
}

type UpdateReserveData struct {
	TotalMinted  big.Int
	TotalReserve big.Int
}

type UpdateReserveAccessor struct {
}

func (irm IReserverManager) UpdateReserveAccessor() UpdateReserveAccessor {
	return UpdateReserveAccessor{}
}

func (ura UpdateReserveAccessor) WriteReport(UpdateReserveData UpdateReserveData) sdk.Promise[evm] {
	panic("")
}
