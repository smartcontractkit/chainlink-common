package wasm

import (
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/backoff"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_toEmitLabels(t *testing.T) {
	t.Run("successfully transforms metadata", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:          "workflow-id",
			WorkflowName:        "workflow-name",
			WorkflowOwner:       "workflow-owner",
			WorkflowExecutionID: "6e2a46e3b6ae611bdb9bcc36ed3f46bb9a30babc3aabdd4eae7f35dd9af0f244",
		}
		empty := make(map[string]string, 0)

		gotLabels, err := toEmitLabels(md, empty)
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"workflow_id":            "workflow-id",
			"workflow_name":          "workflow-name",
			"workflow_owner_address": "workflow-owner",
			"workflow_execution_id":  "6e2a46e3b6ae611bdb9bcc36ed3f46bb9a30babc3aabdd4eae7f35dd9af0f244",
		}, gotLabels)
	})

	t.Run("fails on missing workflow id", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowName:  "workflow-name",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow id")
	})

	t.Run("fails on missing workflow name", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:    "workflow-id",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow name")
	})

	t.Run("fails on missing workflow owner", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:   "workflow-id",
			WorkflowName: "workflow-name",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow owner")
	})

	t.Run("fails on missing workflow execution id", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:    "workflow-id",
			WorkflowName:  "workflow-name",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow execution id")
	})
}

func Test_bufferToPointerLen(t *testing.T) {
	t.Run("fails when no buffer", func(t *testing.T) {
		_, _, err := bufferToPointerLen([]byte{})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "buffer cannot be empty")
	})
}

// testFetcher tracks the number of times its do method has been called, returning
// a success response after returnAfter calls.
type testFetcher struct {
	minCalls int
	calls    int
}

func (f *testFetcher) do(req sdk.FetchRequest) (sdk.FetchResponse, error) {
	f.calls++
	if f.calls > f.minCalls {
		return sdk.FetchResponse{
			StatusCode: 200,
			Body:       []byte("success"),
		}, nil
	}
	return sdk.FetchResponse{}, assert.AnError
}

func Test_Fetch(t *testing.T) {
	t.Run("Fetch with retries when specified", func(t *testing.T) {
		var (
			wantRetries     = 3
			wantReturnAfter = wantRetries - 1
			wantResp        = sdk.FetchResponse{
				StatusCode: 200,
				Body:       []byte("success"),
			}
			tf = &testFetcher{
				minCalls: wantReturnAfter,
			}

			// define a Runtime with the mocked fetch implementation
			r = &Runtime{
				fetchFn: tf.do,
			}
		)

		req := sdk.FetchRequest{
			RetryOptions: []backoff.RetryOption{
				backoff.WithBackOff(&backoff.ZeroBackOff{}),
				backoff.WithMaxTries(uint(wantRetries)),
			},
		}
		resp, err := r.Fetch(req)
		require.NoError(t, err)
		require.Equal(t, wantResp, resp)
		require.Equal(t, wantRetries, tf.calls)
	})

	// Tests that if a TimeoutMs is passed on the request that the retry logic
	// will respect the timeout.  A timeout is given and the fetch requires
	// multiple calls to complete, so it should always error due to the timeout.
	t.Run("Fetch with retries respects timeout", func(t *testing.T) {
		var (
			wantRetries     = 3
			wantReturnAfter = wantRetries - 1

			tf = &testFetcher{
				minCalls: wantReturnAfter,
			}
			timeout = 30

			// define a constant backoff that waits the timeout duration before
			// trying again.
			giveBackoff = backoff.NewConstantBackOff(time.Duration(timeout) * time.Millisecond)

			// define a Runtime with the mocked fetch implementation
			r = &Runtime{
				fetchFn: tf.do,
			}
		)

		req := sdk.FetchRequest{
			TimeoutMs: uint32(timeout),
			RetryOptions: []backoff.RetryOption{
				backoff.WithBackOff(giveBackoff),
				backoff.WithMaxTries(uint(wantRetries)),
			},
		}
		_, err := r.Fetch(req)
		require.Error(t, err)
		require.True(t, tf.calls < wantRetries)
	})
}
