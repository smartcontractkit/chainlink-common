package bindings

import (
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

// TODO figure out how we know the type...?

//go:embed solc/bin/IReserveManager.abi
var IReserveManagerAbi string

var iReserveManagerApi, _ = abi.JSON(strings.NewReader(IReserveManagerAbi))

type IReserverManager struct {
	Structs        Structs
	ContractInputs ContractInputs
}

type Structs struct {
	UpdateReserves UpdateReserves
}

func NewIReserveManager(contracInputs ContractInputs) IReserverManager {
	reserveManager := IReserverManager{ContractInputs: contracInputs}
	reserveManager.Structs = Structs{
		UpdateReserves: UpdateReserves{
			reserveManager: &reserveManager,
		},
	}
	return reserveManager
}

type UpdateReserves struct {
	reserveManager *IReserverManager
}

type UpdateReservesStruct struct {
	TotalMinted  *big.Int
	TotalReserve *big.Int
}

func (ur UpdateReserves) WriteReport(runtime sdk.Runtime, updateReserves UpdateReservesStruct, options *WriteOptions) sdk.Promise[*evm.WriteReportReply] {
	// if ur.reserveManager.gasConfig == nil && options.GasConfig == nil {
	// 	sdk.Primitive
	// }

	body, err := iReserveManagerApi.Methods["updateReserves"].Inputs.Pack(updateReserves.TotalMinted, updateReserves.TotalReserve)
	if err != nil {
		return sdk.PromiseFromResult[*evm.WriteReportReply](nil, err)
	}

	commonReport := GenerateReport(ur.reserveManager.ContractInputs.EVM.ChainSelector, body)
	writeReportReplyPromise := ur.reserveManager.ContractInputs.EVM.WriteReport(runtime, &evm.WriteReportRequest{
		Receiver: ur.reserveManager.ContractInputs.Address,
		Report: &evm.SignedReport{
			RawReport:     commonReport.RawReport,
			ReportContext: commonReport.ReportContext,
			Signatures:    commonReport.Signatures,
			Id:            commonReport.ID,
		},
	})

	return writeReportReplyPromise
}
