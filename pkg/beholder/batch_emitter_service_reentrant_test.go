package beholder

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// hangingNoAckClient never answers PublishBatchNoAck until released or the call's context ends.
type hangingNoAckClient struct {
	chipingress.NoopClient
	release chan struct{}
}

func (c *hangingNoAckClient) PublishBatchNoAck(ctx context.Context, _ *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.release:
		return &chipingress.PublishResponse{}, nil
	}
}

// A callback that re-enters EmitEvent (e.g. to retry a buffer-full drop) while Close is waiting
// to stop the service must not deadlock: the callback must not run under the service's lock.
func TestChipIngressBatchEmitterService_ReentrantCallbackDuringClose(t *testing.T) {
	client := &hangingNoAckClient{release: make(chan struct{})}
	t.Cleanup(func() { close(client.release) })

	svc, err := newChipIngressBatchEmitterService(client, Config{
		ChipIngressBufferSize:         1,
		ChipIngressMaxBatchSize:       1,
		ChipIngressMaxConcurrentSends: 1,
		ChipIngressSendInterval:       10 * time.Millisecond,
		ChipIngressSendTimeout:        30 * time.Second,
		ChipIngressDrainTimeout:       200 * time.Millisecond,
	}, logger.Test(t), "test", "", batch.WithNoAck())
	require.NoError(t, err)
	require.NoError(t, svc.Start(t.Context()))

	event := func(i int) chipingress.CloudEvent {
		ev, err := chipingress.NewEvent("test-domain", "Entity", []byte(fmt.Sprint(i)), nil)
		require.NoError(t, err)
		return ev
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var retried atomic.Bool
		for i := 0; i < 100 && !retried.Load(); i++ {
			_ = svc.EmitEvent(context.Background(), event(i), func(err error) {
				if !errors.Is(err, batch.ErrMessageBufferFull) || !retried.CompareAndSwap(false, true) {
					return
				}
				go func() { _ = svc.Close() }()
				time.Sleep(100 * time.Millisecond) // let Close start waiting to stop the service
				_ = svc.EmitEvent(context.Background(), event(-1), nil)
			})
			time.Sleep(5 * time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock: a callback re-entering EmitEvent during Close never returned")
	}
}
