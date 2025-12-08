package relayer

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

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
	pbTxRequest := &evmpb.SubmitTransactionRequest{
		To:   txRequest.To[:],
		Data: txRequest.Data,
	}

	if txRequest.GasConfig != nil {
		gasCfg, err := evmpb.ConvertGasConfigToProto(*txRequest.GasConfig)
		if err != nil {
			return nil, err
		}
		pbTxRequest.GasConfig = gasCfg
	}

	reply, err := e.grpcClient.SubmitTransaction(ctx, pbTxRequest)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	h, err := evmpb.ConvertHashFromProto(reply.TxHash)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.TransactionResult{
		TxStatus:         evmpb.ConvertTxStatusFromProto(reply.TxStatus),
		TxHash:           h,
		TxIdempotencyKey: reply.TxIdempotencyKey,
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
		IsExternal:      request.IsExternal,
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

	filter, err := evmpb.ConvertFilterToProto(request.FilterQuery)
	if err != nil {
		return nil, net.WrapRPCErr(fmt.Errorf("failed to convert filter request err: %w", err))
	}

	reply, err := e.grpcClient.FilterLogs(ctx, &evmpb.FilterLogsRequest{
		FilterQuery:     filter,
		ConfidenceLevel: protoConfidenceLevel,
		IsExternal:      request.IsExternal,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	logs, err := evmpb.ConvertLogsFromProto(reply.GetLogs())
	if err != nil {
		return nil, err
	}

	return &evmtypes.FilterLogsReply{Logs: logs}, nil
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

func (e *EVMClient) GetTransactionByHash(ctx context.Context, request evmtypes.GetTransactionByHashRequest) (*evmtypes.Transaction, error) {
	reply, err := e.grpcClient.GetTransactionByHash(ctx, &evmpb.GetTransactionByHashRequest{
		Hash:       request.Hash[:],
		IsExternal: request.IsExternal,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return evmpb.ConvertTransactionFromProto(reply.GetTransaction())
}

func (e *EVMClient) GetTransactionReceipt(ctx context.Context, request evmtypes.GeTransactionReceiptRequest) (*evmtypes.Receipt, error) {
	reply, err := e.grpcClient.GetTransactionReceipt(ctx, &evmpb.GetTransactionReceiptRequest{
		Hash:       request.Hash[:],
		IsExternal: request.IsExternal,
	})
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

	header, err := evmpb.ConvertHeaderFromProto(reply.GetHeader())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.HeaderByNumberReply{Header: &header}, nil

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

	logs, err := evmpb.ConvertLogsFromProto(reply.GetLogs())
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (e *EVMClient) GetFiltersNames(ctx context.Context) ([]string, error) {
	// TODO PLEX-1465: once code is moved away, remove this GetFiltersNames method
	names, err := e.grpcClient.GetFiltersNames(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return names.GetItems(), nil
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

func (e *EVMClient) GetLatestLPBlock(ctx context.Context) (*evmtypes.LPBlock, error) {
	reply, err := e.grpcClient.GetLatestLPBlock(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	h, err := evmpb.ConvertHashFromProto(reply.GetLpBlock().GetHash())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.LPBlock{
		BlockTimestamp:       reply.GetLpBlock().GetBlockTimestamp(),
		LatestBlockNumber:    reply.GetLpBlock().GetLatestBlockNumber(),
		FinalizedBlockNumber: reply.GetLpBlock().GetFinalizedBlockNumber(),
		SafeBlockNumber:      reply.GetLpBlock().GetSafeBlockNumber(),
		BlockHash:            h,
	}, nil
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
		IsExternal:      request.IsExternal,
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
		IsExternal:      request.IsExternal,
	})
	if err != nil {
		return nil, err
	}

	logs, err := evmpb.ConvertLogsToProto(reply.Logs)
	if err != nil {
		return nil, err
	}

	return &evmpb.FilterLogsReply{Logs: logs}, nil
}
func (e *evmServer) BalanceAt(ctx context.Context, request *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	conf, err := chaincommonpb.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	addr, err := evmpb.ConvertAddressFromProto(request.Account)
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
	h, err := evmpb.ConvertHashFromProto(request.GetHash())
	if err != nil {
		return nil, err
	}
	tx, err := e.impl.GetTransactionByHash(ctx, evmtypes.GetTransactionByHashRequest{
		Hash:       h,
		IsExternal: request.IsExternal,
	})
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
	h, err := evmpb.ConvertHashFromProto(request.GetHash())
	if err != nil {
		return nil, err
	}
	receipt, err := e.impl.GetTransactionReceipt(ctx, evmtypes.GeTransactionReceiptRequest{
		Hash:       h,
		IsExternal: request.IsExternal,
	})
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

	header, err := evmpb.ConvertHeaderToProto(reply.Header)
	if err != nil {
		return nil, err
	}

	return &evmpb.HeaderByNumberReply{
		Header: header,
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

	logsReply, err := evmpb.ConvertLogsToProto(logs)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{Logs: logsReply}, nil
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
	req, err := evmpb.ConvertSubmitTransactionRequestFromProto(request)
	if err != nil {
		return nil, err
	}

	txResult, err := e.impl.SubmitTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	return &evmpb.SubmitTransactionReply{
		TxHash:           txResult.TxHash[:],
		TxStatus:         evmpb.ConvertTxStatusToProto(txResult.TxStatus),
		TxIdempotencyKey: txResult.TxIdempotencyKey,
	}, nil
}

func (e *evmServer) GetLatestLPBlock(ctx context.Context, _ *emptypb.Empty) (*evmpb.GetLatestLPBlockReply, error) {
	b, err := e.impl.GetLatestLPBlock(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetLatestLPBlockReply{
		LpBlock: &evmpb.LPBlock{
			Hash:                 b.BlockHash[:],
			LatestBlockNumber:    b.LatestBlockNumber,
			FinalizedBlockNumber: b.FinalizedBlockNumber,
			SafeBlockNumber:      b.SafeBlockNumber,
			BlockTimestamp:       b.BlockTimestamp,
		},
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

func (e *evmServer) GetFiltersNames(ctx context.Context, _ *emptypb.Empty) (*evmpb.GetFiltersNamesReply, error) {
	// TODO PLEX-1465: once code is moved away, remove this GetFiltersNames method
	names, err := e.impl.GetFiltersNames(ctx)
	if err != nil {
		return nil, err
	}
	return &evmpb.GetFiltersNamesReply{Items: names}, nil
}
