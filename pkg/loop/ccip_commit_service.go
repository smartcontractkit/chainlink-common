package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const PluginCommitServiceName = "CCIPCommitService"

var _ ocrtypes.ReportingPluginFactory = (*CCIPCommitService)(nil)

// CCIPCommitService is a [types.Service] that maintains an internal [types.PluginCCIPCommit].
type CCIPCommitService struct {
	goplugin.PluginService[*GRPCPluginCCIPCommit, types.ReportingPluginFactory]
}

func NewCCIPCommitService(
	lggr logger.Logger,
	grpcOpts GRPCOpts,
	cmd func() *exec.Cmd,
	contractReaders map[types.RelayID]types.ContractReader,
) *CCIPCommitService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(core.PluginCCIPCommit)
		if !ok {
			return nil, fmt.Errorf("expected PluginCCIPCommit but got %T", instance)
		}
		return plug.NewCCIPCommitFactory(ctx, contractReaders)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "CCIPCommitService")
	var ms CCIPCommitService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginCommitServiceName, &GRPCPluginCCIPCommit{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

// NewReportingPlugin implements types.ReportingPluginFactory.
func (c *CCIPCommitService) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := c.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return c.Service.NewReportingPlugin(config)
}
