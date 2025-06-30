package requests_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
)

func TestOCR3Store(t *testing.T) {
	n := time.Now()

	s := requests.NewStore[*ocr3.ReportRequest]()
	rid := uuid.New().String()
	req := &ocr3.ReportRequest{
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
		_, wasPresent := s.Evict(rid)
		assert.True(t, wasPresent)
		reqs, err := s.FirstN(10)
		require.NoError(t, err)
		assert.Len(t, reqs, 0)
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
			err := s.Add(&ocr3.ReportRequest{WorkflowExecutionID: uuid.New().String(), ExpiresAt: n.Add(1 * time.Hour)})
			require.NoError(t, err)
		}
		items, err := s.FirstN(100)
		require.NoError(t, err)
		assert.Len(t, items, 10)
	})

	t.Run("rangeN", func(t *testing.T) {
		err := s.Add(&ocr3.ReportRequest{WorkflowExecutionID: uuid.New().String(), ExpiresAt: n.Add(1 * time.Hour)})
		r, err := s.RangeN(0, 1)
		assert.NoError(t, err)
		assert.Len(t, r, 1)
	})

	t.Run("rangeN, zero batch size", func(t *testing.T) {
		_, err := s.RangeN(0, 0)
		assert.ErrorContains(t, err, "batchSize must greater than 0")
	})

	t.Run("rangeN, batchSize larger than queue with start offset", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := s.Add(&ocr3.ReportRequest{WorkflowExecutionID: uuid.New().String(), ExpiresAt: n.Add(1 * time.Hour)})
			require.NoError(t, err)
		}
		items, err := s.RangeN(5, 100)
		require.NoError(t, err)
		assert.True(t, len(items) >= 5)
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
	s := requests.NewStore[*ocr3.ReportRequest]()
	rid := uuid.New().String()
	req := &ocr3.ReportRequest{
		WorkflowExecutionID: rid,
	}

	err := s.Add(req)
	require.NoError(t, err)
	reqs, err := s.FirstN(10)
	require.NoError(t, err)

	assert.Len(t, reqs, 1)

	reqs = s.GetByIDs([]string{rid})
	assert.Len(t, reqs, 1)

	_, ok := s.Evict(rid)
	assert.True(t, ok)
	reqs, err = s.FirstN(10)
	require.NoError(t, err)
	assert.Len(t, reqs, 0)

	err = s.Add(req)
	require.NoError(t, err)
	reqs, err = s.FirstN(10)
	require.NoError(t, err)
	assert.Len(t, reqs, 1)
}

func TestOCR3Store_ReadRequestsCopy(t *testing.T) {
	s := requests.NewStore[*ocr3.ReportRequest]()
	rid := uuid.New().String()
	cb := make(chan ocr3.ReportResponse, 1)
	stopCh := make(chan struct{}, 1)
	obs, err := values.NewList(
		[]any{"hello", 1},
	)
	require.NoError(t, err)
	req := &ocr3.ReportRequest{
		WorkflowExecutionID:      rid,
		WorkflowID:               "wid",
		WorkflowName:             "name",
		WorkflowOwner:            "owner",
		WorkflowDonID:            1,
		WorkflowDonConfigVersion: 1,
		ReportID:                 "001",
		KeyID:                    "key-001",

		CallbackCh:   cb,
		StopCh:       stopCh,
		Observations: obs,
	}

	require.NoError(t, s.Add(req))

	testCases := []struct {
		name string
		get  func(ctx context.Context, rid string) *ocr3.ReportRequest
	}{
		{
			name: "get",
			get: func(ctx context.Context, rid string) *ocr3.ReportRequest {
				return s.Get(rid)
			},
		},
		{
			name: "firstN",
			get: func(ctx context.Context, rid string) *ocr3.ReportRequest {
				rs, err2 := s.FirstN(1)
				require.NoError(t, err2)
				assert.Len(t, rs, 1)
				return rs[0]
			},
		},
		{
			name: "getByIDs",
			get: func(ctx context.Context, rid string) *ocr3.ReportRequest {
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
			assert.Equal(t, req, gr2)

			gr.StopCh <- struct{}{}
			<-stopCh

			gr.CallbackCh <- ocr3.ReportResponse{}
			<-cb
		})
	}
}
