package beholder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// The engine's start rollback closes started subs through their registered handle, and its
// teardown closes them again after the parent Close hook: the registered Close must really
// close a started emitter, and treat the re-request on an already-stopped one as success.
func TestIdempotentCloser_Close(t *testing.T) {
	newSvc := func(start bool) *ChipIngressBatchEmitterService {
		svc, err := newChipIngressBatchEmitterService(&chipingress.NoopClient{}, Config{
			ChipIngressSendInterval: 10 * time.Millisecond,
			ChipIngressSendTimeout:  time.Second,
		}, logger.Test(t), "test", "", batch.WithNoAck())
		require.NoError(t, err)
		if start {
			require.NoError(t, svc.Start(t.Context()))
		}
		return svc
	}
	ev, err := chipingress.NewEvent("test-domain", "Entity", []byte("x"), nil)
	require.NoError(t, err)

	// Started: the engine's start rollback closes it for real.
	svc := newSvc(true)
	require.NoError(t, idempotentCloser{Service: svc}.Close())
	require.Error(t, svc.EmitEvent(context.Background(), ev, nil), "closed by the engine's rollback")

	// Already stopped: the engine's teardown re-request after Client.close is a success.
	svc = newSvc(true)
	require.NoError(t, svc.Close())
	require.NoError(t, idempotentCloser{Service: svc}.Close())

	// Never started: errors other than ErrAlreadyStopped pass through.
	require.ErrorIs(t, idempotentCloser{Service: newSvc(false)}.Close(), services.ErrCannotStopUnstarted)
}
