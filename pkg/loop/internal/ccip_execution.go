package internal

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/mwitkow/grpc-proxy/proxy"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	ccipinternal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	cciptypes "github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
)

// ExecutionLOOPClient is a client is run on the core node to connect to the execution LOOP server
type ExecutionLOOPClient struct {
	// hashicorp plugin client
	*PluginClient
	// client to base service
	*ServiceClient

	// creates new execution factory instances
	generator ccippb.ExecutionFactoryGeneratorClient
}

func NewExecutionLOOPClient(broker Broker, brokerCfg BrokerConfig, conn *grpc.ClientConn) *ExecutionLOOPClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "ExecutionLOOPClient")
	pc := NewPluginClient(broker, brokerCfg, conn)
	return &ExecutionLOOPClient{
		PluginClient:  pc,
		ServiceClient: NewServiceClient(pc.BrokerExt, pc),
		generator:     ccippb.NewExecutionFactoryGeneratorClient(pc),
	}
}

// NewExecutionFactory creates a new reporting plugin factory client.
// In practice this client is called by the core node.
// The reporting plugin factory client is a client to the LOOP server, which
// is run as an external process via hashicorp plugin. If the given provider is a GRPCClientConn, then the provider is proxied to the
// to the relayer, which is its own process via hashicorp plugin. If the provider is not a GRPCClientConn, then the provider is a local
// to the core node. The core must wrap the provider in a grpc server and serve it locally.
// func (c *ExecutionLOOPClient) NewExecutionFactory(ctx context.Context, provider types.CCIPExecProvider, config types.CCIPExecFactoryGeneratorConfig) (types.ReportingPluginFactory, error) {
func (c *ExecutionLOOPClient) NewExecutionFactory(ctx context.Context, provider types.CCIPExecProvider) (types.ReportingPluginFactory, error) {
	newExecClientFn := func(ctx context.Context) (id uint32, deps Resources, err error) {
		// TODO are there any local resources that need to be passed to the executor and started as a server?

		// the proxyable resources are the Provider,  which may or may not be local to the client process. (legacy vs loopp)
		var (
			providerID       uint32
			providerResource Resource
		)
		if grpcProvider, ok := provider.(GRPCClientConn); ok {
			// TODO: BCF-3061 ccip provider can create new services. the proxying needs to be augmented
			// to intercept and route to the created services. also, need to prevent leaks.
			providerID, providerResource, err = c.Serve("ExecProvider", proxy.NewProxy(grpcProvider.ClientConn()))
		} else {
			// loop client runs in the core node. if the provider is not a grpc client conn, then we are in legacy mode
			// and need to serve all the required services locally.
			providerID, providerResource, err = c.ServeNew("ExecProvider", func(s *grpc.Server) {
				registerPluginProviderServices(s, provider)
				// TODO LEAKS!!!
				// unlike other products, the provider can create new interface instances, so we need to serve the
				// servers rather than resources
				registerCustomExecutionProviderServices(s, provider, c.BrokerExt)
			})
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerResource)

		resp, err := c.generator.NewExecutionFactory(ctx, &ccippb.NewExecutionFactoryRequest{
			ProviderServiceId: providerID,
		})
		if err != nil {
			return 0, nil, err
		}
		return resp.ExecutionFactoryServiceId, deps, nil
	}
	cc := c.NewClientConn("ExecutionFactory", newExecClientFn)
	return newReportingPluginFactoryClient(c.BrokerExt, cc), nil
}

func registerCustomExecutionProviderServices(s *grpc.Server, provider types.CCIPExecProvider, brokerExt *BrokerExt) {
	// register the handler for the custom methods of the provider eg NewOffRampReader
	ccippb.RegisterExecutionCustomHandlersServer(s, newExecProviderServer(provider, brokerExt))
}

// ExecutionLOOPServer is a server that runs the execution LOOP
type ExecutionLOOPServer struct {
	ccippb.UnimplementedExecutionFactoryGeneratorServer

	*BrokerExt
	impl types.CCIPExecutionFactoryGenerator
}

func RegisterExecutionLOOPServer(s *grpc.Server, b Broker, cfg BrokerConfig, impl types.CCIPExecutionFactoryGenerator) error {
	ext := &BrokerExt{Broker: b, BrokerConfig: cfg}
	ccippb.RegisterExecutionFactoryGeneratorServer(s, newExecutionLOOPServer(impl, ext))
	return nil
}

func newExecutionLOOPServer(impl types.CCIPExecutionFactoryGenerator, b *BrokerExt) *ExecutionLOOPServer {
	return &ExecutionLOOPServer{impl: impl, BrokerExt: b.WithName("ExecutionLOOPServer")}
}

func (r *ExecutionLOOPServer) NewExecutionFactory(ctx context.Context, request *ccippb.NewExecutionFactoryRequest) (*ccippb.NewExecutionFactoryResponse, error) {
	var err error
	var deps Resources
	defer func() {
		if err != nil {
			r.CloseAll(deps...)
		}
	}()

	// lookup the provider service
	providerConn, err := r.Dial(request.ProviderServiceId)
	if err != nil {
		return nil, ErrConnDial{Name: "ExecProvider", ID: request.ProviderServiceId, Err: err}
	}
	deps.Add(Resource{providerConn, "ExecProvider"})
	provider := newExecProviderClient(r.BrokerExt, providerConn)

	// factory, err := r.impl.NewExecutionFactory(ctx, provider, execFactoryConfig(request.Config))
	factory, err := r.impl.NewExecutionFactory(ctx, provider)

	if err != nil {
		return nil, fmt.Errorf("failed to create new execution factory: %w", err)
	}

	id, _, err := r.ServeNew("ExecutionFactory", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &ServiceServer{Srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, newReportingPluginFactoryServer(factory, r.BrokerExt))
	}, deps...)
	if err != nil {
		return nil, fmt.Errorf("failed to serve new execution factory: %w", err)
	}
	return &ccippb.NewExecutionFactoryResponse{ExecutionFactoryServiceId: id}, nil
}

var (
	_ types.CCIPExecProvider = (*execProviderClient)(nil)
	_ GRPCClientConn         = (*execProviderClient)(nil)
)

type execProviderClient struct {
	*pluginProviderClient

	// must be shared with the server
	*BrokerExt
	grpc ccippb.ExecutionCustomHandlersClient
}

// conn
func newExecProviderClient(b *BrokerExt, conn grpc.ClientConnInterface) *execProviderClient {
	pluginProviderClient := newPluginProviderClient(b, conn)
	grpc := ccippb.NewExecutionCustomHandlersClient(conn)
	return &execProviderClient{
		pluginProviderClient: pluginProviderClient,
		BrokerExt:            b,
		grpc:                 grpc,
	}
}

// NewCommitStoreReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewCommitStoreReader(ctx context.Context, addr cciptypes.Address) (cciptypes.CommitStoreReader, error) {
	panic("unimplemented")
}

// NewOffRampReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewOffRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OffRampReader, error) {
	ctx, cancel := e.StopCtx()
	defer cancel()

	req := ccippb.NewOffRampReaderRequest{Address: string(addr)}

	resp, err := e.grpc.NewOffRampReader(ctx, &req)
	if err != nil {
		return nil, err
	}
	// this works because the broker is shared and the id refers to a resource served by the broker
	grpcClient, err := e.BrokerExt.Dial(uint32(resp.OfframpReaderServiceId))

	// need to wrap grpc client into the desired interface
	client := ccipinternal.NewOffRampReaderClient(grpcClient)

	// how to convert resp to cciptypes.OnRampReader? i have an id and need to hydrate that into an instance of OnRampReader
	return client, err
}

// NewOnRampReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewOnRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OnRampReader, error) {
	ctx, cancel := e.StopCtx()
	defer cancel()

	req := ccippb.NewOnRampReaderRequest{Address: string(addr)}

	resp, err := e.grpc.NewOnRampReader(ctx, &req)
	if err != nil {
		return nil, err
	}
	// TODO BCF-3061: make this work for proxied relayer
	// currently this only work for an embedded relayer
	// because the broker is shared  between the core node and relayer
	// this effectively let us proxy connects to resources spawn by the embedded relay
	// by hijacking the shared broker. id refers to a resource served by the shared broker
	grpcClient, err := e.BrokerExt.Dial(uint32(resp.OnrampReaderServiceId))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup on ramp reader service: %w", err)
	}
	// need to wrap grpc client into the desired interface
	client := ccipinternal.NewOnRampReaderClient(grpcClient)

	// how to convert resp to cciptypes.OnRampReader? i have an id and need to hydrate that into an instance of OnRampReader
	return client, err
}

// NewPriceRegistryReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewPriceRegistryReader(ctx context.Context, addr cciptypes.Address) (cciptypes.PriceRegistryReader, error) {
	panic("unimplemented")
}

// NewTokenDataReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewTokenDataReader(ctx context.Context, tokenAddress cciptypes.Address) (cciptypes.TokenDataReader, error) {
	panic("unimplemented")
}

// NewTokenPoolBatchedReader implements types.CCIPExecProvider.
func (e *execProviderClient) NewTokenPoolBatchedReader(ctx context.Context) (cciptypes.TokenPoolBatchedReader, error) {
	panic("unimplemented")
}

// SourceNativeToken implements types.CCIPExecProvider.
func (e *execProviderClient) SourceNativeToken(ctx context.Context) (cciptypes.Address, error) {
	panic("unimplemented")
}

type onRampReaderHandlerClient struct {
	*BrokerExt
	grpc ccippb.ExecutionCustomHandlersClient
}

func newOnRampReaderHandlerClient(b *BrokerExt, conn *grpc.ClientConn) *onRampReaderHandlerClient {
	return &onRampReaderHandlerClient{BrokerExt: b, grpc: ccippb.NewExecutionCustomHandlersClient(conn)}
}

func (c *onRampReaderHandlerClient) NewOnRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OnRampReader, error) {
	ctx, cancel := c.StopCtx()
	defer cancel()

	var req ccippb.NewOnRampReaderRequest

	resp, err := c.grpc.NewOnRampReader(ctx, &req)
	if err != nil {
		return nil, err
	}
	// this works because the id is served from the same broker
	grpcClient, err := c.BrokerExt.Dial(uint32(resp.OnrampReaderServiceId))
	// need to wrap grpc client into the desired interface
	client := ccipinternal.NewOnRampReaderClient(grpcClient)

	// how to convert resp to cciptypes.OnRampReader? i have an id and need to hydrate that into an instance of OnRampReader
	return client, err
}

func execFactoryConfig(config *ccippb.ExecutionFactoryConfig) types.CCIPExecFactoryGeneratorConfig {
	return types.CCIPExecFactoryGeneratorConfig{
		OnRampAddress:      cciptypes.Address(config.OnRampAddress),
		OffRampAddress:     cciptypes.Address(config.OffRampAddress),
		CommitStoreAddress: cciptypes.Address(config.CommitStoreAddress),
		TokenReaderAddress: cciptypes.Address(config.TokenReaderAddress),
	}
}

// execProviderServer is a server that wraps the custom methods of the types.CCIPExecProvider
// this is necessary because those method create new resources that need to be served by the broker
// when we are running in legacy mode
type execProviderServer struct {
	ccippb.UnimplementedExecutionCustomHandlersServer
	// this has to be a shared pointer to the same impl as the client
	*BrokerExt
	impl types.CCIPExecProvider

	deps Resources
}

func newExecProviderServer(impl types.CCIPExecProvider, brokerExt *BrokerExt) *execProviderServer {
	return &execProviderServer{impl: impl, BrokerExt: brokerExt}
}

func (e *execProviderServer) NewOffRampReader(ctx context.Context, req *ccippb.NewOffRampReaderRequest) (*ccippb.NewOffRampReaderResponse, error) {
	reader, err := e.impl.NewOffRampReader(ctx, cciptypes.Address(req.Address))
	if err != nil {
		return nil, err
	}
	// wrap the reader in a grpc server and serve it
	srv := ccipinternal.NewOffRampReaderServer(reader)
	// the id is handle to the broker, we will need it on the other sider to dial the resource
	offRampID, offRampResource, err := e.ServeNew("OffRampReader", func(s *grpc.Server) {
		ccippb.RegisterOffRampReaderServer(s, srv)
	})

	if err != nil {
		return nil, err
	}
	// TODO LEAKS!!!
	// does this need an identifier so a reporting plugin, as a provider client, can hook into closing it?
	e.deps.Add(offRampResource)
	return &ccippb.NewOffRampReaderResponse{OfframpReaderServiceId: int32(offRampID)}, nil
}

func (e *execProviderServer) NewOnRampReader(ctx context.Context, req *ccippb.NewOnRampReaderRequest) (*ccippb.NewOnRampReaderResponse, error) {
	reader, err := e.impl.NewOnRampReader(ctx, cciptypes.Address(req.Address))
	if err != nil {
		return nil, err
	}
	// wrap the reader in a grpc server and serve it
	srv := ccipinternal.NewOnRampReaderServer(reader)
	// the id is handle to the broker, we will need it on the other sider to dial the resource
	onRampID, onRampResource, err := e.ServeNew("OnRampReader", func(s *grpc.Server) {
		ccippb.RegisterOnRampReaderServer(s, srv)
	})

	if err != nil {
		return nil, err
	}
	// TODO LEAKS!!!
	// does this need an identifier so a reporting plugin, as a provider client, can hook into closing it?
	e.deps.Add(onRampResource)
	return &ccippb.NewOnRampReaderResponse{OnrampReaderServiceId: int32(onRampID)}, nil
}
