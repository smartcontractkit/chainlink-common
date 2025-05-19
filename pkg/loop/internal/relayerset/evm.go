package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	evmservice "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	rel "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

// eVMClient wraps the EVMRelayerSetClient to attach RelayerID to EVMClient request.
type eVMClient struct {
	relayID types.RelayID
	client  relayerset.EVMRelayerSetClient
}

var _ evmservice.EVMClient = (*eVMClient)(nil)

func (e eVMClient) GetTransactionFee(ctx context.Context, in *evmservice.GetTransactionFeeRequest, opts ...grpc.CallOption) (*evmservice.GetTransactionFeeReply, error) {
	return e.client.GetTransactionFee(ctx, &relayerset.GetTransactionFeeRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in},
		opts...)
}

func (e eVMClient) CallContract(ctx context.Context, in *evmservice.CallContractRequest, opts ...grpc.CallOption) (*evmservice.CallContractReply, error) {
	return e.client.CallContract(ctx, &relayerset.CallContractRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) FilterLogs(ctx context.Context, in *evmservice.FilterLogsRequest, opts ...grpc.CallOption) (*evmservice.FilterLogsReply, error) {
	return e.client.FilterLogs(ctx, &relayerset.FilterLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) BalanceAt(ctx context.Context, in *evmservice.BalanceAtRequest, opts ...grpc.CallOption) (*evmservice.BalanceAtReply, error) {
	return e.client.BalanceAt(ctx, &relayerset.BalanceAtRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) EstimateGas(ctx context.Context, in *evmservice.EstimateGasRequest, opts ...grpc.CallOption) (*evmservice.EstimateGasReply, error) {
	return e.client.EstimateGas(ctx, &relayerset.EstimateGasRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) TransactionByHash(ctx context.Context, in *evmservice.TransactionByHashRequest, opts ...grpc.CallOption) (*evmservice.TransactionByHashReply, error) {
	return e.client.TransactionByHash(ctx, &relayerset.TransactionByHashRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) TransactionReceipt(ctx context.Context, in *evmservice.TransactionReceiptRequest, opts ...grpc.CallOption) (*evmservice.TransactionReceiptReply, error) {
	return e.client.TransactionReceipt(ctx, &relayerset.ReceiptRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty, opts ...grpc.CallOption) (*evmservice.LatestAndFinalizedHeadReply, error) {
	return e.client.LatestAndFinalizedHead(ctx, &relayerset.LatestHeadRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
	}, opts...)
}

func (e eVMClient) QueryTrackedLogs(ctx context.Context, in *evmservice.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evmservice.QueryTrackedLogsReply, error) {
	return e.client.QueryTrackedLogs(ctx, &relayerset.QueryTrackedLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) RegisterLogTracking(ctx context.Context, in *evmservice.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.RegisterLogTracking(ctx, &relayerset.RegisterLogTrackingRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) UnregisterLogTracking(ctx context.Context, in *evmservice.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.UnregisterLogTracking(ctx, &relayerset.UnregisterLogTrackingRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) GetTransactionStatus(ctx context.Context, in *pb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*pb.GetTransactionStatusReply, error) {
	return e.client.GetTransactionStatus(ctx, &relayerset.GetTransactionStatusRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (s *Server) GetTransactionFee(ctx context.Context, request *relayerset.GetTransactionFeeRequest) (*evmservice.GetTransactionFeeReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionFee(ctx, request.Request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmservice.GetTransactionFeeReply{TransationFee: valuespb.NewBigIntFromInt(reply.TransactionFee)}, nil
}

func (s *Server) CallContract(ctx context.Context, request *relayerset.CallContractRequest) (*evmservice.CallContractReply, error) {
	evmService, err := s.getEVMService(ctx, request.RelayerId)
	if err != nil {
		return nil, err
	}

	callMsg, err := rel.ProtoToCallMsg(request.Request.Call)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.CallContract(ctx, callMsg, valuespb.NewIntFromBigInt(request.Request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmservice.CallContractReply{
		Data: &evmservice.ABIPayload{Abi: reply},
	}, nil
}

func (s *Server) FilterLogs(ctx context.Context, request *relayerset.FilterLogsRequest) (*evmservice.FilterLogsReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	expression, err := rel.ProtoToEvmFilter(request.Request.FilterQuery)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.FilterLogs(ctx, expression)
	if err != nil {
		return nil, err
	}

	return &evmservice.FilterLogsReply{Logs: rel.LogsToProto(reply)}, nil
}

func (s *Server) BalanceAt(ctx context.Context, request *relayerset.BalanceAtRequest) (*evmservice.BalanceAtReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	balance, err := evmService.BalanceAt(ctx, rel.ProtoToAddress(request.GetRequest().GetAccount()), valuespb.NewIntFromBigInt(request.Request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmservice.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance)}, nil
}

func (s *Server) EstimateGas(ctx context.Context, request *relayerset.EstimateGasRequest) (*evmservice.EstimateGasReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	callMsg, err := rel.ProtoToCallMsg(request.Request.GetMsg())
	if err != nil {
		return nil, err
	}

	gasLimit, err := evmService.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmservice.EstimateGasReply{Gas: gasLimit}, nil
}

func (s *Server) TransactionByHash(ctx context.Context, request *relayerset.TransactionByHashRequest) (*evmservice.TransactionByHashReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.TransactionByHash(ctx, rel.ProtoToHash(request.Request.GetHash()))
	if err != nil {
		return nil, err
	}

	tx, err := rel.TransactionToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmservice.TransactionByHashReply{
		Transaction: tx,
	}, nil
}

func (s *Server) TransactionReceipt(ctx context.Context, request *relayerset.ReceiptRequest) (*evmservice.TransactionReceiptReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.TransactionReceipt(ctx, rel.ProtoToHash(request.Request.GetHash()))
	if err != nil {
		return nil, err
	}

	receipt, err := rel.ReceiptToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmservice.TransactionReceiptReply{
		Receipt: receipt,
	}, nil
}

func (s *Server) LatestAndFinalizedHead(ctx context.Context, request *relayerset.LatestHeadRequest) (*evmservice.LatestAndFinalizedHeadReply, error) {
	evmService, err := s.getEVMService(ctx, request.RelayerId)
	if err != nil {
		return nil, err
	}

	latest, finalized, err := evmService.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmservice.LatestAndFinalizedHeadReply{
		Latest:    rel.HeadToProto(latest),
		Finalized: rel.HeadToProto(finalized),
	}, nil
}

func (s *Server) QueryTrackedLogs(ctx context.Context, request *relayerset.QueryTrackedLogsRequest) (*evmservice.QueryTrackedLogsReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	expression, err := rel.ProtoToExpressions(request.GetRequest().GetExpression())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := evmservice.ConvertLimitAndSortFromProto(request.GetRequest().GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	conf, err := evmservice.ConfidenceFromProto(request.GetRequest().GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	logs, err := evmService.QueryTrackedLogs(ctx, expression, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmservice.QueryTrackedLogsReply{Logs: rel.LogsToProto(logs)}, nil
}

func (s *Server) RegisterLogTracking(ctx context.Context, request *relayerset.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	filter, err := rel.ProtoToLpFilter(request.GetRequest().GetFilter())
	if err != nil {
		return nil, err
	}

	if err = evmService.RegisterLogTracking(ctx, filter); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UnregisterLogTracking(ctx context.Context, request *relayerset.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	if err = evmService.UnregisterLogTracking(ctx, request.GetRequest().GetFilterName()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetTransactionStatus(ctx context.Context, request *relayerset.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	txStatus, err := evmService.GetTransactionStatus(ctx, request.Request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(txStatus)}, nil
}

func (s *Server) getEVMService(ctx context.Context, id *relayerset.RelayerId) (types.EVMService, error) {
	rel, err := s.getRelayer(ctx, id)
	if err != nil {
		return nil, err
	}

	return rel.EVM()
}
