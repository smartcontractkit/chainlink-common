package bindingsmock

import (
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	evmmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/capabilitymock"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
)

type IReserverManagerMock struct {
	UpdateReserves func(reserves *bindings.UpdateReservesStruct) error

	// Other methods would be generated as well
}

func NewIReserverManagerMock(address common.Address, clientMock *evmmock.ClientCapability) *IReserverManagerMock {
	reserveManagerMock := &IReserverManagerMock{}
	a := bindings.NewIReserveManagerAbi()
	updateReserves := a.Methods["updateReserves"]
	funcMap := map[string]func([]byte) ([]byte, error){}
	writeReport := func(payload []byte, config *evm.GasConfig) (*evm.WriteReportReply, error) {
		if reserveManagerMock.UpdateReserves == nil {
			return nil, errors.New("method update reserves not found on the contract")
		}

		tmp, err := updateReserves.Inputs.Unpack(payload)
		if err != nil {
			return nil, err
		}

		rTmp := reflect.ValueOf(tmp[0])

		parsedInput := &bindings.UpdateReservesStruct{
			TotalMinted:  rTmp.FieldByIndex([]int{0}).Interface().(*big.Int),
			TotalReserve: rTmp.FieldByIndex([]int{1}).Interface().(*big.Int),
		}

		if err := reserveManagerMock.UpdateReserves(parsedInput); err != nil {
			return nil, err
		}

		// TODO
		return &evm.WriteReportReply{}, nil
	}
	bindings.AddInterfaceMock(address, clientMock, funcMap, writeReport)
	return reserveManagerMock
}
