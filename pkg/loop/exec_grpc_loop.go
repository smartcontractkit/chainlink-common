package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	//"github.com/smartcontractkit/chainlink-common/pkg/loop"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ExecName is the name for [types.PluginMedian]/[NewGRPCPluginMedian].
const ExecName = "exec"

func PluginCCIPExecHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_CCIP_EXEC_MAGIC_COOKIE",
		MagicCookieValue: "b12a697e19748cd695dd1690c09745ee7cc03717179958e8eadd5a7ca4699999",
	}
}

type ExecLoop struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer types.CCIPExecFactoryGenerator

	pluginClient *internal.PluginMedianClient
}

func (p *ExecLoop) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return internal.RegisterExecutionLOOPServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginMedian], updated with the new broker and conn.
func (p *ExecLoop) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = internal.NewPluginMedianClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return types.PluginMedian(p.pluginClient), nil
}

func (p *ExecLoop) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  PluginMedianHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{ExecName: p},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  p.DialOpts,
		Logger:           HCLogLogger(p.Logger),
	}
}

var _ ocrtypes.ReportingPluginFactory = (*ExecFactoryService)(nil)

// ExecFactoryService is a [types.Service] that maintains an internal [types.PluginMedian].
type ExecFactoryService struct {
	internal.PluginService[*ExecLoop, types.ReportingPluginFactory]
}

// NewExecService returns a new [*ExecFactoryService].
// cmd must return a new exec.Cmd each time it is called.
func NewExecService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.CCIPExecProvider, config types.CCIPExecFactoryGeneratorConfig) *ExecFactoryService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.CCIPExecFactoryGenerator)
		if !ok {
			return nil, fmt.Errorf("expected PluginMedian but got %T", instance)
		}
		return plug.NewExecFactory(ctx, provider, config)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MedianService")
	var efs ExecFactoryService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	efs.Init(ExecName, &ExecLoop{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &efs
}

func (m *ExecFactoryService) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := m.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(config)
}
