package ccip

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/mwitkow/grpc-proxy/proxy"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/ocr2"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
)

// ExecutionAdapterClient is a client is run on the core node to connect to the execution adapter server,
// which is run as a loop.
type ExecutionAdapterClient struct {
	*internal.PluginClient
	*internal.ServiceClient

	executorGenerator ccippb.ExecutionFactoryGeneratorClient
}

func NewExecutionAdapterClient(broker internal.Broker, brokerCfg internal.BrokerConfig, conn *grpc.ClientConn) *ExecutionAdapterClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "ExecutionAdapterClient")
	pc := internal.NewPluginClient(broker, brokerCfg, conn)
	return &ExecutionAdapterClient{
		PluginClient:      pc,
		ServiceClient:     internal.NewServiceClient(pc.BrokerExt, pc),
		executorGenerator: ccippb.NewExecutionFactoryGeneratorClient(pc),
	}
}

func (c *ExecutionAdapterClient) NewExecutionFactory(ctx context.Context, provider types.CCIPExecProvider) (types.ReportingPluginFactory, error) {
	newExecClientFn := func(ctx context.Context) (id uint32, deps internal.Resources, err error) {
		// TODO are there any local resources that need to be passed to the executor and started as a server?

		// the proxyable resources are the Provider,  which may or may not be local to the client process. (legacy vs loopp)
		var (
			providerID       uint32
			providerResource internal.Resource
			providerErr      error
		)
		if grpcProvider, ok := provider.(internal.GRPCClientConn); ok {
			providerID, providerResource, err = c.Serve("ExecProvider", proxy.NewProxy(grpcProvider.ClientConn()))
		} else {
			providerID, providerResource, err = c.ServeNew("ExecProvider", func(s *grpc.Server) {
				var onRamp ccip.OnRampReader
				onRamp, providerErr = provider.OnRampReader(ctx)
				if providerErr != nil {
					return
				}
				ccippb.RegisterOnRampReaderServer(s, &OnRampReaderServer{impl: onRamp})

				var offRamp ccip.OffRampReader
				offRamp, providerErr = provider.OffRampReader(ctx)
				if providerErr != nil {
					return
				}
				ccippb.RegisterOffRampReaderServer(s, &OffRampReaderServer{impl: offRamp})
				// TODO: add the rest of the methods
			})
		}
		if providerErr != nil {
			return 0, nil, fmt.Errorf("failed to serve exec provider due to provider error:%w", providerErr)
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerResource)

		resp, err := c.executorGenerator.NewExecutionFactory(ctx, &ccippb.NewExecutionFactoryRequest{
			ProviderServiceId: providerID,
		})
		if err != nil {
			return 0, nil, err
		}
		return uint32(resp.ExecutionFactoryServiceId), deps, nil
	}
	cc := c.NewClientConn("ExecutionFactory", newExecClientFn)
	return ocr2.NewReportingPluginFactoryClient(c.BrokerExt, cc), nil
}

type ExecutionAdapterServer struct {
	ccippb.UnimplementedExecutionFactoryGeneratorServer

	*internal.BrokerExt
	impl types.CCIPExecFactoryGenerator
}

func RegisterExecutionAdapterServer(s *grpc.Server, impl types.CCIPExecFactoryGenerator, b *internal.BrokerExt) {
	ccippb.RegisterExecutionFactoryGeneratorServer(s, &ExecutionAdapterServer{BrokerExt: b, impl: impl})
}

func newExecutionAdapterServer(impl types.CCIPExecFactoryGenerator, b *internal.BrokerExt) *ExecutionAdapterServer {
	return &ExecutionAdapterServer{impl: impl, BrokerExt: b.WithName("ExecutionAdapterServer")}
}

func (r *ExecutionAdapterServer) NewExecutionFactory(ctx context.Context, request *ccippb.NewExecutionFactoryRequest) (*ccippb.NewExecutionFactoryResponse, error) {
	var resources internal.Resources

	providerConn, err := r.Dial(request.ProviderServiceId)
	if err != nil {
		return nil, internal.ErrConnDial{Name: "ExecProvider", ID: request.ProviderServiceId, Err: err}
	}
	resources.Add(internal.Resource{providerConn, "ExecProvider"})
	provider := newExecProviderClient(r.BrokerExt, providerConn)

	factory, err := r.impl.NewExecutionFactory(ctx, provider)
	if err != nil {
		m.CloseAll(resources...)
		return nil, fmt.Errorf("failed to create new execution factory: %w", err)
	}

	id, _, err := r.ServeNew("ExecutionFactory", func(s *grpc.Server) {
		pb.RegisterServiceServer(r, &internal.ServiceServer{Srv: factory})
		pb.RegisterPluginRelayerServer(s, ocr2.NewReportingPluginFactoryServer(factory, r.BrokerExt))
	}, resources...)
	if err != nil {
		r.CloseAll(resources...)
		return nil, fmt.Errorf("failed to serve new execution factory: %w", err)
	}
	return &ccippb.NewExecutionFactoryResponse{ExecutionFactoryServiceId: uint32(id)}, nil
}

type execProviderClient struct {
}
