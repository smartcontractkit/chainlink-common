package common

import (
	"context"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/transport"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.ReportingPluginClient = (*ReportingPluginServiceClient)(nil)

type ReportingPluginServiceClient struct {
	*transport.PluginClient
	*transport.ServiceClient
	transport.ResourceManager

	reportingPluginService pb.ReportingPluginServiceClient
}

// needs full power of the private brokerExt
func NewReportingPluginServiceClient(broker transport.Broker, brokerCfg transport.BrokerConfig, conn *grpc.ClientConn) *ReportingPluginServiceClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "ReportingPluginServiceClient")
	pc := transport.NewPluginClient(broker, brokerCfg, conn)
	rm := transport.NewBrokerManager(broker, brokerCfg)
	return &ReportingPluginServiceClient{
		PluginClient:           pc,
		reportingPluginService: pb.NewReportingPluginServiceClient(pc),
		ResourceManager:        transport.NewBrokerManager(broker, brokerCfg),
		ServiceClient:          transport.NewServiceClient(rm, pc)}
}

func (m *ReportingPluginServiceClient) NewReportingPluginFactory(
	ctx context.Context,
	config types.ReportingPluginServiceConfig,
	grpcProvider grpc.ClientConnInterface,
	pipelineRunner types.PipelineRunnerService,
	telemetry types.TelemetryService,
	errorLog types.ErrorLog,
) (types.ReportingPluginFactory, error) {
	cc := transport.NewClientConn("ReportingPluginServiceFactory", m.ResourceManager, m.BrokerConfig, func(ctx context.Context) (id uint32, deps transport.Resources, err error) {
		providerID, providerRes, err := m.Serve("PluginProvider", proxy.NewProxy(grpcProvider))
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerRes)

		pipelineRunnerID, pipelineRunnerRes, err := m.ServeNew("PipelineRunner", func(s *grpc.Server) {
			pb.RegisterPipelineRunnerServiceServer(s, &pipelineRunnerServiceServer{impl: pipelineRunner})
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(pipelineRunnerRes)

		telemetryID, telemetryRes, err := m.ServeNew("Telemetry", func(s *grpc.Server) {
			pb.RegisterTelemetryServer(s, NewTelemetryServer(telemetry))
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(telemetryRes)

		errorLogID, errorLogRes, err := m.ServeNew("ErrorLog", func(s *grpc.Server) {
			pb.RegisterErrorLogServer(s, &errorLogServer{impl: errorLog})
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(errorLogRes)

		reply, err := m.reportingPluginService.NewReportingPluginFactory(ctx, &pb.NewReportingPluginFactoryRequest{
			ReportingPluginServiceConfig: &pb.ReportingPluginServiceConfig{
				ProviderType:  config.ProviderType,
				Command:       config.Command,
				PluginName:    config.PluginName,
				TelemetryType: config.TelemetryType,
				PluginConfig:  config.PluginConfig,
			},
			ProviderID:       providerID,
			ErrorLogID:       errorLogID,
			PipelineRunnerID: pipelineRunnerID,
			TelemetryID:      telemetryID,
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ID, nil, nil
	})
	return newReportingPluginFactoryClient(m.ResourceManager, cc), nil
}

var _ pb.ReportingPluginServiceServer = (*reportingPluginServiceServer)(nil)

type reportingPluginServiceServer struct {
	pb.UnimplementedReportingPluginServiceServer

	//*brokerExt
	transport.ResourceManager
	impl types.ReportingPluginClient
}

func RegisterReportingPluginServiceServer(server *grpc.Server, broker transport.Broker, brokerCfg transport.BrokerConfig, impl types.ReportingPluginClient) error {
	bm := transport.NewBrokerManager(broker, brokerCfg)
	pb.RegisterReportingPluginServiceServer(server, newReportingPluginServiceServer(bm, impl))
	return nil
}

func newReportingPluginServiceServer(b transport.ResourceManager, gp types.ReportingPluginClient) *reportingPluginServiceServer {
	return &reportingPluginServiceServer{ResourceManager: b.New("ReportingPluginService"), impl: gp}
}

func (m *reportingPluginServiceServer) NewReportingPluginFactory(ctx context.Context, request *pb.NewReportingPluginFactoryRequest) (*pb.NewReportingPluginFactoryReply, error) {
	errorLogConn, err := m.Dial(request.ErrorLogID)
	if err != nil {
		return nil, transport.ErrConnDial{Name: "ErrorLog", ID: request.ErrorLogID, Err: err}
	}
	errorLogRes := transport.NewResource(errorLogConn, "ErrorLog")
	errorLog := newErrorLogClient(errorLogConn)

	providerConn, err := m.Dial(request.ProviderID)
	if err != nil {
		m.CloseAll(errorLogRes)
		return nil, transport.ErrConnDial{Name: "PluginProvider", ID: request.ProviderID, Err: err}
	}
	providerRes := transport.NewResource(providerConn, "PluginProvider")

	pipelineRunnerConn, err := m.Dial(request.PipelineRunnerID)
	if err != nil {
		m.CloseAll(errorLogRes, providerRes)
		return nil, transport.ErrConnDial{Name: "PipelineRunner", ID: request.PipelineRunnerID, Err: err}
	}
	pipelineRunnerRes := transport.NewResource(pipelineRunnerConn, "PipelineRunner")
	pipelineRunner := newPipelineRunnerClient(pipelineRunnerConn)

	telemetryConn, err := m.Dial(request.TelemetryID)
	if err != nil {
		m.CloseAll(errorLogRes, providerRes, pipelineRunnerRes)
		return nil, transport.ErrConnDial{Name: "Telemetry", ID: request.TelemetryID, Err: err}
	}
	telemetryRes := transport.NewResource(telemetryConn, "Telemetry")
	telemetry := NewTelemetryServiceClient(telemetryConn)

	config := types.ReportingPluginServiceConfig{
		ProviderType:  request.ReportingPluginServiceConfig.ProviderType,
		PluginConfig:  request.ReportingPluginServiceConfig.PluginConfig,
		PluginName:    request.ReportingPluginServiceConfig.PluginName,
		Command:       request.ReportingPluginServiceConfig.Command,
		TelemetryType: request.ReportingPluginServiceConfig.TelemetryType,
	}

	factory, err := m.impl.NewReportingPluginFactory(ctx, config, providerConn, pipelineRunner, telemetry, errorLog)
	if err != nil {
		m.CloseAll(providerRes, errorLogRes, pipelineRunnerRes, telemetryRes)
		return nil, err
	}

	id, _, err := m.ServeNew("ReportingPluginProvider", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &serviceServer{srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, newReportingPluginFactoryServer(factory, m.ResourceManager))
	}, providerRes, errorLogRes, pipelineRunnerRes, telemetryRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewReportingPluginFactoryReply{ID: id}, nil
}
