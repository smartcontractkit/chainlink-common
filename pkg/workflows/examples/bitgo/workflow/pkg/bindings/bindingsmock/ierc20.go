package bindingsmock

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	evmmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/capabilitymock"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
)

type IERC20Mock struct {
	address    []byte
	original   *evmmock.ClientCapability
	underlying *evmmock.ClientCapability

	TotalSupply func() *big.Int
	// Other methods would be generated as well
}

func NewIERC20Mock(address common.Address, clientMock *evmmock.ClientCapability) *IERC20Mock {
	// copy the mock so that other contract interfaces can be implemented on the same contract
	original := *clientMock

	a := bindings.NewIERC20Abi()

	erc20Mock := &IERC20Mock{
		address:    address[:],
		original:   &original,
		underlying: clientMock,
	}

	clientMock.CallContract = func(ctx context.Context, input *evm.CallContractRequest) (*evm.CallContractReply, error) {
		if !bytes.Equal(erc20Mock.address, input.Call.To) {
			if original.CallContract == nil {
				return nil, fmt.Errorf("contract %s not found", common.BytesToAddress(erc20Mock.address).Hex())
			} else {
				return original.CallContract(ctx, input)
			}
		}

		data := input.Call.Data
		if len(data) < 4 {
			return nil, errors.New("data too short")
		}

		methodID := data[:4]
		// generate for each one
		totalSupplyMethod := a.Methods["totalSupply"]
		if bytes.Equal(methodID, totalSupplyMethod.ID) {
			supply := erc20Mock.TotalSupply()
			responseData, err := totalSupplyMethod.Outputs.Pack(supply)
			if err != nil {
				return nil, err
			}
			return &evm.CallContractReply{
				Data: responseData,
			}, nil

		} else if original.CallContract != nil {
			return original.CallContract(ctx, input)
		}

		return nil, fmt.Errorf("method with ID %x not implemented", methodID)
	}

	return erc20Mock
}
