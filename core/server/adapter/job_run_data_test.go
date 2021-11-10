package adapter

import (
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockJobResult struct {
	JobID  string
	Result string
}

// https://stackoverflow.com/questions/59186562/unit-testing-with-gin-gonic

func TestNewJobHandler_Run(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name        string
		pipelineErr error
		jobData     interface{}
		code        int
	}{
		{"success", nil, MockJobResult{"test", "1000"}, 201},
		{"bad request", nil, []byte{}, 400},
		{"server error", errors.New("failed run"), MockJobResult{"test", "1000"}, 500},
	}

	for _, i := range inputs {
		t.Run(i.name, func(t *testing.T) {
			// new handler
			job := NewJobHandler(test.MockPipeline{Error: i.pipelineErr})

			// create response recorder and gin context with correct payload
			res, ctx, err := test.MockGinContext(i.jobData)
			require.NoError(t, err)

			job.Run(ctx)
			assert.Equal(t, i.code, res.Code)
		})
	}

}
