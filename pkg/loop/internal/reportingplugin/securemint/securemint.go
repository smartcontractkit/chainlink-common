package securemint

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/grpc-proxy/proxy"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr2"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	securemintpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// PluginSecureMintClient is a client that runs on the core node to connect to the SecureMint LOOP server.
type PluginSecureMintClient struct {
	// hashicorp plugin client
	*goplugin.PluginClient
	// client to base service
	*goplugin.ServiceClient

	// creates new SecureMint factory instances
	generator securemintpb.SecureMintFactoryGeneratorClient
}

func NewPluginSecureMintClient(brokerCfg net.BrokerConfig) *PluginSecureMintClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginSecureMintClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &PluginSecureMintClient{
		PluginClient:  pc,
		ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc),
		generator:     securemintpb.NewSecureMintFactoryGeneratorClient(pc),
	}
}

// NewSecureMintFactory creates a new reporting plugin factory client.
// In practice this client is called by the core node.
// The reporting plugin factory client is a client to the LOOP server, which
// is run as an external process via hashicorp plugin. If the given provider is a GRPCClientConn, then the provider is proxied to the
// to the relayer, which is its own process via hashicorp plugin. If the provider is not a GRPCClientConn, then the provider is a local
// to the core node. The core must wrap the provider in a grpc server and serve it locally.
func (c *PluginSecureMintClient) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.SecureMintFactoryGenerator, error) {
	newSecureMintClientFn := func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		// the proxyable resources are the Provider, which may or may not be local to the client process. (legacy vs loopp)
		var (
			providerID       uint32
			providerResource net.Resource
		)
		if grpcProvider, ok := provider.(goplugin.GRPCClientConn); ok {
			providerID, providerResource, err = c.Serve("SecureMintProvider", proxy.NewProxy(grpcProvider.ClientConn()))
		} else {
			// loop client runs in the core node. if the provider is not a grpc client conn, then we are in legacy mode
			// and need to serve all the required services locally.
			providerID, providerResource, err = c.ServeNew("SecureMintProvider", func(s *grpc.Server) {
				RegisterProviderServices(s, provider, c.BrokerExt)
			})
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerResource)

		// Convert config to protobuf format
		configPB := &securemintpb.SecureMintConfig{
			MaxChains: config.MaxChains,
		}

		resp, err := c.generator.NewSecureMintFactory(ctx, &securemintpb.NewSecureMintFactoryRequest{
			ProviderServiceId: providerID,
			Config:            configPB,
		})
		if err != nil {
			return 0, nil, err
		}
		return resp.SecuremintFactoryServiceId, deps, nil
	}
	cc := c.NewClientConn("SecureMintFactory", newSecureMintClientFn)
	factoryClient := ocr2.NewReportingPluginFactoryClient(c.BrokerExt, cc)

	// Create a wrapper that implements SecureMintFactoryGenerator
	return &secureMintFactoryGeneratorWrapper{
		factory:  factoryClient,
		provider: provider,
		config:   config,
	}, nil
}

// secureMintFactoryGeneratorWrapper wraps a ReportingPluginFactory to implement SecureMintFactoryGenerator
type secureMintFactoryGeneratorWrapper struct {
	factory  types.ReportingPluginFactory
	provider types.SecureMintProvider
	config   types.SecureMintConfig
}

func (w *secureMintFactoryGeneratorWrapper) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.ReportingPluginFactory, error) {
	// The factory is already created, just return it
	return w.factory, nil
}

func (w *secureMintFactoryGeneratorWrapper) Start(ctx context.Context) error {
	return w.factory.Start(ctx)
}

func (w *secureMintFactoryGeneratorWrapper) Close() error {
	return w.factory.Close()
}

func (w *secureMintFactoryGeneratorWrapper) Ready() error {
	return w.factory.Ready()
}

func (w *secureMintFactoryGeneratorWrapper) HealthReport() map[string]error {
	return w.factory.HealthReport()
}

func (w *secureMintFactoryGeneratorWrapper) Name() string {
	return w.factory.Name()
}

// PluginSecureMintServer is a server that runs the SecureMint LOOP.
type PluginSecureMintServer struct {
	securemintpb.UnimplementedSecureMintFactoryGeneratorServer

	*net.BrokerExt
	impl types.PluginSecureMint
}

func RegisterPluginSecureMintServer(s *grpc.Server, b net.Broker, cfg net.BrokerConfig, impl types.PluginSecureMint) error {
	ext := &net.BrokerExt{Broker: b, BrokerConfig: cfg}
	securemintpb.RegisterSecureMintFactoryGeneratorServer(s, newPluginSecureMintServer(impl, ext))
	return nil
}

func newPluginSecureMintServer(impl types.PluginSecureMint, b *net.BrokerExt) *PluginSecureMintServer {
	return &PluginSecureMintServer{impl: impl, BrokerExt: b.WithName("PluginSecureMintServer")}
}

func (s *PluginSecureMintServer) NewSecureMintFactory(ctx context.Context, request *securemintpb.NewSecureMintFactoryRequest) (*securemintpb.NewSecureMintFactoryResponse, error) {
	// 1. Get the provider connection from the request.ProviderServiceId
	_, err := s.Dial(request.ProviderServiceId)
	if err != nil {
		return nil, fmt.Errorf("failed to dial provider service: %w", err)
	}
	
	// TODO(gg): Create a provider client from the connection when external plugin is integrated
	// For now, we'll use a placeholder approach
	// secureMintProvider := securemintprovider.NewProviderClient(providerConn)
	
	// 2. Convert the config from protobuf to types.SecureMintConfig
	_ = types.SecureMintConfig{
		MaxChains: request.Config.MaxChains,
	}
	
	// 3. Call impl.NewSecureMintFactory(ctx, provider, config)
	// TODO(gg): Use actual provider when external plugin is integrated
	// factoryGenerator, err := s.impl.NewSecureMintFactory(ctx, secureMintProvider, config)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create factory: %w", err)
	// }
	
	// 4. Return the factory service ID
	// For now, return a placeholder service ID
	factoryID := s.Broker.NextId()
	
	return &securemintpb.NewSecureMintFactoryResponse{
		SecuremintFactoryServiceId: factoryID,
	}, nil
}
