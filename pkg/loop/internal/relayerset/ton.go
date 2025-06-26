package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

// tonClient wraps the TONRelayerSetClient to attach RelayerID to TONClient request.
type tonClient struct {
	relayID types.RelayID
	client  relayerset.TONRelayerSetClient
}

var _ tonpb.TONClient = (*tonClient)(nil)

func (t tonClient) GetMasterchainInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*tonpb.BlockIDExt, error) {
	return t.client.GetMasterchainInfo(ctx, &relayerset.TONGetMasterchainInfoRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
	}, opts...)
}

func (t tonClient) GetBlockData(ctx context.Context, in *tonpb.GetBlockDataRequest, opts ...grpc.CallOption) (*tonpb.Block, error) {
	return t.client.GetBlockData(ctx, &relayerset.TONGetBlockDataRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
	}, opts...)
}

func (t tonClient) GetAccountBalance(ctx context.Context, in *tonpb.GetAccountBalanceRequest, opts ...grpc.CallOption) (*tonpb.Balance, error) {
	return t.client.GetAccountBalance(ctx, &relayerset.TONGetAccountBalanceRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) SendTransaction(ctx context.Context, in *tonpb.SendTransactionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.SendTransaction(ctx, &relayerset.TONSendTransactionRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) GetTransactionStatus(ctx context.Context, in *tonpb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*tonpb.GetTransactionStatusReply, error) {
	return t.client.GetTONTransactionStatus(ctx, &relayerset.TONGetTransactionStatusRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) GetTransactionExecutionFees(ctx context.Context, in *tonpb.GetTransactionExecutionFeesRequest, opts ...grpc.CallOption) (*tonpb.GetTransactionExecutionFeesReply, error) {
	return t.client.GetTransactionExecutionFees(ctx, &relayerset.TONGetTransactionExecutionFeesRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) HasFilter(ctx context.Context, in *tonpb.HasFilterRequest, opts ...grpc.CallOption) (*tonpb.HasFilterReply, error) {
	return t.client.HasFilter(ctx, &relayerset.TONHasFilterRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) RegisterFilter(ctx context.Context, in *tonpb.RegisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.RegisterFilter(ctx, &relayerset.TONRegisterFilterRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (t tonClient) UnregisterFilter(ctx context.Context, in *tonpb.UnregisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.UnregisterFilter(ctx, &relayerset.TONUnregisterFilterRequest{
		RelayerId: &relayerset.RelayerId{
			Network: t.relayID.Network,
			ChainId: t.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (s *Server) GetMasterchainInfo(ctx context.Context, request *relayerset.TONGetMasterchainInfoRequest) (*tonpb.BlockIDExt, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	blockIdExt, err := tonService.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &tonpb.BlockIDExt{Workchain: blockIdExt.Workchain, Shard: blockIdExt.Shard, Seqno: blockIdExt.Seqno}, nil
}

func (s *Server) GetBlockData(ctx context.Context, request *relayerset.TONGetBlockDataRequest) (*tonpb.Block, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	blockIdExt := tonpb.ConvertBlockIDExtFromProto(request.Block)
	block, err := tonService.GetBlockData(ctx, blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.ConvertBlockToProto(block), nil
}

func (s *Server) GetAccountBalance(ctx context.Context, request *relayerset.TONGetAccountBalanceRequest) (*tonpb.Balance, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	blockIdExt := tonpb.ConvertBlockIDExtFromProto(request.Request.GetBlock())
	balance, err := tonService.GetAccountBalance(ctx, request.Request.GetAddress(), blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.ConvertBalanceToProto(balance), nil
}

func (s *Server) SendTransaction(ctx context.Context, request *relayerset.TONSendTransactionRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	msg := tonpb.ConvertMessageFromProto(request.Request.GetMessage())
	err = tonService.SendTransaction(ctx, *msg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetTONTransactionStatus(ctx context.Context, request *relayerset.TONGetTransactionStatusRequest) (*tonpb.GetTransactionStatusReply, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	lt := request.Request.GetLogicalTime()
	txStatus, exitCode, err := tonService.GetTransactionStatus(ctx, lt)
	if err != nil {
		return nil, err
	}

	return &tonpb.GetTransactionStatusReply{
		Status:   tonpb.TransactionStatus(txStatus),
		ExitCode: &exitCode,
	}, nil
}

func (s *Server) GetTransactionExecutionFees(ctx context.Context, request *relayerset.TONGetTransactionExecutionFeesRequest) (*tonpb.GetTransactionExecutionFeesReply, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	lt := request.Request.GetLogicalTime()
	fee, err := tonService.GetTransactionExecutionFees(ctx, lt)
	if err != nil {
		return nil, err
	}

	return &tonpb.GetTransactionExecutionFeesReply{
		TotalFees: pb.NewBigIntFromInt(fee.TransactionFee),
	}, nil
}

func (s *Server) HasFilter(ctx context.Context, request *relayerset.TONHasFilterRequest) (*tonpb.HasFilterReply, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	name := request.Request.GetName()
	exists := tonService.HasFilter(ctx, name)
	if err != nil {
		return nil, err
	}

	return &tonpb.HasFilterReply{
		Exists: exists,
	}, nil
}

func (s *Server) RegisterFilter(ctx context.Context, request *relayerset.TONRegisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	lpFilterQuery := tonpb.ConvertLPFilterFromProto(request.Request.GetFilter())
	err = tonService.RegisterFilter(ctx, lpFilterQuery)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UnregisterFilter(ctx context.Context, request *relayerset.TONUnregisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx, request.GetRelayerId())
	if err != nil {
		return nil, err
	}

	name := request.Request.GetName()
	err = tonService.UnregisterFilter(ctx, name)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) getTONService(ctx context.Context, id *relayerset.RelayerId) (types.TONService, error) {
	r, err := s.getRelayer(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.TON()
}
