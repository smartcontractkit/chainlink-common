package poller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/mailbox"
)

func Test_Poller(t *testing.T) {
	// Mock polling function that returns a new value every time it's called
	var pollNumber int
	pollFunc := func() (int, error) {
		pollNumber++
		return pollNumber, nil
	}

	poller := NewPoller[int](time.Millisecond, pollFunc, logger.Test(t))
	err := poller.Start()
	require.NoError(t, err)
	defer func() {
		err := poller.Stop()
		require.NoError(t, err)
	}()

	// Create subscriber channel
	subscriber := mailbox.NewHighCapacity[int]()
	poller.Subscribe(subscriber)
	defer poller.Unsubscribe(subscriber)

	// Create goroutine to receive updates from the subscriber
	pollCount := 0
	pollMax := 50
	go func() {
		for ; pollCount < pollMax; pollCount++ {
			<-subscriber.Notify()
			value, exists := subscriber.Retrieve()
			assert.True(t, exists)
			assert.Equal(t, pollNumber, value)
		}
	}()

	// Wait for a short duration to allow for some polling iterations
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, pollMax, pollCount)
}
