package requests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func TestOCR3Store(t *testing.T) {
	n := time.Now()

	s := NewStore()
	rid := uuid.New().String()
	req := &Request{
		WorkflowExecutionID: rid,
		ExpiresAt:           n.Add(10 * time.Second),
	}

	t.Run("Add", func(t *testing.T) {
		err := s.Add(req)
		require.NoError(t, err)
	})

	t.Run("add duplicate", func(t *testing.T) {
		err := s.Add(req)
		require.Error(t, err)
	})

	t.Run("evict", func(t *testing.T) {
		_, wasPresent := s.evict(rid)
		assert.True(t, wasPresent)
		assert.Len(t, s.requests, 0)
	})

	t.Run("firstN", func(t *testing.T) {
		r, err := s.FirstN(1)
		assert.NoError(t, err)
		assert.Len(t, r, 0)
	})

	t.Run("firstN, zero batch size", func(t *testing.T) {
		_, err := s.FirstN(0)
		assert.ErrorContains(t, err, "batchsize cannot be 0")
	})

	t.Run("firstN, batchSize larger than queue", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := s.Add(&Request{WorkflowExecutionID: uuid.New().String(), ExpiresAt: n.Add(1 * time.Hour)})
			require.NoError(t, err)
		}
		items, err := s.FirstN(100)
		require.NoError(t, err)
		assert.Len(t, items, 10)
	})

	t.Run("getByIDs", func(t *testing.T) {
		rid2 := uuid.New().String()
		err := s.Add(req)
		require.NoError(t, err)
		reqs := s.GetByIDs([]string{rid, rid2})
		require.Equal(t, 1, len(reqs))
	})
}

func TestOCR3Store_ManagesStateConsistently(t *testing.T) {
	s := NewStore()
	rid := uuid.New().String()
	req := &Request{
		WorkflowExecutionID: rid,
	}

	err := s.Add(req)
	require.NoError(t, err)
	assert.Len(t, s.requests, 1)
	assert.Len(t, s.requestIDs, 1)

	s.GetByIDs([]string{rid})
	assert.Len(t, s.requests, 1)
	assert.Len(t, s.requestIDs, 1)

	_, ok := s.evict(rid)
	assert.True(t, ok)
	assert.Len(t, s.requests, 0)
	assert.Len(t, s.requestIDs, 0)

	err = s.Add(req)
	require.NoError(t, err)
	assert.Len(t, s.requests, 1)
	assert.Len(t, s.requestIDs, 1)
}

func TestOCR3Store_ReadRequestsCopy(t *testing.T) {
	s := NewStore()
	rid := uuid.New().String()
	cb := make(chan capabilities.CapabilityResponse, 1)
	stopCh := make(chan struct{}, 1)
	obs, err := values.NewList(
		[]any{"hello", 1},
	)
	require.NoError(t, err)
	req := &Request{
		WorkflowExecutionID: rid,
		CallbackCh:          cb,
		StopCh:              stopCh,
		Observations:        obs,
	}

	require.NoError(t, s.Add(req))

	testCases := []struct {
		name string
		get  func(ctx context.Context, rid string) *Request
	}{
		{
			name: "get",
			get: func(ctx context.Context, rid string) *Request {
				return s.Get(rid)
			},
		},
		{
			name: "firstN",
			get: func(ctx context.Context, rid string) *Request {
				rs, err := s.FirstN(1)
				require.NoError(t, err)
				assert.Len(t, rs, 1)
				return rs[0]
			},
		},
		{
			name: "getByIDs",
			get: func(ctx context.Context, rid string) *Request {
				rs := s.GetByIDs([]string{rid})
				assert.Len(t, rs, 1)
				return rs[0]
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			st.Parallel()

			gr := tc.get(tests.Context(st), rid)
			require.NoError(t, err)

			// Mutating the received observations doesn't mutate the store
			assert.Len(t, gr.Observations.Underlying, 2)
			gr.Observations.Underlying = append(gr.Observations.Underlying, values.NewString("world"))
			gr.WorkflowExecutionID = "incorrect mutation"
			assert.Len(t, gr.Observations.Underlying, 3)

			gr2 := tc.get(tests.Context(st), rid)
			assert.Len(t, gr2.Observations.Underlying, 2)
			assert.Equal(t, gr2.WorkflowExecutionID, rid)

			gr.StopCh <- struct{}{}
			<-stopCh

			gr.CallbackCh <- capabilities.CapabilityResponse{}
			<-cb
		})
	}

}
