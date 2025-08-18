package requests_test

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"

	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func Test_Handler_SendsResponse(t *testing.T) {
	lggr := logger.Test(t)
	ctx := t.Context()

	h := requests.NewHandler(lggr, requests.NewStore[*ocr3.ReportRequest](), clockwork.NewFakeClockAt(time.Now()), 1*time.Second)
	servicetest.Run(t, h)

	responseCh := make(chan ocr3.ReportResponse, 10)
	h.SendRequest(ctx, &ocr3.ReportRequest{
		WorkflowExecutionID: "test",
		CallbackCh:          responseCh,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	})

	testVal, err := values.NewMap(map[string]any{"result": "testval"})
	require.NoError(t, err)

	h.SendResponse(ctx, ocr3.ReportResponse{
		WorkflowExecutionID: "test",
		Value:               testVal,
		Err:                 nil,
	})

	resp := <-responseCh
	require.Equal(t, testVal, resp.Value)
}

func Test_Handler_SendsResponseToLateRequest(t *testing.T) {
	lggr := logger.Test(t)
	ctx := t.Context()

	h := requests.NewHandler(lggr, requests.NewStore[*ocr3.ReportRequest](), clockwork.NewFakeClockAt(time.Now()), 1*time.Second)
	servicetest.Run(t, h)

	testVal, err := values.NewMap(map[string]any{"result": "testval"})
	require.NoError(t, err)
	h.SendResponse(ctx, ocr3.ReportResponse{
		WorkflowExecutionID: "test",
		Value:               testVal,
		Err:                 nil,
	})

	responseCh := make(chan ocr3.ReportResponse, 10)
	h.SendRequest(ctx, &ocr3.ReportRequest{
		WorkflowExecutionID: "test",
		CallbackCh:          responseCh,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	})

	resp := <-responseCh
	require.Equal(t, testVal, resp.Value)
}

func Test_Handler_SendsResponseToLateRequestOnlyOnce(t *testing.T) {
	lggr := logger.Test(t)
	ctx := t.Context()

	h := requests.NewHandler(lggr, requests.NewStore[*ocr3.ReportRequest](), clockwork.NewFakeClockAt(time.Now()), 1*time.Second)
	servicetest.Run(t, h)

	testVal, err := values.NewMap(map[string]any{"result": "testval"})
	require.NoError(t, err)

	h.SendResponse(ctx, ocr3.ReportResponse{
		WorkflowExecutionID: "test",
		Value:               testVal,
		Err:                 nil,
	})

	responseCh := make(chan ocr3.ReportResponse, 10)
	h.SendRequest(ctx, &ocr3.ReportRequest{
		WorkflowExecutionID: "test",
		CallbackCh:          responseCh,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	})

	require.NoError(t, err)

	resp := <-responseCh
	require.Equal(t, testVal, resp.Value)

	responseCh = make(chan ocr3.ReportResponse, 10)
	h.SendRequest(ctx, &ocr3.ReportRequest{
		WorkflowExecutionID: "test",
		CallbackCh:          responseCh,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	})

	select {
	case <-responseCh:
		t.Fatal("Should not have received a response")
	default:
	}
}

func Test_Handler_PendingRequestsExpiry(t *testing.T) {
	ctx := t.Context()

	lggr := logger.Test(t)
	clock := clockwork.NewFakeClockAt(time.Now())
	h := requests.NewHandler(lggr, requests.NewStore[*ocr3.ReportRequest](), clock, 1*time.Second)
	servicetest.Run(t, h)

	responseCh := make(chan ocr3.ReportResponse, 10)
	h.SendRequest(ctx, &ocr3.ReportRequest{
		WorkflowExecutionID: "test",
		CallbackCh:          responseCh,
		ExpiresAt:           time.Now().Add(1 * time.Second),
	})

	clock.Advance(2 * time.Second)

	resp := <-responseCh

	assert.ErrorContains(t, resp.Err, "timeout exceeded: could not process request before expiry")
}
