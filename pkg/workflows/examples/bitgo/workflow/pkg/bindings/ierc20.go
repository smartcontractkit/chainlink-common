package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type IERC20 struct {
	Methods   Methods
	ContractInputs ContractInputs
}

func NewIERC20(contractInputs ContractInputs) IERC20 {
	ierc20 := IERC20{ContractInputs: contractInputs}
	ierc20.Methods = Methods{
		TotalSupply: TotalSupply{
			IERC20: &ierc20,
		},
	}

	return ierc20
}

type Methods struct {
	TotalSupply
}

type TotalSupply struct {
	IERC20 *IERC20
}

func (ts TotalSupply) Call(runtime sdk.Runtime, options *ReadOptions) sdk.Promise[*big.Int] {
	callContractReplyPromise := ts.IERC20.ContractInputs.EVM.CallContract(runtime, &evm.CallContractRequest{
		Call: &evm.CallMsg{
			To:   ts.IERC20.ContractInputs.Address,
			Data: []byte{},
		},
	})

	return sdk.Then(callContractReplyPromise, func(callContractReply *evm.CallContractReply) (*big.Int, error) {
		return big.NewInt(int64(10)), nil
	})
}
