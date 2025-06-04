package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IERC20 struct {
	Methods Methods
}

func NewIERC20(chainSelector uint, address []byte, defaultGasConfig *evm.GasConfig) IERC20 {
	return IERC20{}
}

type Methods struct {
   TotalSupply
}

type TotalSupply struct {
	
}

type ReadOptions struct {
	BlockNumber *big.Int
}

func (ts TotalSupply) Call(runtime sdk.Runtime, options *ReadOptions) sdk.Promise[*big.Int]{
	panic("not implemented")
}

