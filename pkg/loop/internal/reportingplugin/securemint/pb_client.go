package securemint

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	net "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	securemintpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	sm "github.com/smartcontractkit/chainlink-common/pkg/types/core/securemint"
)

// PluginSecureMintClient is a client that runs on the core node to connect to the SecureMint LOOP server.
type PluginSecureMintClient struct {
	// hashicorp plugin client
	*goplugin.PluginClient
	// client to base service
	*goplugin.ServiceClient

	reportingPluginService pb.ReportingPluginServiceClient
}

var _ core.PluginSecureMint = (*PluginSecureMintClient)(nil)

func NewPluginSecureMintClient(brokerCfg net.BrokerConfig) *PluginSecureMintClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginSecureMintClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &PluginSecureMintClient{
		PluginClient:           pc,
		ServiceClient:          goplugin.NewServiceClient(pc.BrokerExt, pc),
		reportingPluginService: pb.NewReportingPluginServiceClient(pc),
	}
}

// NewSecureMintFactory is called by the go-plugin client side to create a client-side ReportingPluginFactory.
func (c *PluginSecureMintClient) NewSecureMintFactory(
	ctx context.Context,
	lggr logger.Logger,
	externalAdapter sm.ExternalAdapter,
) (core.ReportingPluginFactory[sm.ChainSelector], error) {
	lggr.Infow("NewSecureMintFactory Client called", "externalAdapter", externalAdapter)

	cc := c.NewClientConn("SecureMintFactory", func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		lggr.Infow("Creating new client connection", "externalAdapter", externalAdapter)

		externalAdapterID, externalAdapterRes, err := c.ServeNew("ExternalAdapter", func(s *grpc.Server) {
			securemintpb.RegisterExternalAdapterServer(s, newExternalAdapterServer(lggr, externalAdapter))
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(externalAdapterRes)

		// this calls into plugin_securemint_server_pb.go#pluginSecureMintServer.NewReportingPluginFactory
		reply, err := c.reportingPluginService.NewReportingPluginFactory(ctx, &pb.NewReportingPluginFactoryRequest{
			ReportingPluginServiceConfig: &pb.ReportingPluginServiceConfig{
				ProviderType:  "",
				Command:       "",
				PluginName:    core.PluginSecureMintName,
				TelemetryType: "",
				PluginConfig:  "",
			},
			PipelineRunnerID: externalAdapterID, // TODO(gg): should probably be `ExternalAdapterID: externalAdapterID`, we're misusing the PipelineRunnerID for now
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ID, deps, nil
	})

	return &ocr3ReportingPluginFactoryBytesToChainSelectorAdapter{
		ocr3.NewReportingPluginFactoryClient(c.BrokerExt, cc), // protobuf client
	}, nil
}
