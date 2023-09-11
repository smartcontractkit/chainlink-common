package internal

import (
	"context"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ types.PluginGeneric = (*PluginGenericClient)(nil)

type PluginGenericClient struct {
	*pluginClient
	*serviceClient

	generic pb.PluginGenericClient
}

func NewPluginGenericClient(broker Broker, brokerCfg BrokerConfig, conn *grpc.ClientConn) *PluginGenericClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginGenericClient")
	pc := newPluginClient(broker, brokerCfg, conn)
	return &PluginGenericClient{pluginClient: pc, generic: pb.NewPluginGenericClient(pc), serviceClient: newServiceClient(pc.brokerExt, pc)}
}

func (m *PluginGenericClient) NewGenericServiceFactory(ctx context.Context, pluginConfig []byte, grpcProvider grpc.ClientConnInterface, errorLog types.ErrorLog) (types.ReportingPluginFactory, error) {
	cc := m.newClientConn("GenericPluginFactory", func(ctx context.Context) (id uint32, deps resources, err error) {
		providerID, providerRes, err := m.serve("GenericServiceProvider", proxy.NewProxy(grpcProvider))
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerRes)

		errorLogID, errorLogRes, err := m.serveNew("ErrorLog", func(s *grpc.Server) {
			pb.RegisterErrorLogServer(s, &errorLogServer{impl: errorLog})
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(errorLogRes)

		reply, err := m.generic.NewGenericServiceFactory(ctx, &pb.NewGenericServiceFactoryRequest{
			PluginConfig: pluginConfig,
			ProviderID:   providerID,
			ErrorLogID:   errorLogID,
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ReportingPluginFactoryID, nil, nil
	})
	return newReportingPluginFactoryClient(m.pluginClient.brokerExt, cc), nil
}

var _ pb.PluginGenericServer = (*pluginGenericServer)(nil)

type pluginGenericServer struct {
	pb.UnimplementedPluginGenericServer

	*brokerExt
	impl types.PluginGeneric
}

func RegisterPluginGenericServer(server *grpc.Server, broker Broker, brokerCfg BrokerConfig, impl types.PluginGeneric) error {
	pb.RegisterPluginGenericServer(server, newPluginGenericServer(&brokerExt{broker, brokerCfg}, impl))
	return nil
}

func newPluginGenericServer(b *brokerExt, gp types.PluginGeneric) *pluginGenericServer {
	return &pluginGenericServer{brokerExt: b.withName("PluginGeneric"), impl: gp}
}

func (m *pluginGenericServer) NewGenericServiceFactory(ctx context.Context, request *pb.NewGenericServiceFactoryRequest) (*pb.NewGenericServiceFactoryReply, error) {
	providerConn, err := m.dial(request.ProviderID)
	if err != nil {
		return nil, ErrConnDial{Name: "PluginProvider", ID: request.ProviderID, Err: err}
	}
	providerRes := resource{providerConn, "PluginProvider"}
	provider := newPluginProviderClient(m.brokerExt, providerConn)

	errorLogConn, err := m.dial(request.ErrorLogID)
	if err != nil {
		m.closeAll(providerRes)
		return nil, ErrConnDial{Name: "ErrorLog", ID: request.ErrorLogID, Err: err}
	}
	errorLogRes := resource{errorLogConn, "ErrorLog"}
	errorLog := newErrorLogClient(errorLogConn)

	factory, err := m.impl.NewGenericServiceFactory(ctx, request.PluginConfig, provider.ClientConn(), errorLog)
	if err != nil {
		m.closeAll(providerRes, errorLogRes)
		return nil, err
	}

	id, _, err := m.serveNew("ReportingPluginProvider", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &serviceServer{srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, newReportingPluginFactoryServer(factory, m.brokerExt))
	}, providerRes, errorLogRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewGenericServiceFactoryReply{ReportingPluginFactoryID: id}, nil
}
