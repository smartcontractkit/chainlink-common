package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// KeystoreService is a [types.Service] that maintains an internal [keystore.Keystore].
type KeystoreService struct {
	goplugin.PluginService[*GRPCPluginKeystore, keystore.GRPCService]
}

func NewKeystoreService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, config []byte) *KeystoreService {
	newService := func(ctx context.Context, instance any) (keystore.GRPCService, services.HealthReporter, error) {
		plug, ok := instance.(*keystore.Client)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginKeystore but got %T", instance)
		}
		return plug, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "KeystoreService")
	var rs KeystoreService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	rs.Init(PluginKeystoreName, &GRPCPluginKeystore{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &rs
}
