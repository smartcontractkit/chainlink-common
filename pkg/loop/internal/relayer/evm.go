package relayer

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"

	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type EVMClient struct {
	grpcClient evmpb.EVMClient
}

// CalculateTransactionFee implements types.EVMService.
func (e *EVMClient) CalculateTransactionFee(ctx context.Context, receiptGasInfo evmtypes.ReceiptGasInfo) (*evmtypes.TransactionFee, error) {
	reply, err := e.grpcClient.CalculateTransactionFee(ctx, &evmpb.CalculateTransactionFeeRequest{GasInfo: &evmpb.ReceiptGasInfo{
		GasUsed:           receiptGasInfo.GasUsed,
		EffectiveGasPrice: valuespb.NewBigIntFromInt(receiptGasInfo.EffectiveGasPrice),
	}})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.TransactionFee{TransactionFee: valuespb.NewIntFromBigInt(reply.GetTransactionFee())}, nil
}

// SubmitTransaction implements types.EVMService.
func (e *EVMClient) SubmitTransaction(ctx context.Context, txRequest evmtypes.SubmitTransactionRequest) (*evmtypes.TransactionResult, error) {
	reply, err := e.grpcClient.SubmitTransaction(ctx, &evmpb.SubmitTransactionRequest{
		To:        txRequest.To[:],
		Data:      txRequest.Data,
		GasConfig: evmpb.ConvertGasConfigToProto(txRequest.GasConfig),
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.TransactionResult{
		TxStatus: evmpb.ConvertTxStatusFromProto(reply.TxStatus),
		TxHash:   evmtypes.Hash(reply.TxHash),
	}, nil
}

func NewEVMCClient(grpcClient evmpb.EVMClient) *EVMClient {
	return &EVMClient{
		grpcClient: grpcClient,
	}
}

var _ types.EVMService = (*EVMClient)(nil)

func (e *EVMClient) GetTransactionFee(ctx context.Context, transactionID string) (*evmtypes.TransactionFee, error) {
	reply, err := e.grpcClient.GetTransactionFee(ctx, &evmpb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.TransactionFee{TransactionFee: valuespb.NewIntFromBigInt(reply.GetTransactionFee())}, nil
}

func (e *EVMClient) CallContract(ctx context.Context, request evmtypes.CallContractRequest) (*evmtypes.CallContractReply, error) {
	protoCallMsg, err := evmpb.ConvertCallMsgToProto(request.Msg)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoConfidenceLevel, err := chaincommonpb.ConvertConfidenceToProto(request.ConfidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.CallContract(ctx, &evmpb.CallContractRequest{
		Call:            protoCallMsg,
		BlockNumber:     valuespb.NewBigIntFromInt(request.BlockNumber),
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.CallContractReply{Data: reply.GetData()}, nil
}

func (e *EVMClient) FilterLogs(ctx context.Context, request evmtypes.FilterLogsRequest) (*evmtypes.FilterLogsReply, error) {
	protoConfidenceLevel, err := chaincommonpb.ConvertConfidenceToProto(request.ConfidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.FilterLogs(ctx, &evmpb.FilterLogsRequest{
		FilterQuery:     evmpb.ConvertFilterToProto(request.FilterQuery),
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.FilterLogsReply{Logs: evmpb.ConvertLogsFromProto(reply.GetLogs())}, nil
}

func (e *EVMClient) BalanceAt(ctx context.Context, request evmtypes.BalanceAtRequest) (*evmtypes.BalanceAtReply, error) {
	protoConfidenceLevel, err := chaincommonpb.ConvertConfidenceToProto(request.ConfidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.BalanceAt(ctx, &evmpb.BalanceAtRequest{
		Account:         request.Address[:],
		BlockNumber:     valuespb.NewBigIntFromInt(request.BlockNumber),
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.BalanceAtReply{Balance: valuespb.NewIntFromBigInt(reply.GetBalance())}, nil
}

func (e *EVMClient) EstimateGas(ctx context.Context, msg *evmtypes.CallMsg) (uint64, error) {
	protoCallMsg, err := evmpb.ConvertCallMsgToProto(msg)
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.EstimateGas(ctx, &evmpb.EstimateGasRequest{Msg: protoCallMsg})
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	return reply.GetGas(), nil
}

func (e *EVMClient) GetTransactionByHash(ctx context.Context, hash evmtypes.Hash) (*evmtypes.Transaction, error) {
	reply, err := e.grpcClient.GetTransactionByHash(ctx, &evmpb.GetTransactionByHashRequest{Hash: hash[:]})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return evmpb.ConvertTransactionFromProto(reply.GetTransaction())
}

func (e *EVMClient) GetTransactionReceipt(ctx context.Context, txHash evmtypes.Hash) (*evmtypes.Receipt, error) {
	reply, err := e.grpcClient.GetTransactionReceipt(ctx, &evmpb.GetTransactionReceiptRequest{Hash: txHash[:]})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return evmpb.ConvertReceiptFromProto(reply.GetReceipt())
}

func (e *EVMClient) HeaderByNumber(ctx context.Context, request evmtypes.HeaderByNumberRequest) (*evmtypes.HeaderByNumberReply, error) {
	protoConfidenceLevel, err := chaincommonpb.ConvertConfidenceToProto(request.ConfidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.HeaderByNumber(ctx, &evmpb.HeaderByNumberRequest{
		BlockNumber:     valuespb.NewBigIntFromInt(request.Number),
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	header, err := evmpb.ConvertHeadFromProto(reply.GetHeader())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &evmtypes.HeaderByNumberReply{Header: header}, nil

}
func (e *EVMClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evmtypes.Log, error) {
	protoExpressions, err := evmpb.ConvertExpressionsToProto(filterQuery)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoLimitAndSort, err := chaincommonpb.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoConfidenceLevel, err := chaincommonpb.ConvertConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.QueryTrackedLogs(ctx, &evmpb.QueryTrackedLogsRequest{
		Expression:      protoExpressions,
		LimitAndSort:    protoLimitAndSort,
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return evmpb.ConvertLogsFromProto(reply.GetLogs()), nil
}

func (e *EVMClient) RegisterLogTracking(ctx context.Context, filter evmtypes.LPFilterQuery) error {
	_, err := e.grpcClient.RegisterLogTracking(ctx, &evmpb.RegisterLogTrackingRequest{Filter: evmpb.ConvertLPFilterToProto(filter)})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.grpcClient.UnregisterLogTracking(ctx, &evmpb.UnregisterLogTrackingRequest{FilterName: filterName})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.grpcClient.GetTransactionStatus(ctx, &evmpb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, net.WrapRPCErr(err)
	}

	return types.TransactionStatus(reply.GetTransactionStatus()), nil
}

func (e *EVMClient) GetForwarderForEOA(ctx context.Context, eoa, ocr2AggregatorID evmtypes.Address, pluginType string) (forwarder evmtypes.Address, err error) {
	reply, err := e.grpcClient.GetForwarderForEOA(ctx, &evmpb.GetForwarderForEOARequest{Addr: eoa[:], Aggr: ocr2AggregatorID[:], PluginType: pluginType})
	if err != nil {
		return evmtypes.Address{}, net.WrapRPCErr(err)
	}
	return evmtypes.Address(reply.GetAddr()), nil
}

type evmServer struct {
	evmpb.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

var _ evmpb.EVMServer = (*evmServer)(nil)

func newEVMServer(impl types.EVMService, b *net.BrokerExt) *evmServer {
	return &evmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}

func (e *evmServer) GetTransactionFee(ctx context.Context, request *evmpb.GetTransactionFeeRequest) (*evmpb.GetTransactionFeeReply, error) {
	txFee, err := e.impl.GetTransactionFee(ctx, request.GetTransactionId())
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionFeeReply{TransactionFee: valuespb.NewBigIntFromInt(txFee.TransactionFee)}, nil
}

func (e *evmServer) CallContract(ctx context.Context, request *evmpb.CallContractRequest) (*evmpb.CallContractReply, error) {
	callMsg, err := evmpb.ConvertCallMsgFromProto(request.GetCall())
	if err != nil {
		return nil, err
	}

	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	reply, err := e.impl.CallContract(ctx, evmtypes.CallContractRequest{
		Msg:             callMsg,
		BlockNumber:     valuespb.NewIntFromBigInt(request.GetBlockNumber()),
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	if reply == nil {
		return nil, errors.New("reply is nil")
	}

	return &evmpb.CallContractReply{Data: reply.Data}, nil
}
func (e *evmServer) FilterLogs(ctx context.Context, request *evmpb.FilterLogsRequest) (*evmpb.FilterLogsReply, error) {
	filter, err := evmpb.ConvertFilterFromProto(request.GetFilterQuery())
	if err != nil {
		return nil, err
	}

	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	reply, err := e.impl.FilterLogs(ctx, evmtypes.FilterLogsRequest{
		FilterQuery:     filter,
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	if reply == nil {
		return nil, errors.New("reply is nil")
	}

	return &evmpb.FilterLogsReply{Logs: evmpb.ConvertLogsToProto(reply.Logs)}, nil
}
func (e *evmServer) BalanceAt(ctx context.Context, request *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	addr, err := evmpb.AddressFromBytes(request.Account)
	if err != nil {
		return nil, err
	}
	balance, err := e.impl.BalanceAt(ctx, evmtypes.BalanceAtRequest{
		Address:         addr,
		BlockNumber:     valuespb.NewIntFromBigInt(request.GetBlockNumber()),
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, errors.New("balance is nil")
	}

	return &evmpb.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance.Balance)}, nil
}

func (e *evmServer) EstimateGas(ctx context.Context, request *evmpb.EstimateGasRequest) (*evmpb.EstimateGasReply, error) {
	callMsg, err := evmpb.ConvertCallMsgFromProto(request.GetMsg())
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmpb.EstimateGasReply{Gas: gas}, nil
}

func (e *evmServer) GetTransactionByHash(ctx context.Context, request *evmpb.GetTransactionByHashRequest) (*evmpb.GetTransactionByHashReply, error) {
	tx, err := e.impl.GetTransactionByHash(ctx, evmtypes.Hash(request.GetHash()))
	if err != nil {
		return nil, err
	}

	protoTx, err := evmpb.ConvertTransactionToProto(tx)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionByHashReply{Transaction: protoTx}, nil
}

func (e *evmServer) GetTransactionReceipt(ctx context.Context, request *evmpb.GetTransactionReceiptRequest) (*evmpb.GetTransactionReceiptReply, error) {
	receipt, err := e.impl.GetTransactionReceipt(ctx, evmtypes.Hash(request.GetHash()))
	if err != nil {
		return nil, err
	}

	protoReceipt, err := evmpb.ConvertReceiptToProto(receipt)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionReceiptReply{Receipt: protoReceipt}, nil
}

func (e *evmServer) HeaderByNumber(ctx context.Context, request *evmpb.HeaderByNumberRequest) (*evmpb.HeaderByNumberReply, error) {
	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}
	reply, err := e.impl.HeaderByNumber(ctx, evmtypes.HeaderByNumberRequest{
		Number:          valuespb.NewIntFromBigInt(request.BlockNumber),
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	if reply == nil {
		return nil, errors.New("reply is nil")
	}

	return &evmpb.HeaderByNumberReply{
		Header: evmpb.ConvertHeadToProto(reply.Header),
	}, nil
}

func (e *evmServer) QueryTrackedLogs(ctx context.Context, request *evmpb.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error) {
	expressions, err := evmpb.ConvertExpressionsFromProto(request.GetExpression())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := chaincommonpb.ConvertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	logs, err := e.impl.QueryTrackedLogs(ctx, expressions, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{Logs: evmpb.ConvertLogsToProto(logs)}, nil
}

func (e *evmServer) RegisterLogTracking(ctx context.Context, request *evmpb.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	lpFilter, err := evmpb.ConvertLPFilterFromProto(request.GetFilter())
	if err != nil {
		return nil, err
	}
	return nil, e.impl.RegisterLogTracking(ctx, lpFilter)
}

func (e *evmServer) UnregisterLogTracking(ctx context.Context, request *evmpb.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, e.impl.UnregisterLogTracking(ctx, request.GetFilterName())
}

func (e *evmServer) GetTransactionStatus(ctx context.Context, request *evmpb.GetTransactionStatusRequest) (*evmpb.GetTransactionStatusReply, error) {
	txStatus, err := e.impl.GetTransactionStatus(ctx, request.GetTransactionId())
	if err != nil {
		return nil, err
	}

	//nolint: gosec // G115
	return &evmpb.GetTransactionStatusReply{TransactionStatus: evmpb.TransactionStatus(txStatus)}, nil
}

func (e *evmServer) SubmitTransaction(ctx context.Context, request *evmpb.SubmitTransactionRequest) (*evmpb.SubmitTransactionReply, error) {
	txResult, err := e.impl.SubmitTransaction(ctx, evmpb.ConvertSubmitTransactionRequestFromProto(request))
	if err != nil {
		return nil, err
	}
	return &evmpb.SubmitTransactionReply{
		TxHash:   txResult.TxHash[:],
		TxStatus: evmpb.ConvertTxStatusToProto(txResult.TxStatus),
	}, nil
}

func (e *evmServer) CalculateTransactionFee(ctx context.Context, request *evmpb.CalculateTransactionFeeRequest) (*evmpb.CalculateTransactionFeeReply, error) {
	txFee, err := e.impl.CalculateTransactionFee(ctx, evmtypes.ReceiptGasInfo{
		GasUsed:           request.GasInfo.GasUsed,
		EffectiveGasPrice: valuespb.NewIntFromBigInt(request.GasInfo.EffectiveGasPrice),
	})
	if err != nil {
		return nil, err
	}
	return &evmpb.CalculateTransactionFeeReply{
		TransactionFee: valuespb.NewBigIntFromInt(txFee.TransactionFee),
	}, nil
}

func (e *evmServer) GetForwarderForEOA(ctx context.Context, request *evmpb.GetForwarderForEOARequest) (*evmpb.GetForwarderForEOAReply, error) {
	forwarder, err := e.impl.GetForwarderForEOA(ctx, evmtypes.Address(request.GetAddr()), evmtypes.Address(request.GetAggr()), request.GetPluginType())
	if err != nil {
		return nil, err
	}
	return &evmpb.GetForwarderForEOAReply{Addr: forwarder[:]}, nil
}
