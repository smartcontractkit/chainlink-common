package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IERC20 struct {
}

func NewIERC20(chainSelector uint, address []byte, defaultGasConfig *evm.GasConfig) IERC20 {
	return IERC20{}
}

func (IERC20) TotalSupplyAccessor() TotalSupplyAccessor {
	return TotalSupplyAccessor{}
}

type TotalSupplyAccessor struct {
}

func (tsa TotalSupplyAccessor) TotalSupply() sdk.Promise[*big.Int] {
	panic("")
}
