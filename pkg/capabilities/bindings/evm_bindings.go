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
	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
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

var contractABI *abi.ABI

// Datastorage is an auto generated Go binding around an Ethereum contract.
type DataStorage struct {
	Address   []byte
	Options   *ContractInitOptions
	ABI       *abi.ABI
	evmClient evmcappb.Client
	codec     DataStorageCodec
}

type FilterLogTriggerDataStore struct {
	//TODO check if this should be specific values for DataStore indexed fields. If not this can go to common.
	Topic1     []byte
	Topic2     []byte
	Topic3     []byte
	Confidence evmcappb.ConfidenceLevel
	BlockDepth uint64
}

func (ds DataStorage) LogTrigger(runtime sdk.DonRuntime, LogTriggerInput *LogTriggerInput) sdk.DonTrigger[*evmpb.FilterLogsReply] {
}

func (ds *DataStorage) LogTriggerDataStoredLog(filterLogTrigger *FilterLogTriggerDataStore) sdk.DonTrigger[*ParsedLog[DataStored]] {
	//TODO properly use input for filters
	ds.evmClient.LogTrigger(&evmcappb.FilterLogTriggerRequest{
		Addresses:  [][]byte{ds.Address},
		Topics0:    [][]byte{ds.codec.DataStoreLogHash()},
		Confidence: evmcappb.ConfidenceLevel_FINALIZED,
	})

	panic("unimplemented")
}

func (ds *DataStorage) DataStoredLogHash() []byte {
	panic("unimplemented")
}

func (ds *DataStorage) DecodeDataStoredLog(abi []byte) (DataStored, any) {
	panic("unimplemented")
}

func (ds *DataStorage) EncodeUserDataStruct(userData UserData) ([]byte, error) {
	panic("unimplemented")
}

func (ds *DataStorage) FilterLogsDataStoredLog(runtime sdk.DonRuntime, options *FilterOptions) ([]ParsedLog[DataStored], any) {
	if options == nil {
		options = &FilterOptions{
			ToBlock: "finalized", //TODO we need a enum / constant
		}
	}
	filterLogsReplyPromise := ds.evmClient.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{ds.Address},
			Topics:    [][]byte{ds.codec.DataStoreLogHash()},
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
	return []ParsedLog[DataStored]{}, nil
}

// Default behaviour could be to get all known logs until the latest one and next called would retrieve from the last head read + 1 until the new head but this means storing internal state in the workflow
func (rda *DataStorage) QueryTrackedLogsDataStoredLog(runtime sdk.DonRuntime, options *QueryTrackedLogsOptions) ([]ParsedLog[DataStored], any) {
	rda.evmClient.QueryTrackedLogs(runtime, &evm.QueryTrackedLogsRequest{
		Expression: []*evm.Expression{
			//TODO add proper expression
			&evm.Expression{Evaluator: &evm.Expression_BooleanExpression{&evm.BooleanExpression{
				Expression: []*evm.Expression{},
			}}},
		},
	})
	//TODO transform []Log to DataStorageLog
	return []ParsedLog[DataStored]{}, nil
}

func (rda *DataStorage) RegisterLogTrackingDataStoredLog(runtime sdk.DonRuntime, options *LogTrackingOptions) {
	//TODO use log tracking options if set
	rda.evmClient.RegisterLogTracking(runtime, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Name:      "DataStored-" + common.Bytes2Hex(rda.Address),
			Addresses: [][]byte{rda.Address},
			EventSigs: [][]byte{rda.codec.DataStoreLogHash()},
		},
	})

}

func (rda *DataStorage) UnregisterLogTrackingDataStoredLog(runtime sdk.DonRuntime) {
	rda.evmClient.UnregisterLogTracking(runtime, &evm.UnregisterLogTrackingRequest{
		FilterName: "DataStored-" + common.Bytes2Hex(rda.Address),
	})
}

type Logs struct {
	DataStored
}

func NewDataStorageCodec() (DataStorageCodec, error) {
	abi, err := getAbi()
	if err != nil {
		return nil, err
	}
	return dataStorageCodec{
		abi: abi,
	}, nil
}

type DataStorageCodec interface {
	EncodeUserDataStruct(userData UserData) ([]byte, error)
	EncodeReadDataMethodCall(readData ReadDataInput) ([]byte, error)
	EncodeReadDataMethodInputs(readData ReadDataInput) ([]byte, error)
	DecodeReadDataMethodOutput(data []byte) (string, error)
	DataStoreLogHash() []byte
}

type dataStorageCodec struct {
	abi *abi.ABI
}

// DataStoreLogHash implements DataStorageCodec.
func (d dataStorageCodec) DataStoreLogHash() []byte {
	panic("unimplemented")
}

// decodeReadDataMethodOutput implements DataStorageCodec.
func (d dataStorageCodec) DecodeReadDataMethodOutput(data []byte) (string, error) {
	returnValue, err := d.abi.Methods["readData"].Outputs.Unpack(data)
	if err != nil {
		return "", err
	}
	castedReturnValue := returnValue[0].(string)
	return castedReturnValue, nil
}

// encodeReadDataMethodCall implements DataStorageCodec.
func (d dataStorageCodec) EncodeReadDataMethodCall(readDataInput ReadDataInput) ([]byte, error) {
	return d.abi.Pack("readData", readDataInput)
}

func (d dataStorageCodec) EncodeReadDataMethodInputs(readDataInput ReadDataInput) ([]byte, error) {
	return d.abi.Pack("readData", readDataInput.User, readDataInput.Key)
}

// encodeUserDataStruct implements DataStorageCodec.
func (d dataStorageCodec) EncodeUserDataStruct(UserData UserData) ([]byte, error) {
	panic("unimplemented")
}

func getAbi() (*abi.ABI, error) {
	if contractABI != nil {
		return contractABI, nil
	}
	tempABI, err := abi.JSON(strings.NewReader(DataStorageABI))
	if err != nil {
		return nil, err
	}
	contractABI = &tempABI
	return contractABI, nil
}

// NewDatastorage creates a new instance of Datastorage, bound to a specific deployed contract.
func NewDataStorage(contractInitOptions ContractInputs) (*DataStorage, error) {
	abi, err := getAbi()
	codec, err := NewDataStorageCodec()
	if err != nil {
		return nil, err
	}
	dataStorage := &DataStorage{
		Address: contractInitOptions.Address,
		Options: contractInitOptions.Options,
		ABI:     abi,
		codec:   codec,
	}
	return dataStorage, nil
}

type ReadDataInput struct {
	User []byte
	Key  string
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
func (rda DataStorage) ReadData(runtime sdk.DonRuntime, args ReadDataInput, options *ReadOptions) (string, error) {
	calldata, err := rda.codec.EncodeReadDataMethodCall(args)
	if err != nil {
		return "", err
	}
	if options == nil {
		options = &ReadOptions{
			BlockNumber: nil,
		}
	}

	callReplyPromise := rda.evmClient.CallContract(runtime, &evm.CallContractRequest{
		Call: &evm.CallMsg{
			To:   rda.Address,
			Data: calldata,
		},
		BlockNumber: toPbBigInt(options.BlockNumber),
	})
	reply, err := callReplyPromise.Await()
	if err != nil {
		return "", err
	}
	return rda.codec.DecodeReadDataMethodOutput(reply.Data)
}

func (ds *DataStorage) WriteReportUserData(runtime sdk.DonRuntime, userDataStruct UserData, options *WriteOptions) (*evmpb.Hash, error) {
	reportEncoded, _ := ds.codec.EncodeUserDataStruct(userDataStruct)
	commonReport := GenerateReport(getChainID(ds.evmClient), reportEncoded)
	var writeReportReplyPromise = ds.evmClient.WriteReport(runtime, &evm.WriteReportRequest{
		Receiver: ds.Address,
		Report: &evm.SignedReport{
			RawReport:     commonReport.ReportContext,
			ReportContext: commonReport.RawReport,
			Signatures:    commonReport.Signatures,
			Id:            commonReport.ID,
		},
		GasConfig: ds.Options.GasConfig,
	})

	writeReportReply, err := writeReportReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	if writeReportReply.TxStatus == evm.TransactionStatus_TX_FAILURE {
		return nil, &TxFatalError{
			//TODO add error message for when Transaction is FATAL
			// Message: writeReportReply.Message,
			Message: "Fatal tx execution",
		}
	}
	for {
		txByHashPromise := ds.evmClient.GetTransactionByHash(runtime, &evm.GetTransactionByHashRequest{
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

func (userData UserData) Encode(userDataStruct UserData) ([]byte, error) {
	//TODO add code to generate struct
	encoded := []byte{}
	return encoded, nil
}

type DataStored struct {
	Sender []byte
	Key    string
	Value  string
}

type UserData struct {
	Key   string
	Value string
}

type LogTriggerInput struct {
	Topic1     []byte
	Topic2     []byte
	Topic3     []byte
	Confidence evmcappb.ConfidenceLevel
	BlockDepth uint64
}

func toPbBigInt(int *big.Int) *pb.BigInt {
	panic("unimplemented")
}

func getChainID(evm evm.EVMClient) uint32 {
	//Return the chain selector
	panic("unimplemented")
}
