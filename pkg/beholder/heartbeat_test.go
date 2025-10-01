package beholder_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestHeartbeat_NewHeartbeat(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)

	heartbeat := beholder.NewHeartbeat(
		1*time.Second,
		lggr,
		beholder.WithAppID("test-app"),
		beholder.WithServiceName("test-service"),
		beholder.WithVersion("1.0.0"),
		beholder.WithCommit("abc123"),
	)
	require.NotNil(t, heartbeat)

	assert.Equal(t, "test-app", heartbeat.AppID)
	assert.Equal(t, "test-service", heartbeat.ServiceName)
	assert.Equal(t, "1.0.0", heartbeat.Version)
	assert.Equal(t, "abc123", heartbeat.Commit)
	assert.Equal(t, 1*time.Second, heartbeat.Beat)
	assert.NotNil(t, heartbeat.Emitter)
	assert.NotNil(t, heartbeat.Meter)
}

func TestHeartbeat_Start(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)

	heartbeat := beholder.NewHeartbeat(100*time.Millisecond, lggr)
	require.NotNil(t, heartbeat)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = heartbeat.Start(ctx)
	require.NoError(t, err)

	// Wait for at least one heartbeat
	time.Sleep(150 * time.Millisecond)

	err = heartbeat.Close()
	require.NoError(t, err)
}

func TestHeartbeat_Defaults(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)

	heartbeat := beholder.NewHeartbeat(1*time.Second, lggr)
	require.NotNil(t, heartbeat)

	// Check defaults
	assert.Equal(t, "chainlink", heartbeat.AppID)
	assert.Equal(t, "github.com/smartcontractkit/chainlink-common", heartbeat.ServiceName)
	assert.Equal(t, "(devel)", heartbeat.Version)
	assert.Equal(t, "unset", heartbeat.Commit)
	assert.Equal(t, 1*time.Second, heartbeat.Beat)
}
