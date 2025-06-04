package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IReserverManager struct {
	client  evmcappb.Client
	Structs Structs
}

type Structs struct {
	UpdateReserves UpdateReserves
}

func NewIReserveManager(chainSelector uint, address []byte, defaultGasConfig *evm.GasConfig) IReserverManager {
	c := &evmcappb.Client{}
	return IReserverManager{client: *c}
}

type UpdateReserves struct {
	TotalMinted  big.Int
	TotalReserve big.Int
}

func (ur UpdateReserves) WriteReport(UpdateReserves UpdateReserves) sdk.Promise[evm.WriteReportReply] {
	panic("")
}
