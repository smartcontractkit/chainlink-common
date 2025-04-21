package evm

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.EVMChain = (*Client)(nil)

type ClientOpt func(*Client)

type Client struct {
	serviceClient *goplugin.ServiceClient
	grpc          pb.EVMChainClient
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewEVMChainClient(cc),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Start(ctx context.Context) error {
	return c.serviceClient.Start(ctx)
}

func (c *Client) Close() error {
	return c.serviceClient.Close()
}

func (c *Client) Ready() error {
	return c.serviceClient.Ready()
}

func (c *Client) HealthReport() map[string]error {
	return c.serviceClient.HealthReport()
}

func (c *Client) Name() string {
	return c.serviceClient.Name()
}

var _ pb.EVMChainServer = (*Server)(nil)

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedEVMChainServer
	impl types.EVMChain
}

func NewServer(impl types.EVMChain, opts ...ServerOpt) pb.EVMChainServer {
	server := &Server{
		impl: impl,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func RegisterEVMChain(s *grpc.Server, EVMChain types.EVMChain) {
	service := goplugin.ServiceServer{Srv: EVMChain}
	pb.RegisterServiceServer(s, &service)
	pb.RegisterEVMChainServer(s, NewServer(EVMChain))
}

func (c *Client) GetTransactionFee(ctx context.Context, transactionID string) (*types.TransactionFee, error) {
	reply, err := c.grpc.GetTransactionFee(ctx, &pb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &types.TransactionFee{
		TransactionFee: reply.TransationFee.Int(),
	}, nil
}

func (s *Server) GetTransactionFee(ctx context.Context, req *pb.GetTransactionFeeRequest) (*pb.GetTransactionFeeReply, error) {
	reply, err := s.impl.GetTransactionFee(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}
