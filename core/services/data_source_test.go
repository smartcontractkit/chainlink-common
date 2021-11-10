package services

import (
	"context"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataSource_Observe(t *testing.T) {
	t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		data <- big.NewInt(1000) // mock trigger returned response by sending to channel
		data <- big.NewInt(1000) // return data for juelsPerX data
	})
	defer server.Close()

	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// trigger observation using the chainlink module
	res, err := ds.Price.Observe(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), res.Uint64())
}

func TestDataSource_Observe_WithContextCancel(t *testing.T) {
	t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		// do not return data, context cancel should handle ending function call
	})
	defer server.Close()

	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// trigger observation using the chainlink module
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()

	// run observe function
	go ds.Price.Observe(ctx)

	// context should timeout before data is returned in channel
	select {
	case <-data:
		require.True(t, false, "Unexpected answer was received")
	case <-ctx.Done():
		require.True(t, true, "Context timeout exceeded")
	}
}

func TestDataSource_Observe_MultipleCalls(t *testing.T) {
	t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	called := 0
	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		called++
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		data <- big.NewInt(1000) // mock trigger returned response by sending to channel
		data <- big.NewInt(1000) // return data for juelsPerX data
	})
	defer server.Close()

	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// trigger simultaneous observation using the chainlink module
	go ds.JuelsToX.Observe(context.TODO())
	res, err := ds.Price.Observe(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), res.Uint64())
	assert.Equal(t, 1, called) // server should only be called once even though two observes called
}
