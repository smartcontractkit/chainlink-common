package internal

import (
	"context"

	"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.ReportingPluginClient = (*ReportingPluginServiceClient)(nil)

type ReportingPluginServiceClient struct {
	*pluginClient
	*serviceClient

	reportingPluginService pb.ReportingPluginServiceClient
}

func NewReportingPluginServiceClient(broker Broker, brokerCfg BrokerConfig, conn *grpc.ClientConn) *ReportingPluginServiceClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "ReportingPluginServiceClient")
	pc := newPluginClient(broker, brokerCfg, conn)
	return &ReportingPluginServiceClient{pluginClient: pc, reportingPluginService: pb.NewReportingPluginServiceClient(pc), serviceClient: newServiceClient(pc.brokerExt, pc)}
}

// NewReportingPluginFactory TODO: save possible errors with the errorLog(?)
func (m *ReportingPluginServiceClient) NewReportingPluginFactory(
	ctx context.Context,
	config types.ReportingPluginServiceConfig,
	grpcProvider grpc.ClientConnInterface,
	pipelineRunner types.PipelineRunnerService,
	telemetry types.TelemetryService,
	errorLog types.ErrorLog,
) (types.ReportingPluginFactory, error) {
	cc := m.newClientConn("ReportingPluginServiceFactory", func(ctx context.Context) (id uint32, deps resources, err error) {
		providerID, providerRes, err := m.serve("PluginProvider", proxy.NewProxy(grpcProvider))
		if err != nil {
			if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
				return 0, nil, errors.Wrap(saveErr, err.Error())
			}
			return 0, nil, err
		}
		deps.Add(providerRes)

		pipelineRunnerID, pipelineRunnerRes, err := m.serveNew("PipelineRunner", func(s *grpc.Server) {
			pb.RegisterPipelineRunnerServiceServer(s, &pipelineRunnerServiceServer{impl: pipelineRunner})
		})
		if err != nil {
			if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
				return 0, nil, errors.Wrap(saveErr, err.Error())
			}
			return 0, nil, err
		}
		deps.Add(pipelineRunnerRes)

		telemetryID, telemetryRes, err := m.serveNew("Telemetry", func(s *grpc.Server) {
			pb.RegisterTelemetryServer(s, NewTelemetryServer(telemetry))
		})
		if err != nil {
			if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
				return 0, nil, errors.Wrap(saveErr, err.Error())
			}
			return 0, nil, err
		}
		deps.Add(telemetryRes)

		errorLogID, errorLogRes, err := m.serveNew("ErrorLog", func(s *grpc.Server) {
			pb.RegisterErrorLogServer(s, &errorLogServer{impl: errorLog})
		})
		if err != nil {
			if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
				return 0, nil, errors.Wrap(saveErr, err.Error())
			}
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
			if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
				return 0, nil, errors.Wrap(saveErr, err.Error())
			}
			return 0, nil, err
		}
		return reply.ID, nil, nil
	})
	return newReportingPluginFactoryClient(m.pluginClient.brokerExt, cc, errorLog), nil
}

var _ pb.ReportingPluginServiceServer = (*reportingPluginServiceServer)(nil)

type reportingPluginServiceServer struct {
	pb.UnimplementedReportingPluginServiceServer

	*brokerExt
	impl types.ReportingPluginClient
}

func RegisterReportingPluginServiceServer(server *grpc.Server, broker Broker, brokerCfg BrokerConfig, impl types.ReportingPluginClient) error {
	pb.RegisterReportingPluginServiceServer(server, newReportingPluginServiceServer(&brokerExt{broker, brokerCfg}, impl))
	return nil
}

func newReportingPluginServiceServer(b *brokerExt, gp types.ReportingPluginClient) *reportingPluginServiceServer {
	return &reportingPluginServiceServer{brokerExt: b.withName("ReportingPluginService"), impl: gp}
}

func (m *reportingPluginServiceServer) NewReportingPluginFactory(ctx context.Context, request *pb.NewReportingPluginFactoryRequest) (*pb.NewReportingPluginFactoryReply, error) {
	errorLogConn, err := m.dial(request.ErrorLogID)
	if err != nil {
		return nil, ErrConnDial{Name: "ErrorLog", ID: request.ErrorLogID, Err: err}
	}
	errorLogRes := resource{errorLogConn, "ErrorLog"}
	errorLog := newErrorLogClient(errorLogConn)

	providerConn, err := m.dial(request.ProviderID)
	if err != nil {
		m.closeAll(errorLogRes)
		err := ErrConnDial{Name: "PluginProvider", ID: request.ProviderID, Err: err}
		if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
			return nil, errors.Wrap(saveErr, err.Error())
		}
		return nil, err
	}
	providerRes := resource{providerConn, "PluginProvider"}

	pipelineRunnerConn, err := m.dial(request.PipelineRunnerID)
	if err != nil {
		m.closeAll(errorLogRes, providerRes)
		err := ErrConnDial{Name: "PipelineRunner", ID: request.PipelineRunnerID, Err: err}
		if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
			return nil, errors.Wrap(saveErr, err.Error())
		}
		return nil, err
	}
	pipelineRunnerRes := resource{pipelineRunnerConn, "PipelineRunner"}
	pipelineRunner := newPipelineRunnerClient(pipelineRunnerConn)

	telemetryConn, err := m.dial(request.TelemetryID)
	if err != nil {
		m.closeAll(errorLogRes, providerRes, pipelineRunnerRes)
		err := ErrConnDial{Name: "Telemetry", ID: request.TelemetryID, Err: err}
		if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
			return nil, errors.Wrap(saveErr, err.Error())
		}
		return nil, err
	}
	telemetryRes := resource{telemetryConn, "Telemetry"}
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
		m.closeAll(providerRes, errorLogRes, pipelineRunnerRes, telemetryRes)
		if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
			return nil, errors.Wrap(saveErr, err.Error())
		}
		return nil, err
	}

	id, _, err := m.serveNew("ReportingPluginProvider", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &serviceServer{srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, newReportingPluginFactoryServer(factory, m.brokerExt, errorLog))
	}, providerRes, errorLogRes, pipelineRunnerRes, telemetryRes)
	if err != nil {
		if saveErr := errorLog.SaveError(ctx, err.Error()); saveErr != nil {
			return nil, errors.Wrap(saveErr, err.Error())
		}
		return nil, err
	}

	return &pb.NewReportingPluginFactoryReply{ID: id}, nil
}
