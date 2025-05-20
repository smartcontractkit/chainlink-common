package relayerset

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-capabilities/evm/chain-service"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	rel "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

// eVMClient wraps the EVMRelayerSetClient to attach RelayerID to EVMClient request.
type eVMClient struct {
	relayID types.RelayID
	client  relayerset.EVMRelayerSetClient
}

var _ evmpb.EVMClient = (*eVMClient)(nil)

func (e eVMClient) GetTransactionFee(ctx context.Context, in *evmpb.GetTransactionFeeRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionFeeReply, error) {
	return e.client.GetTransactionFee(ctx, &relayerset.GetTransactionFeeRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in},
		opts...)
}

func (e eVMClient) CallContract(ctx context.Context, in *evmpb.CallContractRequest, opts ...grpc.CallOption) (*evmpb.CallContractReply, error) {
	fmt.Println("client is ", e.client)
	return e.client.CallContract(ctx, &relayerset.CallContractRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) FilterLogs(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error) {
	return e.client.FilterLogs(ctx, &relayerset.FilterLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) BalanceAt(ctx context.Context, in *evmpb.BalanceAtRequest, opts ...grpc.CallOption) (*evmpb.BalanceAtReply, error) {
	return e.client.BalanceAt(ctx, &relayerset.BalanceAtRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) EstimateGas(ctx context.Context, in *evmpb.EstimateGasRequest, opts ...grpc.CallOption) (*evmpb.EstimateGasReply, error) {
	return e.client.EstimateGas(ctx, &relayerset.EstimateGasRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) GetTransactionByHash(ctx context.Context, in *evmpb.GetTransactionByHashRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionByHashReply, error) {
	return e.client.GetTransactionByHash(ctx, &relayerset.GetTransactionByHashRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) GetTransactionReceipt(ctx context.Context, in *evmpb.GetTransactionReceiptRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionReceiptReply, error) {
	return e.client.GetTransactionReceipt(ctx, &relayerset.GetTransactionReceiptRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty, opts ...grpc.CallOption) (*evmpb.LatestAndFinalizedHeadReply, error) {
	return e.client.LatestAndFinalizedHead(ctx, &relayerset.LatestHeadRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
	}, opts...)
}

func (e eVMClient) QueryTrackedLogs(ctx context.Context, in *evmpb.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evmpb.QueryTrackedLogsReply, error) {
	return e.client.QueryTrackedLogs(ctx, &relayerset.QueryTrackedLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) RegisterLogTracking(ctx context.Context, in *evmpb.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.RegisterLogTracking(ctx, &relayerset.RegisterLogTrackingRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e eVMClient) UnregisterLogTracking(ctx context.Context, in *evmpb.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
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

func (s *Server) GetTransactionFee(ctx context.Context, request *relayerset.GetTransactionFeeRequest) (*evmpb.GetTransactionFeeReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionFee(ctx, request.Request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionFeeReply{TransationFee: valuespb.NewBigIntFromInt(reply.TransactionFee)}, nil
}

func (s *Server) CallContract(ctx context.Context, request *relayerset.CallContractRequest) (*evmpb.CallContractReply, error) {
	evmService, err := s.getEVMService(ctx, request.RelayerId)
	if err != nil {
		return nil, err
	}

	callMsg, err := rel.ConvertCallMsgFromProto(request.Request.Call)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.CallContract(ctx, callMsg, valuespb.NewIntFromBigInt(request.Request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmpb.CallContractReply{
		Data: &evmpb.ABIPayload{Abi: reply},
	}, nil
}

func (s *Server) FilterLogs(ctx context.Context, request *relayerset.FilterLogsRequest) (*evmpb.FilterLogsReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	expression, err := rel.ConvertFilterFromProto(request.Request.FilterQuery)
	if err != nil {
		return nil, err
	}

	reply, err := evmService.FilterLogs(ctx, expression)
	if err != nil {
		return nil, err
	}

	return &evmpb.FilterLogsReply{Logs: rel.ConvertLogsToProto(reply)}, nil
}

func (s *Server) BalanceAt(ctx context.Context, request *relayerset.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	balance, err := evmService.BalanceAt(ctx, rel.ConvertAddressFromProto(request.GetRequest().GetAccount()), valuespb.NewIntFromBigInt(request.Request.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmpb.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance)}, nil
}

func (s *Server) EstimateGas(ctx context.Context, request *relayerset.EstimateGasRequest) (*evmpb.EstimateGasReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	callMsg, err := rel.ConvertCallMsgFromProto(request.Request.GetMsg())
	if err != nil {
		return nil, err
	}

	gasLimit, err := evmService.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmpb.EstimateGasReply{Gas: gasLimit}, nil
}

func (s *Server) GetTransactionByHash(ctx context.Context, request *relayerset.GetTransactionByHashRequest) (*evmpb.GetTransactionByHashReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionByHash(ctx, rel.ConvertHashFromProto(request.Request.GetHash()))
	if err != nil {
		return nil, err
	}

	tx, err := rel.ConvertTransactionToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionByHashReply{
		Transaction: tx,
	}, nil
}

func (s *Server) GetTransactionReceipt(ctx context.Context, request *relayerset.GetTransactionReceiptRequest) (*evmpb.GetTransactionReceiptReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	reply, err := evmService.GetTransactionReceipt(ctx, rel.ConvertHashFromProto(request.Request.GetHash()))
	if err != nil {
		return nil, err
	}

	receipt, err := rel.ConvertReceiptToProto(reply)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionReceiptReply{
		Receipt: receipt,
	}, nil
}

func (s *Server) LatestAndFinalizedHead(ctx context.Context, request *relayerset.LatestHeadRequest) (*evmpb.LatestAndFinalizedHeadReply, error) {
	evmService, err := s.getEVMService(ctx, request.RelayerId)
	if err != nil {
		return nil, err
	}

	latest, finalized, err := evmService.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb.LatestAndFinalizedHeadReply{
		Latest:    rel.ConvertHeadToProto(latest),
		Finalized: rel.ConvertHeadToProto(finalized),
	}, nil
}

func (s *Server) QueryTrackedLogs(ctx context.Context, request *relayerset.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	expression, err := rel.ConvertExpressionsFromProto(request.GetRequest().GetExpression())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := contractreader.ConvertLimitAndSortFromProto(request.GetRequest().GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	conf, err := contractreader.ConfidenceFromProto(request.GetRequest().GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	logs, err := evmService.QueryTrackedLogs(ctx, expression, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{Logs: rel.ConvertLogsToProto(logs)}, nil
}

func (s *Server) RegisterLogTracking(ctx context.Context, request *relayerset.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	evmService, err := s.getEVMService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	filter, err := rel.ConvertLPFilterFromProto(request.GetRequest().GetFilter())
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
	r, err := s.getRelayer(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.EVM()
}
