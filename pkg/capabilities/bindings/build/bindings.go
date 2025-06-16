// Code generated — DO NOT EDIT.

package bindings

import (
	"bytes"
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
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/bindings"
	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

var (
	_ = bytes.Equal
	_ = errors.New
	_ = fmt.Sprintf
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

var DataStorageMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"DataNotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"AccessLogged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"logAccess\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"}],\"name\":\"onReport\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"readData\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"storeData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"internalType\":\"structDataStorage.UserData\",\"name\":\"userData\",\"type\":\"tuple\"}],\"name\":\"storeUserData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b506111338061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c80634ece5b4c14610059578063805f21321461007557806398458c5d14610091578063ccf15827146100ad578063f5bfa815146100c9575b5f5ffd5b610073600480360381019061006e9190610591565b6100f9565b005b61008f600480360381019061008a9190610664565b6101bd565b005b6100ab60048036038101906100a69190610704565b61029a565b005b6100c760048036038101906100c2919061074b565b610391565b005b6100e360048036038101906100de91906107f0565b6103e5565b6040516100f091906108bd565b60405180910390f35b81815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f208686604051610147929190610919565b90815260200160405180910390209182610162929190610b6e565b503373ffffffffffffffffffffffffffffffffffffffff167fc95c7d5d3ac582f659cd004afbea77723e1315567b6557f3c059e8eb9586518f858585856040516101af9493929190610c67565b60405180910390a250505050565b5f82828101906101cd9190610e1a565b905080602001515f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20825f01516040516102219190610e91565b9081526020016040518091039020908161023b9190610ea7565b503373ffffffffffffffffffffffffffffffffffffffff167fc95c7d5d3ac582f659cd004afbea77723e1315567b6557f3c059e8eb9586518f825f0151836020015160405161028b929190610f76565b60405180910390a25050505050565b8080602001906102aa9190610fb7565b5f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2083805f01906102f59190610fb7565b604051610303929190610919565b9081526020016040518091039020918261031e929190610b6e565b503373ffffffffffffffffffffffffffffffffffffffff167fc95c7d5d3ac582f659cd004afbea77723e1315567b6557f3c059e8eb9586518f82805f01906103669190610fb7565b8480602001906103769190610fb7565b6040516103869493929190610c67565b60405180910390a250565b3373ffffffffffffffffffffffffffffffffffffffff167fe2ab1536af9681ad9e5927bca61830526c4cd932e970162eef77328af1fdcfb583836040516103d9929190611019565b60405180910390a25050565b60605f5f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f208484604051610434929190610919565b9081526020016040518091039020805461044d90610995565b80601f016020809104026020016040519081016040528092919081815260200182805461047990610995565b80156104c45780601f1061049b576101008083540402835291602001916104c4565b820191905f5260205f20905b8154815290600101906020018083116104a757829003601f168201915b505050505090505f815103610514578484846040517ff1e5020900000000000000000000000000000000000000000000000000000000815260040161050b939291906110ba565b60405180910390fd5b809150509392505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f84011261055157610550610530565b5b8235905067ffffffffffffffff81111561056e5761056d610534565b5b60208301915083600182028301111561058a57610589610538565b5b9250929050565b5f5f5f5f604085870312156105a9576105a8610528565b5b5f85013567ffffffffffffffff8111156105c6576105c561052c565b5b6105d28782880161053c565b9450945050602085013567ffffffffffffffff8111156105f5576105f461052c565b5b6106018782880161053c565b925092505092959194509250565b5f5f83601f84011261062457610623610530565b5b8235905067ffffffffffffffff81111561064157610640610534565b5b60208301915083600182028301111561065d5761065c610538565b5b9250929050565b5f5f5f5f6040858703121561067c5761067b610528565b5b5f85013567ffffffffffffffff8111156106995761069861052c565b5b6106a58782880161060f565b9450945050602085013567ffffffffffffffff8111156106c8576106c761052c565b5b6106d48782880161060f565b925092505092959194509250565b5f5ffd5b5f604082840312156106fb576106fa6106e2565b5b81905092915050565b5f6020828403121561071957610718610528565b5b5f82013567ffffffffffffffff8111156107365761073561052c565b5b610742848285016106e6565b91505092915050565b5f5f6020838503121561076157610760610528565b5b5f83013567ffffffffffffffff81111561077e5761077d61052c565b5b61078a8582860161053c565b92509250509250929050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107bf82610796565b9050919050565b6107cf816107b5565b81146107d9575f5ffd5b50565b5f813590506107ea816107c6565b92915050565b5f5f5f6040848603121561080757610806610528565b5b5f610814868287016107dc565b935050602084013567ffffffffffffffff8111156108355761083461052c565b5b6108418682870161053c565b92509250509250925092565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61088f8261084d565b6108998185610857565b93506108a9818560208601610867565b6108b281610875565b840191505092915050565b5f6020820190508181035f8301526108d58184610885565b905092915050565b5f81905092915050565b828183375f83830152505050565b5f61090083856108dd565b935061090d8385846108e7565b82840190509392505050565b5f6109258284866108f5565b91508190509392505050565b5f82905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806109ac57607f821691505b6020821081036109bf576109be610968565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302610a217fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826109e6565b610a2b86836109e6565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f610a6f610a6a610a6584610a43565b610a4c565b610a43565b9050919050565b5f819050919050565b610a8883610a55565b610a9c610a9482610a76565b8484546109f2565b825550505050565b5f5f905090565b610ab3610aa4565b610abe818484610a7f565b505050565b5b81811015610ae157610ad65f82610aab565b600181019050610ac4565b5050565b601f821115610b2657610af7816109c5565b610b00846109d7565b81016020851015610b0f578190505b610b23610b1b856109d7565b830182610ac3565b50505b505050565b5f82821c905092915050565b5f610b465f1984600802610b2b565b1980831691505092915050565b5f610b5e8383610b37565b9150826002028217905092915050565b610b788383610931565b67ffffffffffffffff811115610b9157610b9061093b565b5b610b9b8254610995565b610ba6828285610ae5565b5f601f831160018114610bd3575f8415610bc1578287013590505b610bcb8582610b53565b865550610c32565b601f198416610be1866109c5565b5f5b82811015610c0857848901358255600182019150602085019450602081019050610be3565b86831015610c255784890135610c21601f891682610b37565b8355505b6001600288020188555050505b50505050505050565b5f610c468385610857565b9350610c538385846108e7565b610c5c83610875565b840190509392505050565b5f6040820190508181035f830152610c80818688610c3b565b90508181036020830152610c95818486610c3b565b905095945050505050565b5f5ffd5b610cad82610875565b810181811067ffffffffffffffff82111715610ccc57610ccb61093b565b5b80604052505050565b5f610cde61051f565b9050610cea8282610ca4565b919050565b5f5ffd5b5f5ffd5b5f67ffffffffffffffff821115610d1157610d1061093b565b5b610d1a82610875565b9050602081019050919050565b5f610d39610d3484610cf7565b610cd5565b905082815260208101848484011115610d5557610d54610cf3565b5b610d608482856108e7565b509392505050565b5f82601f830112610d7c57610d7b610530565b5b8135610d8c848260208601610d27565b91505092915050565b5f60408284031215610daa57610da9610ca0565b5b610db46040610cd5565b90505f82013567ffffffffffffffff811115610dd357610dd2610cef565b5b610ddf84828501610d68565b5f83015250602082013567ffffffffffffffff811115610e0257610e01610cef565b5b610e0e84828501610d68565b60208301525092915050565b5f60208284031215610e2f57610e2e610528565b5b5f82013567ffffffffffffffff811115610e4c57610e4b61052c565b5b610e5884828501610d95565b91505092915050565b5f610e6b8261084d565b610e7581856108dd565b9350610e85818560208601610867565b80840191505092915050565b5f610e9c8284610e61565b915081905092915050565b610eb08261084d565b67ffffffffffffffff811115610ec957610ec861093b565b5b610ed38254610995565b610ede828285610ae5565b5f60209050601f831160018114610f0f575f8415610efd578287015190505b610f078582610b53565b865550610f6e565b601f198416610f1d866109c5565b5f5b82811015610f4457848901518255600182019150602085019450602081019050610f1f565b86831015610f615784890151610f5d601f891682610b37565b8355505b6001600288020188555050505b505050505050565b5f6040820190508181035f830152610f8e8185610885565b90508181036020830152610fa28184610885565b90509392505050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83356001602003843603038112610fd357610fd2610fab565b5b80840192508235915067ffffffffffffffff821115610ff557610ff4610faf565b5b60208301925060018202360383131561101157611010610fb3565b5b509250929050565b5f6020820190508181035f830152611032818486610c3b565b90509392505050565b611044816107b5565b82525050565b7f4e6f2064617461206173736f63696174656420776974682074686973206b65795f8201527f2e00000000000000000000000000000000000000000000000000000000000000602082015250565b5f6110a4602183610857565b91506110af8261104a565b604082019050919050565b5f6060820190506110cd5f83018661103b565b81810360208301526110e0818486610c3b565b905081810360408301526110f381611098565b905094935050505056fea26469706673582212207936b380efb76d4d5d1017df160ba6bbd5a1b4390c6989dab812208b5ab2fc3464736f6c634300081e0033",
}

type DataStorageCodec interface {
	EncodeReadDataMethodCall(in ReadDataInput) ([]byte, error)
	DecodeReadDataMethodOutput(data []byte) (string, error)
	EncodeDataStorageUserDataStruct(in DataStorageUserDataInput) ([]byte, error)
	AccessLoggedLogHash() []byte
	DecodeAccessLogged(log *evm.Log) (*AccessLogged, error)
	DataStoredLogHash() []byte
	DecodeDataStored(log *evm.Log) (*DataStored, error)
}

type dataStorageCodecImpl struct {
	abi *abi.ABI
}

func NewDataStorageCodec() (DataStorageCodec, error) {
	parsed, err := abi.JSON(strings.NewReader(DataStorageMetaData.ABI))
	if err != nil {
		return nil, err
	}
	return &dataStorageCodecImpl{abi: &parsed}, nil
}

func (c *dataStorageCodecImpl) EncodeReadDataMethodCall(in ReadDataInput) ([]byte, error) {
	return c.abi.Pack("readData", in.User, in.Key)
}
func (c *dataStorageCodecImpl) DecodeReadDataMethodOutput(data []byte) (string, error) {
	vals, err := c.abi.Methods["readData"].Outputs.Unpack(data)
	if err != nil {
		return *new(string), err
	}
	return vals[0].(string), nil
}

func (c *dataStorageCodecImpl) EncodeDataStorageUserDataStruct(in DataStorageUserDataInput) ([]byte, error) {
	return c.abi.Pack("dataStorageUserData", in)
}

func (c *dataStorageCodecImpl) AccessLoggedLogHash() []byte {
	return c.abi.Events["AccessLogged"].ID.Bytes()
}

// DecodeAccessLogged decodes a log into a AccessLogged struct.
func (c *dataStorageCodecImpl) DecodeAccessLogged(log *evm.Log) (*AccessLogged, error) {
	event := new(AccessLogged)
	if err := c.abi.UnpackIntoInterface(event, "AccessLogged", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["AccessLogged"].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *dataStorageCodecImpl) DataStoredLogHash() []byte {
	return c.abi.Events["DataStored"].ID.Bytes()
}

// DecodeDataStored decodes a log into a DataStored struct.
func (c *dataStorageCodecImpl) DecodeDataStored(log *evm.Log) (*DataStored, error) {
	event := new(DataStored)
	if err := c.abi.UnpackIntoInterface(event, "DataStored", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["DataStored"].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

type DataStorage struct {
	Address   []byte
	Options   *bindings.ContractInitOptions
	ABI       *abi.ABI
	evmClient evmcappb.Client
	codec     DataStorageCodec
}

func NewDataStorage(
	client evmcappb.Client,
	address []byte,
	options *bindings.ContractInitOptions,
) (*DataStorage, error) {
	parsed, err := abi.JSON(strings.NewReader(DataStorageMetaData.ABI))
	if err != nil {
		return nil, err
	}
	codec, err := NewDataStorageCodec()
	if err != nil {
		return nil, err
	}
	return &DataStorage{
		Address:   address,
		Options:   options,
		ABI:       &parsed,
		evmClient: client,
		codec:     codec,
	}, nil
}

type ReadDataInput struct {
	User common.Address
	Key  string
}

func (c DataStorage) ReadData(
	runtime sdk.DonRuntime,
	args ReadDataInput,
	options *bindings.ReadOptions,
) (string, error) {
	calldata, err := c.codec.EncodeReadDataMethodCall(args)
	if err != nil {
		return *new(string), err
	}
	if options == nil {
		options = &bindings.ReadOptions{BlockNumber: nil}
	}
	promise := c.evmClient.CallContract(runtime, &evm.CallContractRequest{
		Call:        &evm.CallMsg{To: c.Address, Data: calldata},
		BlockNumber: toPbBigInt(options.BlockNumber),
	})
	reply, err := promise.Await()
	if err != nil {
		return *new(string), err
	}
	return c.codec.DecodeReadDataMethodOutput(reply.Data)
}

type DataStorageUserDataInput struct {
	Key   string
	Value string
}

func (c DataStorage) WriteReportDataStorageUserData(
	runtime sdk.DonRuntime,
	input DataStorageUserDataInput,
	options *bindings.WriteOptions,
) (*big.Int, error) {
	encoded, err := c.codec.EncodeDataStorageUserDataStruct(input)
	if err != nil {
		return nil, err
	}
	report := bindings.GenerateReport(getChainID(c.evmClient), encoded)
	writeReportReplyPromise := c.evmClient.WriteReport(runtime, &evm.WriteReportRequest{
		Receiver: c.Address,
		Report: &evm.SignedReport{
			RawReport:     report.RawReport,
			ReportContext: report.ReportContext,
			Signatures:    report.Signatures,
			Id:            report.ID,
		},
		GasConfig: options.GasConfig,
	})
	reply, err := writeReportReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	if reply.TxStatus == evm.TransactionStatus_TX_FAILURE {
		return nil, &bindings.TxFatalError{
			Message: "Fatal tx execution",
		}
	}
	for {
		txByHashPromise := c.evmClient.GetTransactionByHash(runtime, &evm.GetTransactionByHashRequest{
			Hash: reply.TxHash,
		})
		getTxResult, err := txByHashPromise.Await()
		if err != nil {
			return nil, err
		}
		if getTxResult.Transaction.IsFinalized {
			if reply.ReceiverContractExecutionStatus == evm.ReceiverContractExecutionStatus_FAILURE {
				return nil, &bindings.ReceiverContractError{
					Message: "Transaction finalized but receiver contract failed to execute",
					TxHash:  reply.TxHash,
				}
			}
			return reply.TxHash, nil
		}
	}
}

// DataNotFound represents the DataNotFound error raised by the DataStorage contract.
type DataNotFound struct {
	Requester common.Address
	Key       string
	Reason    string
}

// TODO: possibly clean this up
// DecodeDataNotFoundError decodes a DataNotFound error from revert data.
func (c *DataStorage) DecodeDataNotFoundError(data []byte) (*DataNotFound, error) {
	args := c.ABI.Errors["DataNotFound"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 3 {
		return nil, fmt.Errorf("expected 3 values, got %%d", len(values))
	}

	requester, ok0 := values[0].(common.Address)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for requester in DataNotFound error")
	}

	key, ok1 := values[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("unexpected type for key in DataNotFound error")
	}

	reason, ok2 := values[2].(string)
	if !ok2 {
		return nil, fmt.Errorf("unexpected type for reason in DataNotFound error")
	}

	return &DataNotFound{
		Requester: requester,
		Key:       key,
		Reason:    reason,
	}, nil
}

func (c *DataStorage) UnpackError(data []byte) (any, error) {
	switch common.Bytes2Hex(data[:4]) {

	case common.Bytes2Hex(c.ABI.Errors["DataNotFound"].ID.Bytes()):
		return c.DecodeDataNotFoundError(data)

	default:
		return nil, errors.New("unknown error selector")
	}
}

// Error implements the error interface for DataNotFound.
func (e *DataNotFound) Error() string {
	return fmt.Sprintf("DataNotFound error: requester=%%v; key=%%v; reason=%%v;", e.Requester, e.Key, e.Reason)
}

// AccessLogged represents a AccessLogged event raised by the DataStorage contract.
type AccessLogged struct {
	Caller  common.Address
	Message string
}

func (c *DataStorage) RegisterLogTrackingAccessLogged(runtime sdk.DonRuntime, options *bindings.LogTrackingOptions) {
	//TODO use log tracking options if set
	c.evmClient.RegisterLogTracking(runtime, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Name:      "AccessLogged-" + common.Bytes2Hex(c.Address),
			Addresses: [][]byte{c.Address},
			EventSigs: [][]byte{c.codec.AccessLoggedLogHash()},
		},
	})
}

func (c *DataStorage) UnregisterLogTrackingAccessLogged(runtime sdk.DonRuntime) {
	c.evmClient.UnregisterLogTracking(runtime, &evm.UnregisterLogTrackingRequest{
		FilterName: "AccessLogged-" + common.Bytes2Hex(c.Address),
	})
}

func (c *DataStorage) QueryTrackedLogsAccessLogged(runtime sdk.DonRuntime, options *bindings.QueryTrackedLogsOptions) ([]bindings.ParsedLog[AccessLogged], any) {
	promise := c.evmClient.QueryTrackedLogs(runtime, &evm.QueryTrackedLogsRequest{
		Expression: []*evm.Expression{
			//TODO add proper expression
			&evm.Expression{Evaluator: &evm.Expression_BooleanExpression{&evm.BooleanExpression{
				Expression: []*evm.Expression{},
			}}},
		},
	})
	reply, err := promise.Await()
	if err != nil {
		return nil, fmt.Errorf("failed to query tracked logs: %w", err)
	}
	logs := reply.Logs
	parsedLogs := make([]bindings.ParsedLog[AccessLogged], len(logs))
	for i, log := range logs {
		decodedLog, err := c.codec.DecodeAccessLogged(log)
		if err != nil {
			return nil, fmt.Errorf("failed to decode AccessLogged log: %w", err)
		}
		parsedLogs[i] = bindings.ParsedLog[AccessLogged]{
			LogData: *decodedLog,
			RawLog:  log,
		}
	}

	return parsedLogs, nil
}

func (c *DataStorage) FilterLogsAccessLogged(runtime sdk.DonRuntime, options *bindings.FilterOptions) ([]bindings.ParsedLog[AccessLogged], error) {
	if options == nil {
		options = &bindings.FilterOptions{
			ToBlock: "finalized", //TODO we need a enum / constant
		}
	}
	filterLogsReplyPromise := c.evmClient.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.codec.AccessLoggedLogHash()}},
			}, BlockHash: options.BlockHash,
			FromBlock: toPbBigInt(options.FromBlock),
			ToBlock:   toPbBigInt(options.ToBlock),
		},
	})
	reply, err := filterLogsReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	logs := reply.Logs
	parsedLogs := make([]bindings.ParsedLog[AccessLogged], len(logs))
	for i, log := range logs {
		decodedLog, err := c.codec.DecodeAccessLogged(log)
		if err != nil {
			return nil, fmt.Errorf("failed to decode AccessLogged log: %w", err)
		}
		parsedLogs[i] = bindings.ParsedLog[AccessLogged]{
			LogData: *decodedLog,
			RawLog:  log,
		}
	}

	return parsedLogs, nil
}

// DataStored represents a DataStored event raised by the DataStorage contract.
type DataStored struct {
	Sender common.Address
	Key    string
	Value  string
}

func (c *DataStorage) RegisterLogTrackingDataStored(runtime sdk.DonRuntime, options *bindings.LogTrackingOptions) {
	//TODO use log tracking options if set
	c.evmClient.RegisterLogTracking(runtime, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Name:      "DataStored-" + common.Bytes2Hex(c.Address),
			Addresses: [][]byte{c.Address},
			EventSigs: [][]byte{c.codec.DataStoredLogHash()},
		},
	})
}

func (c *DataStorage) UnregisterLogTrackingDataStored(runtime sdk.DonRuntime) {
	c.evmClient.UnregisterLogTracking(runtime, &evm.UnregisterLogTrackingRequest{
		FilterName: "DataStored-" + common.Bytes2Hex(c.Address),
	})
}

func (c *DataStorage) QueryTrackedLogsDataStored(runtime sdk.DonRuntime, options *bindings.QueryTrackedLogsOptions) ([]bindings.ParsedLog[DataStored], any) {
	promise := c.evmClient.QueryTrackedLogs(runtime, &evm.QueryTrackedLogsRequest{
		Expression: []*evm.Expression{
			//TODO add proper expression
			&evm.Expression{Evaluator: &evm.Expression_BooleanExpression{&evm.BooleanExpression{
				Expression: []*evm.Expression{},
			}}},
		},
	})
	reply, err := promise.Await()
	if err != nil {
		return nil, fmt.Errorf("failed to query tracked logs: %w", err)
	}
	logs := reply.Logs
	parsedLogs := make([]bindings.ParsedLog[DataStored], len(logs))
	for i, log := range logs {
		decodedLog, err := c.codec.DecodeDataStored(log)
		if err != nil {
			return nil, fmt.Errorf("failed to decode DataStored log: %w", err)
		}
		parsedLogs[i] = bindings.ParsedLog[DataStored]{
			LogData: *decodedLog,
			RawLog:  log,
		}
	}

	return parsedLogs, nil
}

func (c *DataStorage) FilterLogsDataStored(runtime sdk.DonRuntime, options *bindings.FilterOptions) ([]bindings.ParsedLog[DataStored], error) {
	if options == nil {
		options = &bindings.FilterOptions{
			ToBlock: "finalized", //TODO we need a enum / constant
		}
	}
	filterLogsReplyPromise := c.evmClient.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.codec.DataStoredLogHash()}},
			}, BlockHash: options.BlockHash,
			FromBlock: toPbBigInt(options.FromBlock),
			ToBlock:   toPbBigInt(options.ToBlock),
		},
	})
	reply, err := filterLogsReplyPromise.Await()
	if err != nil {
		return nil, err
	}
	logs := reply.Logs
	parsedLogs := make([]bindings.ParsedLog[DataStored], len(logs))
	for i, log := range logs {
		decodedLog, err := c.codec.DecodeDataStored(log)
		if err != nil {
			return nil, fmt.Errorf("failed to decode DataStored log: %w", err)
		}
		parsedLogs[i] = bindings.ParsedLog[DataStored]{
			LogData: *decodedLog,
			RawLog:  log,
		}
	}

	return parsedLogs, nil
}

func toPbBigInt(i *big.Int) *pb.BigInt    { panic("unimplemented") }
func getChainID(e evmcappb.Client) uint32 { panic("unimplemented") }
