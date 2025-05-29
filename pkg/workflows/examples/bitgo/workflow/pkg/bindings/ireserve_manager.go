package bindings

import (
	"math/big"

	evm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IReserverManager struct {
}

func NewIReserveManager(chainSelector uint, address []byte, defaultGasConfig evm.GasConfig) IReserverManager {
	return IReserverManager{}
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

func (ura UpdateReserveAccessor) WriteReport(UpdateReserveData UpdateReserveData) sdk.Promise[evm.WriteReportReply] {
	panic("")
}
