package bindings

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/capability"
	evm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// DatastorageMetaData contains all meta data concerning the Datastorage contract.
var DataStorageMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"DataNotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"AccessLogged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"logAccess\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"readData\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"storeData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DatastorageABI is the input ABI used to generate the binding from.
// Deprecated: Use DatastorageMetaData.ABI instead.
var DataStorageABI = DatastorageMetaData.ABI

// Datastorage is an auto generated Go binding around an Ethereum contract.
type DataStorage struct {
	Address   evm.Address
	Options   *ContractInitOptions
	ABI       abi.ABI
	EVM       evmcappb.EVM
	StoreData StoreData
	ReadData  ReadData
	Methods   Methods
	Logs      Logs
	Structs   Structs
}

type Logs struct {
	DataStored
}

type Methods struct {
	ReadData
}

type Errors struct {
	dataStorage *DataStorage
	ReadData    ReadData
	StoreData   StoreData
}

type Structs struct {
	UserData UserData
}

// NewDatastorage creates a new instance of Datastorage, bound to a specific deployed contract.
func NewDataStorage(contractInitOptions ContractInputs) (*DataStorage, error) {
	abi, err := abi.JSON(strings.NewReader(DatastorageABI))
	if err != nil {
		return nil, err
	}
	dataStorage := &DataStorage{
		Address: contractInitOptions.Address,
		Options: contractInitOptions.Options,
		ABI:     abi,
	}
	return dataStorage, nil
}

type ReadDataArgs struct {
	User evm.Address
	Key  string
}

type ReadData struct {
	_Datastorage *DataStorage
}

func (rda ReadData) EncodeCalldata(readDataArgs ReadDataArgs) ([]byte, error) {
	return rda._Datastorage.ABI.Pack("readData", readDataArgs)
}

func (rda ReadData) EncodeInput(args ReadDataArgs) ([]byte, error) {
	return rda._Datastorage.ABI.Pack("readData", args.User, args.Key)

}

func (rda ReadData) DecodeOutput(data []byte) (string, error) {
	returnValue, err := rda._Datastorage.ABI.Methods["readData"].Outputs.Unpack(data)
	if err != nil {
		return "", err
	}
	castedReturnValue := returnValue[0].(string)
	return castedReturnValue, nil

}

type DataNotFoundError struct {
	Requested evmpb.Address
	Key       string
	Reason    string
}

// Implement the error interface
func (e *DataNotFoundError) Error() string {
	return fmt.Sprintf("Error requested: %s, key: %s, reason :%", e.Requested, e.Key, e.Reason)
}

func (e Errors) DecodeDataNotFoundError(data []byte) (DataNotFoundError, error) {
	//TODO missing logic to unpack error
	// e.dataStorage.ABI.Errors["DataNotFound"].Inputs.Unpack(data[4:])
	return DataNotFoundError{}, nil
}

// ReadData is a free data retrieval call binding the contract method 0xf5bfa815.
//
// Solidity: function readData(address user, string key) view returns(string)
func (rda ReadData) Call(runtime sdk.DonRuntime, args ReadDataArgs, options *ReadOptions) (string, error) {
	calldata, err := rda.EncodeInput(args)
	if err != nil {
		return "", err
	}
	if options == nil {
		options = &ReadOptions{
			BlockNumber: nil,
		}
	}

	callReplyPromise := rda._Datastorage.EVM.CallContract(runtime, &evm.CallContractRequest{
		Call: &evm.CallMsg{
			To: &rda._Datastorage.Address,
			Data: &evm.ABIPayload{
				Abi: calldata,
			},
		},
		BlockNumber: toPbBigInt(options.BlockNumber),
	})
	reply, err := callReplyPromise.Await()
	if err != nil {
		return "", err
	}
	return rda.DecodeOutput(reply.Data.Abi)
}

type StoreData struct {
	_Datastorage *DataStorage
}

type UserData struct {
	dataStorage *DataStorage
}

func (userData UserData) WriteReport(runtime sdk.DonRuntime, userDataStruct UserDataStruct, options *WriteOptions) (*evmpb.Hash, error) {
	reportEncoded, _ := userData.Encode(userDataStruct)
	commonReport := GenerateReport(getChainID(userData.dataStorage.EVM), reportEncoded)
	var writeReportReplyPromise = userData.dataStorage.EVM.WriteReport(runtime, &evm.WriteReportRequest{
		Receiver: &userData.dataStorage.Address,
		Report: &evm.SignedReport{
			RawReport:     commonReport.ReportContext,
			ReportContext: commonReport.RawReport,
			Signatures:    commonReport.Signatures,
			Id:            commonReport.ID,
		},
		GasConfig: userData.dataStorage.Options.GasConfig,
	})
	
	writerReportReply, err := writeReportReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	if writerReportReply.TxStatus == evm.TransactionStatus_TX_FAILURE {
		return nil, &TxFatalError{
			//TODO add error message for when Transaction is FATAL
			// Message: writeReportReply.Message,
			Message: "Fatal tx execution",
		}
	}
	for {
		txByHashPromise := userData.dataStorage.EVM.GetTransactionByHash(runtime, &evm.GetTransactionByHashRequest{
			Hash: writeReportReply.TxHash,
		})
		getTxResult, err := txByHashPromise.Await()
		if err != nil {
			return nil, err
		}
		//TODO we need more logic to call eth_call for simulate TX with same data to get the error and return an error from the smart contract
		//TODO we need to add info about finalization for easy of use
		if getTxResult.Transaction.IsFinalized {
			if writeReportReply.ReceiverContractExecutionStatus == evm.ReceiverContractExecutionStatus_FAILURE {
				return nil, &ReceiverContractError{
					Message: "Transaction finalized but receiver smart contract failed to execute",
					TxHash:  writeReportReply.TxHash,
				}
			}
			return writeReportReply.TxHash, nil
		}
	}
}

func (userData UserData) Encode(userDataStruct UserDataStruct) ([]byte, error) {
	//TODO add code to generate struct
	encoded := []byte{}
	return encoded, nil
}

type DataStored struct {
	dataStorage *DataStorage
	Hash        *evmpb.Hash
}

type DataStoredValue struct {
	Sender evm.Address
	Key    string
	Value  string
}

func (DataStored) Decode(encodedLog []byte) (DataStoredValue, error) {
	panic("implement")
}

func (dsl DataStored) FilterLogs(runtime sdk.DonRuntime, options *FilterOptions) ([]ParsedLog[DataStoredValue], error) {
	if options == nil {
		options = &FilterOptions{
			ToBlock: "finalized", //TODO we need a enum / constant
		}
	}
	filterLogsReplyPromise := dsl.dataStorage.EVM.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: []*evm.Address{
				&evm.Address{
					Address: dsl.dataStorage.Address.Address,
				},
			},
			Topics: []*evm.Topics{
				&evm.Topics{
					Topic: []*evm.Hash{
						&evm.Hash{
							Hash: []byte(dsl.dataStorage.ABI.Events["DataStored"].Sig),
						},
					},
				},
			},
			BlockHash: options.BlockHash,
			FromBlock: toPbBigInt(options.FromBlock),
			ToBlock:   toPbBigInt(options.ToBlock),
		},
	})
	_, err := filterLogsReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	//TODO convert []Log into []EnhancedLog
	return []ParsedLog[DataStoredValue]{}, nil
}

type UserDataStruct struct {
	Key   string
	Value string
}

func (dsl DataStored) RegisterLogTracking(runtime sdk.DonRuntime, options *LogTrackingOptions) {
	//TODO use log tracking options if set
	dsl.dataStorage.EVM.RegisterLogTracking(runtime, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Name:      dsl.getLogTrackingName(),
			Addresses: []*evm.Address{&dsl.dataStorage.Address},
			EventSigs: []*evm.Hash{
				&evm.Hash{
					Hash: []byte(dsl.dataStorage.ABI.Events["DataStored"].Sig),
				},
			},
		},
	})
}

func (dsl DataStored) UnregisterLogTracking(runtime sdk.DonRuntime) {
	dsl.dataStorage.EVM.UnregisterLogTracking(runtime, &evm.UnregisterLogTrackingRequest{
		FilterName: dsl.getLogTrackingName(),
	})
}

func (dsl DataStored) getLogTrackingName() string {
	return "DataStorageLog" + string(dsl.dataStorage.Address.Address)
}

// Default behaviour could be to get all known logs until the latest one and next called would retrieve from the last head read + 1 until the new head but this means storing internal state in the workflow
func (dsl DataStored) QueryTrackedLogs(runtime sdk.DonRuntime, options *QueryTrackedLogsOptions) ([]ParsedLog[DataStoredValue], error) {
	dsl.dataStorage.EVM.QueryTrackedLogs(runtime, &evm.QueryTrackedLogsRequest{
		Expression: []*evm.Expression{
			//TODO add proper expression
			&evm.Expression{Evaluator: &evm.Expression_BooleanExpression{&evm.BooleanExpression{
				Expression: []*evm.Expression{},
			}}},
		},
	})
	//TODO transform []Log to DataStorageLog
	return []ParsedLog[DataStoredValue]{}, nil
}

type LogTriggerInput struct {
	Topic1     []byte
	Topic2     []byte
	Topic3     []byte
	Confidence evmcappb.ConfidenceLevel
	BlockDepth uint64
}

func (dsl DataStored) LogTrigger(runtime sdk.DonRuntime, LogTriggerInput *LogTriggerInput) sdk.DonTrigger[*evmpb.FilterLogsReply] {
	//TODO properly use input for filters
	return dsl.dataStorage.EVM.LogTrigger(&evmcappb.FilterLogTriggerRequest{
		Addresses:  [][]byte{dsl.dataStorage.Address.Address},
		Topics0:    [][]byte{dsl.Hash.Hash},
		Confidence: evmcappb.ConfidenceLevel_FINALIZED,
	})
}

func toPbBigInt(int *big.Int) *pb.BigInt {
	panic("unimplemented")
}

func getChainID(evm evmcappb.EVM) uint32 {
	//Return the chain selector
	panic("unimplemented")
}
