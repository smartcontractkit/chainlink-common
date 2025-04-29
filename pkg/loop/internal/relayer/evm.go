package relayer

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.EVMService = (*evmClient)(nil)

type evmClient struct {
	relayer pb.EVMClient
}

func (e *evmClient) GetTransactionFee(ctx context.Context, transactionID string) (*types.TransactionFee, error) {
	reply, err := e.relayer.GetTransactionFee(ctx, &pb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, err
	}

	return &types.TransactionFee{
		TransactionFee: reply.TransationFee.Int(),
	}, nil
}

var _ pb.EVMServer = (*evmServer)(nil)

type evmServer struct {
	pb.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

func (e *evmServer) GetTransactionFee(ctx context.Context, request *pb.GetTransactionFeeRequest) (*pb.GetTransactionFeeReply, error) {
	reply, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}

func newEVMServer(impl types.EVMService, b *net.BrokerExt) *evmServer {
	return &evmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}
