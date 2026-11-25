package net

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

// A refresh holds the connection's lock until it has a new connection or its context
// ends. The state must be readable all that time: the capabilities registry reads it to
// decide whether the capability behind the connection can be replaced.
func TestClientConn_GetStateDoesNotWaitForRefresh(t *testing.T) {
	t.Parallel()
	refreshing := make(chan struct{})
	var once sync.Once
	b := &BrokerExt{BrokerConfig: BrokerConfig{Logger: logger.Test(t), StopCh: make(chan struct{})}}
	// The plugin is gone: every attempt to connect waits, then fails.
	c := b.NewClientConn("test", func(ctx context.Context) (uint32, Resources, error) {
		once.Do(func() { close(refreshing) })
		<-ctx.Done()
		return 0, nil, errors.New("plugin is gone")
	})
	assert.Equal(t, connectivity.Shutdown, c.GetState())

	// A caller without a deadline: its refresh lasts until the test ends it.
	ctx, cancel := context.WithCancel(t.Context())
	invoked := make(chan error, 1)
	go func() { invoked <- c.Invoke(ctx, "/service/Method", nil, nil) }()
	<-refreshing

	state := make(chan connectivity.State, 1)
	go func() { state <- c.GetState() }()
	select {
	case got := <-state:
		assert.Equal(t, connectivity.TransientFailure, got)
	case <-time.After(tests.WaitTimeout(t)):
		require.FailNow(t, "GetState waited for the refresh")
	}

	cancel()
	require.Error(t, <-invoked)
	assert.Equal(t, connectivity.Shutdown, c.GetState())
}
