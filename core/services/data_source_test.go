package services

import (
	"context"
	"fmt"
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
	// t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		data <- big.NewInt(1000) // mock trigger returned response by sending to channel
		data <- big.NewInt(1001) // return data for juelsPerX data
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
	// t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		// do not return data, context cancel should handle ending function call
	})
	defer server.Close()

	// trigger observation using the chainlink module
	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// create context timeout
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()

	// run observe function
	_, err := ds.Price.Observe(ctx)
	assert.Equal(t, fmt.Sprintf("[%s - PriceFeed] [%s] Observe (job run) cancelled", jobID, jobID), err.Error()) // validate Price error
}

func TestDataSource_Observe_MultipleCalls(t *testing.T) {
	// t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	called := 0
	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		called++
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
		data <- big.NewInt(1000) // mock trigger returned response by sending to channel
		data <- big.NewInt(1001) // return data for juelsPerX data
	})
	defer server.Close()

	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// trigger simultaneous observation using the chainlink module
	ansChan := make(chan uint64)
	go func() {
		ans, err := ds.JuelsToX.Observe(context.TODO())
		assert.NoError(t, err)
		ansChan <- ans.Uint64()
	}()
	res, err := ds.Price.Observe(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), res.Uint64()) // validate Price observation
	assert.Equal(t, uint64(1001), <-ansChan)    // validate JuelsToX observation
	assert.Equal(t, 1, called)                  // server should only be called once even though two observes called
}

func TestDataSource_Observe_MultipleCalls_WithContextCancel(t *testing.T) {
	// t.Parallel()

	jobID := "test-job-id"
	data := make(chan *big.Int)

	called := 0
	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		called++
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
	})
	defer server.Close()

	trigger := webhook.NewTrigger(server.URL, &test.MockWebhookConfig{})
	ds := NewDataSources(jobID, &trigger, &data, logger.Default)

	// create context timeout
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()

	// trigger simultaneous observation using the chainlink module
	errChan := make(chan string)
	go func() {
		_, err := ds.JuelsToX.Observe(ctx)
		fmt.Println(err)
		errChan <- err.Error()
	}()
	_, err := ds.Price.Observe(ctx)
	assert.Equal(t, fmt.Sprintf("[%s - PriceFeed] [%s] Observe (job run) cancelled", jobID, jobID), err.Error()) // validate Price error
	assert.Equal(t, fmt.Sprintf("[%s - JuelsToX] Observe context cancelled", jobID), <-errChan)                  // validate JuelsToX error
	assert.Equal(t, 1, called)                                                                                   // server should only be called once even though two observes called
}
