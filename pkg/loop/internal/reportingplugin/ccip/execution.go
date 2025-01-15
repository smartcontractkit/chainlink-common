package ccip

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/grpc-proxy/proxy"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr2"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	ccipprovider "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ExecutionLOOPClient is a client is run on the core node to connect to the execution LOOP server.
type ExecutionLOOPClient struct {
	// hashicorp plugin client
	*goplugin.PluginClient
	// client to base service
	*goplugin.ServiceClient

	// creates new execution factory instances
	generator ccippb.ExecutionFactoryGeneratorClient
}

func NewExecutionLOOPClient(brokerCfg net.BrokerConfig) *ExecutionLOOPClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "ExecutionLOOPClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &ExecutionLOOPClient{
		PluginClient:  pc,
		ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc),
		generator:     ccippb.NewExecutionFactoryGeneratorClient(pc),
	}
}

// NewExecutionFactory creates a new reporting plugin factory client.
// In practice this client is called by the core node.
// The reporting plugin factory client is a client to the LOOP server, which
// is run as an external process via hashicorp plugin. If the given provider is a GRPCClientConn, then the provider is proxied to the
// to the relayer, which is its own process via hashicorp plugin. If the provider is not a GRPCClientConn, then the provider is a local
// to the core node. The core must wrap the provider in a grpc server and serve it locally.
func (c *ExecutionLOOPClient) NewExecutionFactory(ctx context.Context, srcProvider types.CCIPExecProvider, dstProvider types.CCIPExecProvider, srcChainID int64, dstChainID int64, srcTokenAddress string) (types.ReportingPluginFactory, error) {
	newExecClientFn := func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		// TODO are there any local resources that need to be passed to the executor and started as a server?

		// the proxyable resources are the Provider,  which may or may not be local to the client process. (legacy vs loopp)
		var (
			srcProviderID       uint32
			srcProviderResource net.Resource
			dstProviderID       uint32
			dstProviderResource net.Resource
		)
		if srcGrpcProvider, ok := srcProvider.(goplugin.GRPCClientConn); ok {
			// TODO: BCF-3061 ccip provider can create new services. the proxying needs to be augmented
			// to intercept and route to the created services. also, need to prevent leaks.
			srcProviderID, srcProviderResource, err = c.Serve("ExecProvider", proxy.NewProxy(srcGrpcProvider.ClientConn()))
		} else {
			// loop client runs in the core node. if the provider is not a grpc client conn, then we are in legacy mode
			// and need to serve all the required services locally.
			srcProviderID, srcProviderResource, err = c.ServeNew("ExecProvider", func(s *grpc.Server) {
				ccipprovider.RegisterExecutionProviderServices(s, srcProvider, c.BrokerExt)
			})
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(srcProviderResource)

		if dstGrpcProvider, ok := dstProvider.(goplugin.GRPCClientConn); ok {
			// TODO: BCF-3061 ccip provider can create new services. the proxying needs to be augmented
			// to intercept and route to the created services. also, need to prevent leaks.
			dstProviderID, dstProviderResource, err = c.Serve("ExecProvider", proxy.NewProxy(dstGrpcProvider.ClientConn()))
		} else {
			// loop client runs in the core node. if the provider is not a grpc client conn, then we are in legacy mode
			// and need to serve all the required services locally.
			dstProviderID, dstProviderResource, err = c.ServeNew("ExecProvider", func(s *grpc.Server) {
				ccipprovider.RegisterExecutionProviderServices(s, dstProvider, c.BrokerExt)
			})
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(dstProviderResource)

		resp, err := c.generator.NewExecutionFactory(ctx, &ccippb.NewExecutionFactoryRequest{
			SrcProviderServiceId: srcProviderID,
			DstProviderServiceId: dstProviderID,
			SrcChain:             uint32(srcChainID),
			DstChain:             uint32(dstChainID),
			SrcTokenAddress:      srcTokenAddress,
		})
		if err != nil {
			return 0, nil, err
		}
		return resp.ExecutionFactoryServiceId, deps, nil
	}
	cc := c.NewClientConn("ExecutionFactory", newExecClientFn)
	return ocr2.NewReportingPluginFactoryClient(c.BrokerExt, cc), nil
}

/*
func RegisterExecutionProviderServices(s *grpc.Server, provider types.CCIPExecProvider, brokerExt *net.BrokerExt) {
	// register the handler for the custom methods of the provider eg NewOffRampReader
	ccippb.RegisterExecutionCustomHandlersServer(s, NewExecProviderServer(provider, brokerExt))
}
*/

// ExecutionLOOPServer is a server that runs the execution LOOP.
type ExecutionLOOPServer struct {
	ccippb.UnimplementedExecutionFactoryGeneratorServer

	*net.BrokerExt
	impl types.CCIPExecutionFactoryGenerator
}

func RegisterExecutionLOOPServer(s *grpc.Server, b net.Broker, cfg net.BrokerConfig, impl types.CCIPExecutionFactoryGenerator) error {
	ext := &net.BrokerExt{Broker: b, BrokerConfig: cfg}
	ccippb.RegisterExecutionFactoryGeneratorServer(s, newExecutionLOOPServer(impl, ext))
	return nil
}

func newExecutionLOOPServer(impl types.CCIPExecutionFactoryGenerator, b *net.BrokerExt) *ExecutionLOOPServer {
	return &ExecutionLOOPServer{impl: impl, BrokerExt: b.WithName("ExecutionLOOPServer")}
}

func (r *ExecutionLOOPServer) NewExecutionFactory(ctx context.Context, request *ccippb.NewExecutionFactoryRequest) (*ccippb.NewExecutionFactoryResponse, error) {
	var err error
	var deps net.Resources
	defer func() {
		if err != nil {
			r.CloseAll(deps...)
		}
	}()

	// lookup the source provider service
	srcProviderConn, err := r.Dial(request.SrcProviderServiceId)
	if err != nil {
		return nil, net.ErrConnDial{Name: "ExecProvider", ID: request.SrcProviderServiceId, Err: err}
	}
	deps.Add(net.Resource{Closer: srcProviderConn, Name: "ExecProvider"})
	srcProvider := ccipprovider.NewExecProviderClient(r.BrokerExt, srcProviderConn)

	// lookup the dest provider service
	dstProviderConn, err := r.Dial(request.DstProviderServiceId)
	if err != nil {
		return nil, net.ErrConnDial{Name: "ExecProvider", ID: request.DstProviderServiceId, Err: err}
	}
	deps.Add(net.Resource{Closer: dstProviderConn, Name: "ExecProvider"})
	dstProvider := ccipprovider.NewExecProviderClient(r.BrokerExt, dstProviderConn)

	factory, err := r.impl.NewExecutionFactory(ctx, srcProvider, dstProvider, int64(request.SrcChain), int64(request.DstChain), request.SrcTokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create new execution factory: %w", err)
	}

	id, _, err := r.ServeNew("ExecutionFactory", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, ocr2.NewReportingPluginFactoryServer(factory, r.BrokerExt))
	}, deps...)
	if err != nil {
		return nil, fmt.Errorf("failed to serve new execution factory: %w", err)
	}
	return &ccippb.NewExecutionFactoryResponse{ExecutionFactoryServiceId: id}, nil
}
