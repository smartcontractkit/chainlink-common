package ccip

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"google.golang.org/grpc"
)

var _ core.PluginCCIPCommit = (*PluginCCIPCommitClient)(nil)
var _ pb.PluginCCIPCommitServer = (*pluginCCIPCommitServer)(nil)

type PluginCCIPCommitClient struct {
	*goplugin.PluginClient
	*goplugin.ServiceClient

	ccipCommit pb.PluginCCIPCommitClient
}

// NewCCIPCommitFactory implements core.PluginCCIPCommit.
func (p *PluginCCIPCommitClient) NewCCIPCommitFactory(ctx context.Context, contractReaders map[types.RelayID]types.ContractReader) (types.ReportingPluginFactory, error) {
	panic("unimplemented")
}

func NewPluginCCIPCommitClient(broker net.Broker, brokerCfg net.BrokerConfig, conn *grpc.ClientConn) *PluginCCIPCommitClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginCCIPCommitClient")
	pc := goplugin.NewPluginClient(broker, brokerCfg, conn)
	return &PluginCCIPCommitClient{PluginClient: pc, ccipCommit: pb.NewPluginCCIPCommitClient(pc), ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc)}
}

func RegisterPluginCCIPCommitServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl core.PluginCCIPCommit) error {
	pb.RegisterPluginCCIPCommitServer(server, newPluginCCIPCommitServer(&net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl))
	return nil
}

func newPluginCCIPCommitServer(b *net.BrokerExt, mp core.PluginCCIPCommit) *pluginCCIPCommitServer {
	return &pluginCCIPCommitServer{BrokerExt: b.WithName("PluginCCIPCommit"), impl: mp}
}

type pluginCCIPCommitServer struct {
	pb.UnimplementedPluginCCIPCommitServer

	*net.BrokerExt
	impl core.PluginCCIPCommit
}
