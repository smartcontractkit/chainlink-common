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

const metadataTONChain = "ton-chain-id"
const metadataTONNetwork = "ton-network"

// tonClient wraps the TONRelayerSetClient by attaching a RelayerID to TONClient requests.
// The attached RelayerID is stored in the context metadata.
type tonClient struct {
	relayID types.RelayID
	client  tonpb.TONClient
}

var _ tonpb.TONClient = (*tonClient)(nil)

var tonMetadata = relayerMetadata{
	chain:   metadataTONChain,
	network: metadataTONNetwork,
}

func (t tonClient) GetMasterchainInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*tonpb.BlockIDExt, error) {
	return t.client.GetMasterchainInfo(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetBlockData(ctx context.Context, in *tonpb.GetBlockDataRequest, opts ...grpc.CallOption) (*tonpb.Block, error) {
	return t.client.GetBlockData(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetAccountBalance(ctx context.Context, in *tonpb.GetAccountBalanceRequest, opts ...grpc.CallOption) (*tonpb.Balance, error) {
	return t.client.GetAccountBalance(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) SendTx(ctx context.Context, in *tonpb.SendTxRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.SendTx(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetTxStatus(ctx context.Context, in *tonpb.GetTxStatusRequest, opts ...grpc.CallOption) (*tonpb.GetTxStatusReply, error) {
	return t.client.GetTxStatus(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) GetTxExecutionFees(ctx context.Context, in *tonpb.GetTxExecutionFeesRequest, opts ...grpc.CallOption) (*tonpb.GetTxExecutionFeesReply, error) {
	return t.client.GetTxExecutionFees(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) HasFilter(ctx context.Context, in *tonpb.HasFilterRequest, opts ...grpc.CallOption) (*tonpb.HasFilterReply, error) {
	return t.client.HasFilter(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) RegisterFilter(ctx context.Context, in *tonpb.RegisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.RegisterFilter(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (t tonClient) UnregisterFilter(ctx context.Context, in *tonpb.UnregisterFilterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.client.UnregisterFilter(tonMetadata.appendRelayID(ctx, t.relayID), in, opts...)
}

func (s *Server) GetMasterchainInfo(ctx context.Context, request *emptypb.Empty) (*tonpb.BlockIDExt, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt, err := tonService.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &tonpb.BlockIDExt{Workchain: blockIdExt.Workchain, Shard: blockIdExt.Shard, Seqno: blockIdExt.Seqno}, nil
}

func (s *Server) GetBlockData(ctx context.Context, request *tonpb.GetBlockDataRequest) (*tonpb.Block, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt := tonpb.ConvertBlockIDExtFromProto(request.GetBlock())
	block, err := tonService.GetBlockData(ctx, blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.ConvertBlockToProto(block), nil
}

func (s *Server) GetAccountBalance(ctx context.Context, request *tonpb.GetAccountBalanceRequest) (*tonpb.Balance, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	blockIdExt := tonpb.ConvertBlockIDExtFromProto(request.GetBlock())
	balance, err := tonService.GetAccountBalance(ctx, request.GetAddress(), blockIdExt)
	if err != nil {
		return nil, err
	}

	return tonpb.ConvertBalanceToProto(balance), nil
}

func (s *Server) SendTx(ctx context.Context, request *tonpb.SendTxRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	msg := tonpb.ConvertMessageFromProto(request.GetMessage())
	err = tonService.SendTx(ctx, *msg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetTxStatus(ctx context.Context, request *tonpb.GetTxStatusRequest) (*tonpb.GetTxStatusReply, error) {
	tonService, err := s.getTONService(ctx)
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

func (s *Server) GetTxExecutionFees(ctx context.Context, request *tonpb.GetTxExecutionFeesRequest) (*tonpb.GetTxExecutionFeesReply, error) {
	tonService, err := s.getTONService(ctx)
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

func (s *Server) HasFilter(ctx context.Context, request *tonpb.HasFilterRequest) (*tonpb.HasFilterReply, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	name := request.GetName()
	exists := tonService.HasFilter(ctx, name)
	if err != nil {
		return nil, err
	}

	return &tonpb.HasFilterReply{
		Exists: exists,
	}, nil
}

func (s *Server) RegisterFilter(ctx context.Context, request *tonpb.RegisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx)
	if err != nil {
		return nil, err
	}

	lpFilterQuery := tonpb.ConvertLPFilterFromProto(request.GetFilter())
	err = tonService.RegisterFilter(ctx, lpFilterQuery)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UnregisterFilter(ctx context.Context, request *tonpb.UnregisterFilterRequest) (*emptypb.Empty, error) {
	tonService, err := s.getTONService(ctx)
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
	id, err := tonMetadata.readRelayID(ctx)
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
