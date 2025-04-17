package chaincapabilities

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.AptosChainService = (*Client)(nil)

type ClientOpt func(*Client)

type Client struct {
	types.UnimplementedAptosChainService

	serviceClient *goplugin.ServiceClient
	grpc          pb.AptosChainServiceClient
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewAptosChainServiceClient(cc),
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

var _ pb.AptosChainServiceServer = (*Server)(nil)

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedAptosChainServiceServer
	impl types.AptosChainService
}

func NewServer(impl types.AptosChainService, opts ...ServerOpt) pb.AptosChainServiceServer {
	server := &Server{
		impl: impl,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func RegisterAptosChainService(s *grpc.Server, AptosChainService types.AptosChainService) {
	service := goplugin.ServiceServer{Srv: AptosChainService}
	pb.RegisterServiceServer(s, &service)
	pb.RegisterAptosChainServiceServer(s, NewServer(AptosChainService))
}

// ReadContract calls the EVM method, passing encodedParams directly.
func (c *Client) ReadContract(ctx context.Context, method string, encodedParams []byte) ([]byte, error) {
	req := &pb.ReadContractRequest{
		Method:        method,
		EncodedParams: encodedParams,
	}

	resp, err := c.grpc.ReadContract(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// ReadContract handles the EVM RPC call by passing the raw bytes directly.
func (s *Server) ReadContract(ctx context.Context, req *pb.ReadContractRequest) (*pb.ReadContractReply, error) {
	result, err := s.impl.ReadContract(ctx, req.Method, req.EncodedParams)
	if err != nil {
		return nil, fmt.Errorf("ReadContract: %w", err)
	}
	return &pb.ReadContractReply{Result: result}, nil
}
