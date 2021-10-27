package webhook

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrigger_TriggerJob(t *testing.T) {
	t.Parallel()

	jobID := "test-job-id"
	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, fmt.Sprintf("/v2/jobs/%s/runs", jobID), req.URL.Path)
		assert.Equal(t, "AccessKey", req.Header.Get("X-Chainlink-EA-AccessKey"))
		assert.Equal(t, "Secret", req.Header.Get("X-Chainlink-EA-Secret"))
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
	})
	defer server.Close()

	// create client and trigger job
	webhook := NewTrigger(server.URL, "AccessKey", "Secret")
	webhook.TriggerJob(jobID)
}

func TestNewTrigger_TriggerJob_Retryable(t *testing.T) {
	t.Parallel()

	jobID := "test-job-id"

	// simulate a server error to test the retryable http
	reqCount := 0
	retries := 2
	server := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		if reqCount < retries {
			require.NoError(t, test.WriteResponse(rw, http.StatusInternalServerError, nil))
			reqCount++
			return
		}
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
	})
	defer server.Close()

	// create client and trigger job
	webhook := NewTrigger(server.URL, "", "")
	webhook.TriggerJob(jobID)
	assert.Equal(t, retries, reqCount, "Retry amount did not match expected number of requests")
}
