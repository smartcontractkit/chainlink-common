package beholder

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"

	clock "github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logger "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestHeartbeat(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Initialize a test logger
	logger := logger.Test(t)
	// Create a string builder to capture output
	var output strings.Builder
	// Create a new writer client
	beholderClient, err := NewWriterClient(&output)
	require.NoError(t, err)
	// Create a mock clock for testing
	mockClock := clock.NewFakeClockAt(time.Now())
	// Set the heartbeat interval
	heartbeatInterval := 1 * time.Second
	// Create a new heartbeat instance
	heartbeat, err := NewHeartbeat(beholderClient, heartbeatInterval, logger, mockClock)
	require.NoError(t, err)
	assert.Len(t, heartbeat.attributes, 4)
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	// Start the heartbeat
	err = heartbeat.Start(ctx)
	require.NoError(t, err) // Ensure no error occurred during heartbeat start
	// Attempt to start the heartbeat again
	err = heartbeat.Start(ctx)
	require.Error(t, err, "heartbeat already started")
	// Advance the clock to trigger the heartbeat
	mockClock.Advance(2 * time.Second)

	// Force the tick to trigger the heartbeat
	// heartbeat.tickCh <- struct{}{}

	// Send the heartbeat now
	heartbeat.Send()

	// Stop the heartbeat, call's Close
	cancel()
	// Check if the channel is closed
	_, ok := <-heartbeat.done
	// Close to flush metrics
	beholderClient.Close()

	// TODO: fix flaky assertion
	// Verify the heartbeat counter is in the output
	// assert.Contains(t, output.String(), `"beholder_heartbeat_counter"`)

	assert.False(t, ok, "channel is closed")
	// Be able to call Close multiple times
	err = heartbeat.Close()
	require.Error(t, err, "heartbeat already closed")
}
