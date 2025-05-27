package bindings

import (
	"context"
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
	evm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
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
	ChainID   uint32
	Options   *ContractInitOptions
	ABI       abi.ABI
	EVMClient evm.EVMClient
}

// NewDatastorage creates a new instance of Datastorage, bound to a specific deployed contract.
func NewDataStorage(contractInitOptions ContractInitInputs) (*DataStorage, error) {
	abi, err := abi.JSON(strings.NewReader(DatastorageABI))
	if err != nil {
		return nil, err
	}
	dataStorage := &DataStorage{
		Address: contractInitOptions.Address,
		ChainID: contractInitOptions.ChainID,
		Options: contractInitOptions.Options,
		ABI:     abi,
	}
	return dataStorage, nil
}

type ReadDataArgs struct {
	User evm.Address
	Key  string
}

type ReadDataAccessor struct {
	_Datastorage DataStorage
}

func (rda ReadDataAccessor) EncodeCalldata(readDataArgs ReadDataArgs) ([]byte, error) {
	return rda._Datastorage.ABI.Pack("readData", readDataArgs)
}

func (ds DataStorage) ReadDataAccessor() ReadDataAccessor {
	return ReadDataAccessor{
		_Datastorage: ds,
	}
}

func (rda ReadDataAccessor) EncodeInput(args ReadDataArgs) ([]byte, error) {
	return rda._Datastorage.ABI.Pack("readData", args.User, args.Key)

}

func (rda ReadDataAccessor) DecodeOutput(data []byte) (string, error) {
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

type Errors struct {
	dataStorage *DataStorage
}

func (e Errors) DecodeDataNotFoundError(data []byte) (DataNotFoundError, error) {
	//TODO missing logic to unpack error
	// e.dataStorage.ABI.Errors["DataNotFound"].Inputs.Unpack(data[4:])
	return DataNotFoundError{}, nil
}

// ReadData is a free data retrieval call binding the contract method 0xf5bfa815.
//
// Solidity: function readData(address user, string key) view returns(string)
func (rda ReadDataAccessor) ReadData(ctx context.Context, args ReadDataArgs, options *ReadOptions) (string, error) {
	calldata, err := rda.EncodeInput(args)
	if err != nil {
		return "", err
	}
	if options == nil {
		options = &ReadOptions{
			BlockNumber: nil,
		}
	}
	reply, err := rda._Datastorage.EVMClient.CallContract(ctx, &evm.CallContractRequest{
		Call: &evm.CallMsg{
			To: &rda._Datastorage.Address,
			Data: &evm.ABIPayload{
				Abi: calldata,
			},
		},
		BlockNumber: options.BlockNumber,
	})
	if err != nil {
		return "", err
	}
	return rda.DecodeOutput(reply.Data.Abi)
}

type StoreDataAccessor struct {
	_Datastorage DataStorage
}

func (dsl DataStorage) UserDataStructAccessor() UserDataStructAccessor {
	return UserDataStructAccessor{
		dataStorage: &dsl,
	}
}

type UserDataStructAccessor struct {
	dataStorage *DataStorage
}

func (accessor UserDataStructAccessor) Encode(userData UserDataStruct) ([]byte, any) {
	panic("unimplemented")
}

func (accessor UserDataStructAccessor) WriteReport(ctx context.Context, userDataStruct UserDataStruct, options *WriteOptions) (*evmpb.Hash, error) {
	reportEncoded, _ := accessor.EncodeUserDataStruct(userDataStruct)
	commonReport := GenerateReport(accessor.dataStorage.ChainID, reportEncoded)
	var writeReportReply, err = accessor.dataStorage.EVMClient.WriteReport(ctx, &evm.WriteReportRequest{
		Receiver: &accessor.dataStorage.Address,
		Report: &evm.SignedReport{
			RawReport:     commonReport.ReportContext,
			ReportContext: commonReport.RawReport,
			Signatures:    commonReport.Signatures,
			Id:            commonReport.ID,
		},
		GasConfig: accessor.dataStorage.Options.GasConfig,
	})
	if err != nil {
		return nil, err
	}
	if writeReportReply.TxStatus == evm.TransactionStatus_TRANSACTION_STATUS_FATAL {
		return nil, &TxFatalError{
			//TODO add error message for when Transaction is FATAL
			// Message: writeReportReply.Message,
			Message: "Fatal tx execution",
		}
	}
	for {
		getTxResult, err := accessor.dataStorage.EVMClient.GetTransactionByHash(ctx, &evm.GetTransactionByHashRequest{
			Hash: writeReportReply.TxHash,
		})
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

func (accessor UserDataStructAccessor) EncodeUserDataStruct(userDataStruct UserDataStruct) ([]byte, error) {
	//TODO add code to generate struct
	encoded := []byte{}
	return encoded, nil
}

func (dsl DataStorage) DataStoredLogAccessor() DataStoredLogAccessor {
	return DataStoredLogAccessor{
		dataStorage: &dsl,
		Hash: dsl.ABI.Events["DataStored"].ID,
	}
}

type DataStoredLogAccessor struct {
	dataStorage *DataStorage
	Hash *evmpb.Hash
}

type DataStoredLog struct {
	Sender evm.Address
	Key    string
	Value  string
}

func (DataStoredLogAccessor) Decode(encodedLog []byte) (DataStoredLog, error) {
	panic("implement")
}

func (dsl DataStoredLogAccessor) FilterLogs(ctx context.Context, options *FilterOptions) ([]ParsedLog[DataStoredLog], error) {
	if options == nil {
		options = &FilterOptions{
			ToBlock: "finalized", //TODO we need a enum / constant
		}
	}
	filterLogsReply, err := dsl.dataStorage.EVMClient.FilterLogs(ctx, &evm.FilterLogsRequest{
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
			FromBlock: options.FromBlock,
			ToBlock:   options.ToBlock,
		},
	})
	if err != nil {
		return nil, err
	}
	//TODO convert []Log into []EnhancedLog
	return []ParsedLog[DataStoredLog]{}, nil
}

type UserDataStruct struct {
	Key   string
	Value string
}

func (dsl DataStoredLogAccessor) RegisterLogTracking(ctx context.Context, options *LogTrackingOptions) {
	//TODO use log tracking options if set
	dsl.dataStorage.EVMClient.RegisterLogTracking(ctx, &evm.RegisterLogTrackingRequest{
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

func (dsl DataStoredLogAccessor) UnregisterLogTracking(ctx context.Context) {
	dsl.dataStorage.EVMClient.UnregisterLogTracking(ctx, &evm.UnregisterLogTrackingRequest{
		FilterName: dsl.getLogTrackingName(),
	})
}

func (dsl DataStoredLogAccessor) getLogTrackingName() string {
	return "DataStorageLog" + string(dsl.dataStorage.Address.Address)
}

// Default behaviour could be to get all known logs until the latest one and next called would retrieve from the last head read + 1 until the new head but this means storing internal state in the workflow
func (dsl DataStoredLogAccessor) QueryTrackedLogs(ctx context.Context, options *QueryTrackedLogsOptions) ([]ParsedLog[DataStoredLog], error) {
	dsl.dataStorage.EVMClient.QueryTrackedLogs(ctx, &evm.QueryTrackedLogsRequest{
		Expression: []*evm.Expression{
			//TODO add proper expression
			&evm.Expression{Evaluator: &evm.Expression_BooleanExpression{&evm.BooleanExpression{
				Expression: []*evm.Expression{},
			}}},
		},
	})
	//TODO transform []Log to DataStorageLog
	return []ParsedLog[DataStoredLog]{}, nil
}

func (dsl DataStoredLogAccessor) DataStorageLogTrigger(ctx context.Context) sdk.DonTrigger[ParsedLog[DataStoredLog]] {
	//TODO implement once LogTrigger is ready
	panic("to be implemented")
}

// contract Caller {
//     function safeCall(address otherAddr, uint x) public returns (uint) {
//         Other other = Other(otherAddr);
//         try other.doSomething(x) returns (uint result) {
//             return result;
//         } catch Error(string memory reason) {
//             // Revert with original error message
//             revert(string(abi.encodePacked("External call failed: ", reason)));
//         } catch (bytes memory lowLevelData) {
//             // Fallback for non-standard errors - We would have to capture this data into an event, trigger then event, then fetch the event from the off-chain code and get the encoded error
//             revert(string(abi.encodePacked("Unknown error: ", string(lowLevelData))));
//         }
//     }
// }
