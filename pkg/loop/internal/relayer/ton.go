package relayer

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	tontypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type TONClient struct {
	grpcClient tonpb.TONClient
}

func NewTONClient(grpcClient tonpb.TONClient) *TONClient {
	return &TONClient{
		grpcClient: grpcClient,
	}
}

var _ types.TONService = (*TONClient)(nil)

// ---- LiteClient ----

func (c *TONClient) GetMasterchainInfo(ctx context.Context) (*tontypes.BlockIDExt, error) {
	block, err := c.grpcClient.GetMasterchainInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return tonpb.ConvertBlockIDExtFromProto(block), nil
}

func (c *TONClient) GetBlockData(ctx context.Context, block *tontypes.BlockIDExt) (*tontypes.Block, error) {
	req := &tonpb.GetBlockDataRequest{Block: tonpb.ConvertBlockIDExtToProto(block)}
	resp, err := c.grpcClient.GetBlockData(ctx, req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &tontypes.Block{
		GlobalID: resp.GlobalId,
	}, nil
}

func (c *TONClient) GetAccountBalance(ctx context.Context, addr string, block *tontypes.BlockIDExt) (*tontypes.Balance, error) {
	req := &tonpb.GetAccountBalanceRequest{
		Address: addr,
		Block:   tonpb.ConvertBlockIDExtToProto(block),
	}
	resp, err := c.grpcClient.GetAccountBalance(ctx, req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	balance := valuespb.NewIntFromBigInt(resp.GetBalance())
	return &tontypes.Balance{
		Balance: balance,
	}, nil

}

// ---- Transaction Management ----

func (c *TONClient) SendTransaction(ctx context.Context, msg tontypes.Message) error {
	req := &tonpb.SendTransactionRequest{
		Message: tonpb.ConvertMessageToProto(&msg),
	}
	_, err := c.grpcClient.SendTransaction(ctx, req)
	return net.WrapRPCErr(err)
}

func (c *TONClient) GetTransactionStatus(ctx context.Context, lt uint64) (types.TransactionStatus, tontypes.ExitCode, error) {
	req := &tonpb.GetTransactionStatusRequest{LogicalTime: lt}
	resp, err := c.grpcClient.GetTransactionStatus(ctx, req)
	if err != nil {
		return types.Unknown, 0, net.WrapRPCErr(err)
	}
	return types.TransactionStatus(resp.Status), *resp.ExitCode, nil
}

func (c *TONClient) GetTransactionExecutionFees(ctx context.Context, lt uint64) (*tontypes.TransactionFee, error) {
	req := &tonpb.GetTransactionExecutionFeesRequest{LogicalTime: lt}
	resp, err := c.grpcClient.GetTransactionExecutionFees(ctx, req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	fee := valuespb.NewIntFromBigInt(resp.GetTotalFees())
	return &tontypes.TransactionFee{
		TransactionFee: fee,
	}, nil
}

// ---- Log Poller ----

func (c *TONClient) HasFilter(ctx context.Context, name string) bool {
	req := &tonpb.HasFilterRequest{Name: name}
	resp, err := c.grpcClient.HasFilter(ctx, req)
	if err != nil {
		return false
	}
	return resp.Exists
}

func (c *TONClient) RegisterFilter(ctx context.Context, filter tontypes.LPFilterQuery) error {
	req := &tonpb.RegisterFilterRequest{
		Filter: tonpb.ConvertLPFilterToProto(filter),
	}
	_, err := c.grpcClient.RegisterFilter(ctx, req)
	return net.WrapRPCErr(err)
}

func (c *TONClient) UnregisterFilter(ctx context.Context, name string) error {
	req := &tonpb.UnregisterFilterRequest{Name: name}
	_, err := c.grpcClient.UnregisterFilter(ctx, req)
	return net.WrapRPCErr(err)
}

type tonServer struct {
	tonpb.UnimplementedTONServer
	*net.BrokerExt
	impl types.TONService
}

var _ tonpb.TONServer = (*tonServer)(nil)

func newTONServer(impl types.TONService, b *net.BrokerExt) *tonServer {
	return &tonServer{impl: impl, BrokerExt: b.WithName("TONServer")}
}

func (s *tonServer) GetMasterchainInfo(ctx context.Context, _ *emptypb.Empty) (*tonpb.BlockIDExt, error) {
	block, err := s.impl.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}
	return tonpb.ConvertBlockIDExtToProto(block), nil
}

func (s *tonServer) GetBlockData(ctx context.Context, req *tonpb.GetBlockDataRequest) (*tonpb.Block, error) {
	block, err := s.impl.GetBlockData(ctx, tonpb.ConvertBlockIDExtFromProto(req.Block))
	if err != nil {
		return nil, err
	}
	return &tonpb.Block{GlobalId: block.GlobalID}, nil
}

func (s *tonServer) GetAccountBalance(ctx context.Context, req *tonpb.GetAccountBalanceRequest) (*tonpb.Balance, error) {
	bal, err := s.impl.GetAccountBalance(ctx, req.Address, tonpb.ConvertBlockIDExtFromProto(req.Block))
	if err != nil {
		return nil, err
	}
	return &tonpb.Balance{Balance: valuespb.NewBigIntFromInt(bal.Balance)}, nil
}

func (s *tonServer) SendTransaction(ctx context.Context, req *tonpb.SendTransactionRequest) (*emptypb.Empty, error) {
	msg := tonpb.ConvertMessageFromProto(req.Message)
	return nil, s.impl.SendTransaction(ctx, *msg)
}

func (s *tonServer) GetTransactionStatus(ctx context.Context, req *tonpb.GetTransactionStatusRequest) (*tonpb.GetTransactionStatusReply, error) {
	status, code, err := s.impl.GetTransactionStatus(ctx, req.LogicalTime)
	if err != nil {
		return nil, err
	}
	return &tonpb.GetTransactionStatusReply{
		Status:   tonpb.TransactionStatus(status),
		ExitCode: &code,
	}, nil
}

func (s *tonServer) GetTransactionExecutionFees(ctx context.Context, req *tonpb.GetTransactionExecutionFeesRequest) (*tonpb.GetTransactionExecutionFeesReply, error) {
	fee, err := s.impl.GetTransactionExecutionFees(ctx, req.LogicalTime)
	if err != nil {
		return nil, err
	}
	return &tonpb.GetTransactionExecutionFeesReply{TotalFees: valuespb.NewBigIntFromInt(fee.TransactionFee)}, nil
}

func (s *tonServer) HasFilter(ctx context.Context, req *tonpb.HasFilterRequest) (*tonpb.HasFilterReply, error) {
	exists := s.impl.HasFilter(ctx, req.Name)
	return &tonpb.HasFilterReply{Exists: exists}, nil
}

func (s *tonServer) RegisterFilter(ctx context.Context, req *tonpb.RegisterFilterRequest) (*emptypb.Empty, error) {
	f := tonpb.ConvertLPFilterFromProto(req.Filter)
	return nil, s.impl.RegisterFilter(ctx, f)
}

func (s *tonServer) UnregisterFilter(ctx context.Context, req *tonpb.UnregisterFilterRequest) (*emptypb.Empty, error) {
	return nil, s.impl.UnregisterFilter(ctx, req.Name)
}
