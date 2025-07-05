package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

// evmClient wraps the EVMRelayerSetClient by attaching a RelayerID to EVMClient requests.
// The attached RelayerID is stored in the context metadata.
type evmClient struct {
	relayID types.RelayID
	client  evmpb.EVMClient
}

var _ evmpb.EVMClient = (*evmClient)(nil)

func (e *evmClient) CalculateTransactionFee(ctx context.Context, in *evmpb.CalculateTransactionFeeRequest, opts ...grpc.CallOption) (*evmpb.CalculateTransactionFeeReply, error) {
	return e.client.CalculateTransactionFee(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e *evmClient) SubmitTransaction(ctx context.Context, in *evmpb.SubmitTransactionRequest, opts ...grpc.CallOption) (*evmpb.SubmitTransactionReply, error) {
	return e.client.SubmitTransaction(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) GetTransactionFee(ctx context.Context, in *evmpb.GetTransactionFeeRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionFeeReply, error) {
	return e.client.GetTransactionFee(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) CallContract(ctx context.Context, in *evmpb.CallContractRequest, opts ...grpc.CallOption) (*evmpb.CallContractReply, error) {
	return e.client.CallContract(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) FilterLogs(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error) {
	return e.client.FilterLogs(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) BalanceAt(ctx context.Context, in *evmpb.BalanceAtRequest, opts ...grpc.CallOption) (*evmpb.BalanceAtReply, error) {
	return e.client.BalanceAt(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) EstimateGas(ctx context.Context, in *evmpb.EstimateGasRequest, opts ...grpc.CallOption) (*evmpb.EstimateGasReply, error) {
	return e.client.EstimateGas(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) GetTransactionByHash(ctx context.Context, in *evmpb.GetTransactionByHashRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionByHashReply, error) {
	return e.client.GetTransactionByHash(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) GetTransactionReceipt(ctx context.Context, in *evmpb.GetTransactionReceiptRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionReceiptReply, error) {
	return e.client.GetTransactionReceipt(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) LatestAndFinalizedHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*evmpb.LatestAndFinalizedHeadReply, error) {
	return e.client.LatestAndFinalizedHead(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) QueryTrackedLogs(ctx context.Context, in *evmpb.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evmpb.QueryTrackedLogsReply, error) {
	return e.client.QueryTrackedLogs(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) RegisterLogTracking(ctx context.Context, in *evmpb.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.RegisterLogTracking(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) UnregisterLogTracking(ctx context.Context, in *evmpb.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.UnregisterLogTracking(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) GetTransactionStatus(ctx context.Context, in *evmpb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionStatusReply, error) {
	return e.client.GetTransactionStatus(appendRelayID(ctx, e.relayID), in, opts...)
}

func (e evmClient) GetForwarderForEOA(ctx context.Context, in *evmpb.GetForwarderForEOARequest, opts ...grpc.CallOption) (*evmpb.GetForwarderForEOAReply, error) {
	return e.client.GetForwarderForEOA(appendRelayID(ctx, e.relayID), in, opts...)
}

func (s *Server) GetTransactionFee(ctx context.Context, request *evmpb.GetTransactionFeeRequest) (*evmpb.GetTransactionFeeReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionFeeReply{TransactionFee: valuespb.NewBigIntFromInt(reply.TransactionFee)}, nil
}

func (s *Server) CallContract(ctx context.Context, request *evmpb.CallContractRequest) (*evmpb.CallContractReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	callMsg, err := evmpb.ConvertCallMsgFromProto(request.Call)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.CallContract(ctx, callMsg, valuespb.NewIntFromBigInt(request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmpb.CallContractReply{
		Data: reply,
	}, nil
}

func (s *Server) FilterLogs(ctx context.Context, request *evmpb.FilterLogsRequest) (*evmpb.FilterLogsReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	expression, err := evmpb.ConvertFilterFromProto(request.FilterQuery)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.FilterLogs(ctx, expression)
	if err != nil {
		return nil, err
	}

	return &evmpb.FilterLogsReply{Logs: evmpb.ConvertLogsToProto(reply)}, nil
}

func (s *Server) BalanceAt(ctx context.Context, request *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	balance, err := evmService.BalanceAt(ctx, evm.Address(request.GetAccount()), valuespb.NewIntFromBigInt(request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmpb.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance)}, nil
}

func (s *Server) EstimateGas(ctx context.Context, request *evmpb.EstimateGasRequest) (*evmpb.EstimateGasReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	callMsg, err := evmpb.ConvertCallMsgFromProto(request.GetMsg())
	if err != nil {
		return nil, err
	}

	gasLimit, err := evmService.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmpb.EstimateGasReply{Gas: gasLimit}, nil
}

func (s *Server) GetTransactionByHash(ctx context.Context, request *evmpb.GetTransactionByHashRequest) (*evmpb.GetTransactionByHashReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionByHash(ctx, evm.Hash(request.GetHash()))
	if err != nil {
		return nil, err
	}

	tx, err := evmpb.ConvertTransactionToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionByHashReply{
		Transaction: tx,
	}, nil
}

func (s *Server) GetTransactionReceipt(ctx context.Context, request *evmpb.GetTransactionReceiptRequest) (*evmpb.GetTransactionReceiptReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionReceipt(ctx, evm.Hash(request.GetHash()))
	if err != nil {
		return nil, err
	}

	receipt, err := evmpb.ConvertReceiptToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionReceiptReply{
		Receipt: receipt,
	}, nil
}

func (s *Server) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmpb.LatestAndFinalizedHeadReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	latest, finalized, err := evmService.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb.LatestAndFinalizedHeadReply{
		Latest:    evmpb.ConvertHeadToProto(latest),
		Finalized: evmpb.ConvertHeadToProto(finalized),
	}, nil
}

func (s *Server) QueryTrackedLogs(ctx context.Context, request *evmpb.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	expression, err := evmpb.ConvertExpressionsFromProto(request.GetExpression())
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

	logs, err := evmService.QueryTrackedLogs(ctx, expression, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{Logs: evmpb.ConvertLogsToProto(logs)}, nil
}

func (s *Server) RegisterLogTracking(ctx context.Context, request *evmpb.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	filter, err := evmpb.ConvertLPFilterFromProto(request.GetFilter())
	if err != nil {
		return nil, err
	}

	if err = evmService.RegisterLogTracking(ctx, filter); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UnregisterLogTracking(ctx context.Context, request *evmpb.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	if err = evmService.UnregisterLogTracking(ctx, request.GetFilterName()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetTransactionStatus(ctx context.Context, request *evmpb.GetTransactionStatusRequest) (*evmpb.GetTransactionStatusReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	txStatus, err := evmService.GetTransactionStatus(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	//nolint: gosec // G115
	return &evmpb.GetTransactionStatusReply{TransactionStatus: evmpb.TransactionStatus(txStatus)}, nil
}

func (s *Server) SubmitTransaction(ctx context.Context, request *evmpb.SubmitTransactionRequest) (*evmpb.SubmitTransactionReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	txResult, err := evmService.SubmitTransaction(ctx, evm.SubmitTransactionRequest{
		To:        evm.Address(request.To),
		Data:      evm.ABIPayload(request.Data),
		GasConfig: evmpb.ConvertGasConfigFromProto(request.GetGasConfig()),
	})
	if err != nil {
		return nil, err
	}

	return &evmpb.SubmitTransactionReply{
		TxHash:   txResult.TxHash[:],
		TxStatus: evmpb.ConvertTxStatusToProto(txResult.TxStatus),
	}, nil
}

func (s *Server) CalculateTransactionFee(ctx context.Context, request *evmpb.CalculateTransactionFeeRequest) (*evmpb.CalculateTransactionFeeReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	fee, err := evmService.CalculateTransactionFee(ctx, evm.ReceiptGasInfo{
		GasUsed:           request.GasInfo.GasUsed,
		EffectiveGasPrice: valuespb.NewIntFromBigInt(request.GasInfo.EffectiveGasPrice),
	})
	if err != nil {
		return nil, err
	}

	return &evmpb.CalculateTransactionFeeReply{
		TransactionFee: valuespb.NewBigIntFromInt(fee.TransactionFee),
	}, nil
}

func (s *Server) GetForwarderForEOA(ctx context.Context, request *evmpb.GetForwarderForEOARequest) (*evmpb.GetForwarderForEOAReply, error) {
	evmService, err := s.getEVMService(ctx)
	if err != nil {
		return nil, err
	}

	forwarder, err := evmService.GetForwarderForEOA(ctx, evm.Address(request.GetAddr()), evm.Address(request.GetAggr()), request.PluginType)
	if err != nil {
		return nil, err
	}
	return &evmpb.GetForwarderForEOAReply{Addr: forwarder[:]}, nil
}

func (s *Server) getEVMService(ctx context.Context) (types.EVMService, error) {
	id, err := readRelayID(ctx)
	if err != nil {
		return nil, err
	}
	idT := relayerset.RelayerId{Network: id.Network, ChainId: id.ChainID}
	r, err := s.getRelayer(ctx, &idT)
	if err != nil {
		return nil, err
	}

	return r.EVM()
}
