package securemint

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

var _ pb.ReportingPluginServiceServer = (*pluginSecureMintServer)(nil)

type pluginSecureMintServer struct {
	pb.UnimplementedReportingPluginServiceServer

	*net.BrokerExt
	impl core.PluginSecureMint
}

func (m *pluginSecureMintServer) NewValidationService(ctx context.Context, request *pb.ValidationServiceRequest) (*pb.ValidationServiceResponse, error) {
	m.Logger.Infof("NewValidationService called, not implemented")
	return &pb.ValidationServiceResponse{}, nil
}

func (m *pluginSecureMintServer) NewReportingPluginFactory(ctx context.Context, request *pb.NewReportingPluginFactoryRequest) (*pb.NewReportingPluginFactoryReply, error) {
	m.Logger.Infof("NewReportingPluginFactory called, delegating to impl.NewSecureMintFactory")

	externalAdapterConn, err := m.Dial(request.PipelineRunnerID) // TODO(gg): misusing pipeline runner id here, should be ExternalAdapterID
	if err != nil {
		return nil, net.ErrConnDial{Name: "ExternalAdapter", ID: request.PipelineRunnerID, Err: err}
	}
	externalAdapterRes := net.Resource{Closer: externalAdapterConn, Name: "ExternalAdapter"}
	externalAdapter := newExternalAdapterClient(m.Logger, externalAdapterConn)

	reportingPluginFactory, err := m.impl.NewSecureMintFactory(ctx, m.Logger, externalAdapter)
	if err != nil {
		m.CloseAll(externalAdapterRes)
		return nil, err
	}

	id, _, err := m.ServeNew("ReportingPluginProvider", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: reportingPluginFactory})
		ocr3pb.RegisterReportingPluginFactoryServer(s, ocr3.NewReportingPluginFactoryServer(&reportingPluginFactoryChainSelectorToBytesAdapter{reportingPluginFactory}, m.BrokerExt))
	}, externalAdapterRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewReportingPluginFactoryReply{ID: id}, nil
}

func RegisterPluginSecureMintServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl core.PluginSecureMint) error {
	pb.RegisterServiceServer(server, &goplugin.ServiceServer{Srv: impl})
	pb.RegisterReportingPluginServiceServer(server, newPluginSecureMintServer(&net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl))
	return nil
}

func newPluginSecureMintServer(b *net.BrokerExt, gp core.PluginSecureMint) *pluginSecureMintServer {
	return &pluginSecureMintServer{BrokerExt: b.WithName("PluginSecureMintServer"), impl: gp}
}

var _ ocr3types.ReportingPluginFactory[[]byte] = (*reportingPluginFactoryChainSelectorToBytesAdapter)(nil)

// reportingPluginFactoryChainSelectorToBytesAdapter is a wrapper around the ReportingPluginFactory to implement ocr3types.ReportingPluginFactory[[]byte]
type reportingPluginFactoryChainSelectorToBytesAdapter struct {
	ocr3types.ReportingPluginFactory[core.ChainSelector]
}

func (r *reportingPluginFactoryChainSelectorToBytesAdapter) NewReportingPlugin(ctx context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	plugin, info, err := r.ReportingPluginFactory.NewReportingPlugin(ctx, config)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	return &reportingPluginChainSelectorToBytesAdapter{plugin: plugin}, info, nil
}
