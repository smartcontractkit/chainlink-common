package requests_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
)

type testStatsCollector struct {
	requestCount int
}

func (t *testStatsCollector) SetRequestCount(requestCount int) {
	t.requestCount = requestCount
}

func TestOCR3Store_Stats(t *testing.T) {
	// Create a new store with stats collector

	statsCollector := &testStatsCollector{}
	s := requests.NewStoreWithStatsCollector[*ocr3.ReportRequest](statsCollector)
	rid := uuid.New().String()
	req := &ocr3.ReportRequest{
		WorkflowExecutionID: rid,
	}

	err := s.Add(req)
	require.NoError(t, err)

	assert.Equal(t, 1, statsCollector.requestCount)

	// Add the same request again
	err = s.Add(req)
	require.Error(t, err)

	assert.Equal(t, 1, statsCollector.requestCount)

	// Evict the request
	_, ok := s.Evict(rid)
	assert.True(t, ok)

	assert.Equal(t, 0, statsCollector.requestCount)

	// Evict the request again
	_, ok = s.Evict(rid)
	assert.False(t, ok)

	assert.Equal(t, 0, statsCollector.requestCount)
}
