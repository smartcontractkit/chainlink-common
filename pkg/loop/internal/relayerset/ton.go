package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// tonClient wraps the TONRelayerSetClient by attaching a RelayerID to TONClient requests.
// The attached RelayerID is stored in the context metadata.
type tonClient struct {
	relayID types.RelayID
	client  tonpb.TONClient
}

var _ tonpb.TONClient = (*tonClient)(nil)

func (t tonClient) GetMasterchainInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*tonpb.BlockIDExt, error) {
	return t.client.GetMasterchainInfo(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetBlockData(ctx context.Context, in *tonpb.GetBlockDataRequest, opts ...grpc.CallOption) (*tonpb.Block, error) {
	return t.client.GetBlockData(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetAccountBalance(ctx context.Context, in *tonpb.GetAccountBalanceRequest, opts ...grpc.CallOption) (*tonpb.Balance, error) {
	return t.client.GetAccountBalance(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) SendTx(ctx context.Context, in *tonpb.SendTxRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.SendTx(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetTxStatus(ctx context.Context, in *tonpb.GetTxStatusRequest, opts ...grpc.CallOption) (*tonpb.GetTxStatusReply, error) {
	return t.client.GetTxStatus(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetTxExecutionFees(ctx context.Context, in *tonpb.GetTxExecutionFeesRequest, opts ...grpc.CallOption) (*tonpb.GetTxExecutionFeesReply, error) {
	return t.client.GetTxExecutionFees(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) HasFilter(ctx context.Context, in *tonpb.HasFilterRequest, opts ...grpc.CallOption) (*tonpb.HasFilterReply, error) {
	return t.client.HasFilter(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) RegisterFilter(ctx context.Context, in *tonpb.RegisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.RegisterFilter(appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) UnregisterFilter(ctx context.Context, in *tonpb.UnregisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.UnregisterFilter(appendRelayID(ctx, t.relayID), in, opts...)
}

type tonServer struct {
	tonpb.UnimplementedTONServer
	parent *Server
}

var _ tonpb.TONServer = (*tonServer)(nil)

func (ts *tonServer) GetMasterchainInfo(ctx context.Context, request *emptypb.Empty) (*tonpb.BlockIDExt, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt, err := tonService.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &tonpb.BlockIDExt{Workchain: blockIdExt.Workchain, Shard: blockIdExt.Shard, SeqNo: blockIdExt.SeqNo}, nil
}

func (ts *tonServer) GetBlockData(ctx context.Context, request *tonpb.GetBlockDataRequest) (*tonpb.Block, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt := request.GetBlock().AsBlockIDExt()
	block, err := tonService.GetBlockData(ctx, blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.NewBlock(block), nil
}

func (ts *tonServer) GetAccountBalance(ctx context.Context, request *tonpb.GetAccountBalanceRequest) (*tonpb.Balance, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt := request.GetBlock().AsBlockIDExt()
	balance, err := tonService.GetAccountBalance(ctx, request.GetAddress(), blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.NewBalance(balance), nil
}

func (ts *tonServer) SendTx(ctx context.Context, request *tonpb.SendTxRequest) (*emptypb.Empty, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	msg := request.GetMessage().AsMessage()
	err = tonService.SendTx(ctx, *msg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (ts *tonServer) GetTxStatus(ctx context.Context, request *tonpb.GetTxStatusRequest) (*tonpb.GetTxStatusReply, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	lt := request.GetLogicalTime()
	txStatus, exitCode, err := tonService.GetTxStatus(ctx, lt)
	if err != nil {
		return nil, err
	}

	return &tonpb.GetTxStatusReply{
		Status:   tonpb.TransactionStatus(txStatus),
		ExitCode: &exitCode,
	}, nil
}

func (ts *tonServer) GetTxExecutionFees(ctx context.Context, request *tonpb.GetTxExecutionFeesRequest) (*tonpb.GetTxExecutionFeesReply, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	lt := request.GetLogicalTime()
	fee, err := tonService.GetTxExecutionFees(ctx, lt)
	if err != nil {
		return nil, err
	}

	return &tonpb.GetTxExecutionFeesReply{
		TotalFees: pb.NewBigIntFromInt(fee.TransactionFee),
	}, nil
}

func (ts *tonServer) HasFilter(ctx context.Context, request *tonpb.HasFilterRequest) (*tonpb.HasFilterReply, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	name := request.GetName()
	exists := tonService.HasFilter(ctx, name)

	return &tonpb.HasFilterReply{
		Exists: exists,
	}, nil
}

func (ts *tonServer) RegisterFilter(ctx context.Context, request *tonpb.RegisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	lpFilterQuery := request.GetFilter().AsLPFilter()
	err = tonService.RegisterFilter(ctx, lpFilterQuery)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (ts *tonServer) UnregisterFilter(ctx context.Context, request *tonpb.UnregisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := ts.parent.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	name := request.GetName()
	err = tonService.UnregisterFilter(ctx, name)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) getTONService(ctx context.Context) (types.TONService, error) {
	id, err := readRelayID(ctx)
	if err != nil {
		return nil, err
	}
	idT := relayerset.RelayerId{Network: id.Network, ChainId: id.ChainID}
	r, err := s.getRelayer(ctx, &idT)
	if err != nil {
		return nil, err
	}

	return r.TON()
}
