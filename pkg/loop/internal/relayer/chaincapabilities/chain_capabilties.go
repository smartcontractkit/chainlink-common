package chaincapabilities

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.ChainCapabilities = (*Client)(nil)

type ClientOpt func(*Client)

type Client struct {
	types.UnimplementedSolanaChainReader
	types.UnimplementedEVMChainReader

	serviceClient *goplugin.ServiceClient
	grpc          pb.ChainCapabilitiesClient
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewChainCapabilitiesClient(cc),
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

var _ pb.ChainCapabilitiesServer = (*Server)(nil)

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedChainCapabilitiesServer
	impl types.ChainCapabilities
}

func NewServer(impl types.ChainCapabilities, opts ...ServerOpt) pb.ChainCapabilitiesServer {
	server := &Server{
		impl: impl,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func RegisterChainCapabilitiesService(s *grpc.Server, chainCapabilities types.ChainCapabilities) {
	service := goplugin.ServiceServer{Srv: chainCapabilities}
	pb.RegisterServiceServer(s, &service)
	pb.RegisterChainCapabilitiesServer(s, NewServer(chainCapabilities))
}
