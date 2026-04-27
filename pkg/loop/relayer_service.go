package loop

import (
	"context"
	"fmt"
	"math/big"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ Relayer = (*RelayerService)(nil)

// RelayerService is a [types.Service] that maintains an internal [Relayer].
type RelayerService struct {
	goplugin.PluginService[*GRPCPluginRelayer, Relayer]
}

// NewRelayerService returns a new [*RelayerService].
// cmd must return a new exec.Cmd each time it is called.
func NewRelayerService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, config string, keystore core.Keystore, csaKeystore core.Keystore, capabilityRegistry core.CapabilitiesRegistry) *RelayerService {
	newService := func(ctx context.Context, instance any) (Relayer, services.HealthReporter, error) {
		plug, ok := instance.(PluginRelayer)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginRelayer but got %T", instance)
		}
		r, err := plug.NewRelayer(ctx, config, keystore, csaKeystore, capabilityRegistry)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create Relayer: %w", err)
		}
		return r, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "RelayerService")
	var rs RelayerService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	rs.Init(PluginRelayerName, &GRPCPluginRelayer{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &rs
}

func (r *RelayerService) EVM() (types.EVMService, error) {
	if err := r.Wait(); err != nil {
		return nil, err
	}
	return r.CurrentService().EVM()
}

func (r *RelayerService) TON() (types.TONService, error) {
	if err := r.Wait(); err != nil {
		return nil, err
	}
	return r.CurrentService().TON()
}

func (r *RelayerService) Solana() (types.SolanaService, error) {
	if err := r.Wait(); err != nil {
		return nil, err
	}
	return r.CurrentService().Solana()
}

func (r *RelayerService) Aptos() (types.AptosService, error) {
	if err := r.Wait(); err != nil {
		return nil, err
	}
	return r.CurrentService().Aptos()
}

func (r *RelayerService) NewContractReader(ctx context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewContractReader(ctx, contractReaderConfig)
}

func (r *RelayerService) NewContractWriter(ctx context.Context, contractWriterConfig []byte) (types.ContractWriter, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewContractWriter(ctx, contractWriterConfig)
}

func (r *RelayerService) NewConfigProvider(ctx context.Context, args types.RelayArgs) (types.ConfigProvider, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewConfigProvider(ctx, args)
}

func (r *RelayerService) NewPluginProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.PluginProvider, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewPluginProvider(ctx, rargs, pargs)
}

func (r *RelayerService) NewLLOProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.LLOProvider, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewLLOProvider(ctx, rargs, pargs)
}

func (r *RelayerService) NewCCIPProvider(ctx context.Context, cargs types.CCIPProviderArgs) (types.CCIPProvider, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return r.CurrentService().NewCCIPProvider(ctx, cargs)
}

func (r *RelayerService) LatestHead(ctx context.Context) (types.Head, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return types.Head{}, err
	}
	return r.CurrentService().LatestHead(ctx)
}

func (r *RelayerService) FinalizedHead(ctx context.Context) (types.Head, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return types.Head{}, err
	}
	return r.CurrentService().FinalizedHead(ctx)
}

func (r *RelayerService) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return types.ChainStatus{}, err
	}
	return r.CurrentService().GetChainStatus(ctx)
}

func (r *RelayerService) GetChainInfo(ctx context.Context) (types.ChainInfo, error) {
	if err := r.WaitCtx(ctx); err != nil {
		return types.ChainInfo{}, err
	}
	return r.CurrentService().GetChainInfo(ctx)
}

func (r *RelayerService) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (nodes []types.NodeStatus, nextPageToken string, total int, err error) {
	if err := r.WaitCtx(ctx); err != nil {
		return nil, "", -1, err
	}
	return r.CurrentService().ListNodeStatuses(ctx, pageSize, pageToken)
}

func (r *RelayerService) Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error {
	if err := r.WaitCtx(ctx); err != nil {
		return err
	}
	return r.CurrentService().Transact(ctx, from, to, amount, balanceCheck)
}

func (r *RelayerService) Replay(ctx context.Context, fromBlock string, args map[string]any) error {
	if err := r.WaitCtx(ctx); err != nil {
		return err
	}
	return r.CurrentService().Replay(ctx, fromBlock, args)
}
