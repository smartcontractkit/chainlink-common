package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/types"
)

//var _ internal.PluginKeystore = (*KeystoreService)(nil)

// KeystoreService is a [types.Service] that maintains an internal [keystore.Keystore].
type KeystoreService struct {
	goplugin.PluginService[*GRPCPluginKeystore, internal.Keystore]
}

func NewKeystoreService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, config []byte) *KeystoreService {
	newService := func(ctx context.Context, instance any) (internal.Keystore, error) {
		plug, ok := instance.(internal.PluginKeystoreFactory)
		if !ok {
			return nil, fmt.Errorf("expected PluginKeystore but got %T", instance)
		}
		r, err := plug.NewKeystore(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Keystore: %w", err)
		}
		return r, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "KeystoreService")
	var rs KeystoreService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	rs.Init(PluginKeystoreName, &GRPCPluginKeystore{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &rs
}
