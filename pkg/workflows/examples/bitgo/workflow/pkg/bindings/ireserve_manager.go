package bindings

import (
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

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
	TotalMinted  big.Int
	TotalReserve big.Int
}

func (ur UpdateReserves) WriteReport(runtime sdk.Runtime, UpdateReservesStruct UpdateReservesStruct, options *WriteOptions) sdk.Promise[*evm.WriteReportReply] {
	// if ur.reserveManager.gasConfig == nil && options.GasConfig == nil {
	// 	sdk.Primitive
	// }
	commonReport := GenerateReport(10, []byte{})
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
