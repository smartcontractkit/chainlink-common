package beholder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// The engine closes sub-services after the parent Close hook; noopCloser suppresses that, so
// Client.close stays the emitter's sole closer.
func TestNoopCloser_SuppressesClose(t *testing.T) {
	svc, err := newChipIngressBatchEmitterService(&chipingress.NoopClient{}, Config{
		ChipIngressSendInterval: 10 * time.Millisecond,
		ChipIngressSendTimeout:  time.Second,
	}, logger.Test(t), "test", "", batch.WithNoAck())
	require.NoError(t, err)
	require.NoError(t, svc.Start(t.Context()))

	ev, err := chipingress.NewEvent("test-domain", "Entity", []byte("x"), nil)
	require.NoError(t, err)

	require.NoError(t, noopCloser{Service: svc}.Close())
	require.NoError(t, svc.EmitEvent(context.Background(), ev, nil), "still started after the suppressed close")

	require.NoError(t, svc.Close(), "the real closer works")
	require.Error(t, svc.EmitEvent(context.Background(), ev, nil), "stopped after the real close")
}
