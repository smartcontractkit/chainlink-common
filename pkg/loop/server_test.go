package loop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestServerStartBeholderClientUsesBackgroundContext(t *testing.T) {
	prevNewBeholderClient := newBeholderClient
	prevStartBeholderClient := startBeholderClient
	prevGlobalBeholderClient := beholder.GetClient()
	t.Cleanup(func() {
		newBeholderClient = prevNewBeholderClient
		startBeholderClient = prevStartBeholderClient
		beholder.SetClient(prevGlobalBeholderClient)
	})

	newBeholderClient = func(cfg beholder.Config) (*beholder.Client, error) {
		return beholder.NewNoopClient(), nil
	}

	var startCtx context.Context
	startBeholderClient = func(_ *beholder.Client, ctx context.Context) error {
		startCtx = ctx
		return nil
	}

	srv := &Server{Logger: logger.Sugared(logger.Test(t))}

	require.NoError(t, srv.startBeholderClient(beholder.Config{
		ChipIngressBatchEmitterEnabled: true,
		ChipIngressLogger:              logger.Sugared(logger.Test(t)),
	}))
	t.Cleanup(func() {
		if srv.beholderClient != nil {
			_ = srv.beholderClient.Close()
		}
	})

	require.NotNil(t, startCtx)
	require.NoError(t, startCtx.Err())
	require.Nil(t, startCtx.Done())
}
