package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/types"
)

var _ internal.PluginKeystore = (*KeystoreService)(nil)

// KeystoreService is a [types.Service] that maintains an internal [keystore.Keystore].
type KeystoreService struct {
	goplugin.PluginService[*GRPCPluginKeystore, internal.PluginKeystore]
}

func (k *KeystoreService) Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error) {
	if err := k.WaitCtx(ctx); err != nil {
		return nil, err
	}
	return k.Service.Sign(ctx, keyID, data)
}

func (k *KeystoreService) SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeystoreService) Verify(ctx context.Context, keyID []byte, data []byte) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeystoreService) VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeystoreService) Get(ctx context.Context, tags []string) ([][]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KeystoreService) RunUDF(ctx context.Context, udfName string, keyID []byte, data []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func NewKeystoreService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, config []byte) *KeystoreService {
	newService := func(ctx context.Context, instance any) (internal.PluginKeystore, error) {
		plug, ok := instance.(internal.PluginKeystore)
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
