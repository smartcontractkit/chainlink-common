package net

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestWrappedErrorError(t *testing.T) {
	t.Parallel()
	t.Run("Is returns false for different error code", func(t *testing.T) {
		// silly, but to verify that it's only looking at error code here, we need to make the message the same
		err := types.NotFoundError(types.ErrInvalidType.Error())
		assert.NotErrorIs(t, WrapRPCErr(types.ErrInvalidType), err)
	})

	t.Run("Is returns false for different message", func(t *testing.T) {
		// Both are InvalidArgumentError
		assert.NotErrorIs(t, WrapRPCErr(types.ErrInvalidType), types.ErrInvalidEncoding)
	})

	t.Run("Is returns true if the message and code are the same", func(t *testing.T) {
		assert.ErrorIs(t, WrapRPCErr(types.ErrInvalidType), types.ErrInvalidType)
	})

	t.Run("Is returns true if the message is contained and the code is the same", func(t *testing.T) {
		wrapped := WrapRPCErr(fmt.Errorf("%w: %w", types.ErrInvalidType, errors.New("some other error")))
		assert.ErrorIs(t, wrapped, types.ErrInvalidType)
	})
}

func TestClientConn_refresh_doesNotBlockPastDeadline(t *testing.T) {
	t.Parallel()

	dialing := make(chan struct{}) // signals the stuck refresh has reached newClient
	var dialingOnce sync.Once

	be := &BrokerExt{BrokerConfig: BrokerConfig{Logger: logger.Test(t)}}
	c := be.NewClientConn("test", func(ctx context.Context) (uint32, Resources, error) {
		dialingOnce.Do(func() { close(dialing) })
		// Simulate an unavailable relayer: block until the caller's context ends.
		<-ctx.Done()
		return 0, nil, errors.New("relayer unavailable")
	})

	// First caller: gets stuck inside newClient holding the refresh lock until we cancel it.
	stuckCtx, cancelStuck := context.WithCancel(context.Background())
	stuckDone := make(chan struct{})
	go func() {
		defer close(stuckDone)
		_, _ = c.refresh(stuckCtx, nil)
	}()
	t.Cleanup(func() {
		cancelStuck()
		<-stuckDone
	})

	select {
	case <-dialing:
	case <-time.After(time.Second):
		t.Fatal("first refresh never reached newClient")
	}

	// Second caller: has a deadline; must return promptly rather than waiting out the first refresh.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := c.refresh(ctx, nil)
		done <- err
	}()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(2 * time.Second):
		t.Fatal("second refresh blocked past its deadline waiting for the stuck refresh")
	}
}
