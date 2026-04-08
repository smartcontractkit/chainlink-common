package host

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// awaitOrderStub implements ExecutionHelper for testing awaitCapabilities ordering.
type awaitOrderStub struct {
	unblock chan struct{}
}

func (a *awaitOrderStub) CallCapability(_ context.Context, req *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error) {
	payload, err := anypb.New(&emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	ok := &sdkpb.CapabilityResponse{
		Response: &sdkpb.CapabilityResponse_Payload{
			Payload: payload,
		},
	}
	if req.CallbackId == 1 {
		<-a.unblock
	}
	return ok, nil
}

func (a *awaitOrderStub) GetSecrets(context.Context, *sdkpb.GetSecretsRequest) ([]*sdkpb.SecretResponse, error) {
	return nil, nil
}

func (a *awaitOrderStub) GetWorkflowExecutionID() string { return "" }

func (a *awaitOrderStub) GetNodeTime() time.Time { return time.Time{} }

func (a *awaitOrderStub) GetDONTime() (time.Time, error) { return time.Time{}, nil }

func (a *awaitOrderStub) EmitUserLog(string) error { return nil }

func (a *awaitOrderStub) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error { return nil }

var _ ExecutionHelper = (*awaitOrderStub)(nil)

// TestAwaitCapabilities_headOfLineBlocksOnEarlierID proves awaitCapabilities receives from
// callback channels in acr.Ids order: it cannot finish until an earlier ID completes, even when
// a later callback finishes first.
func TestAwaitCapabilities_headOfLineBlocksOnEarlierID(t *testing.T) {
	t.Parallel()

	unblock := make(chan struct{})
	stub := &awaitOrderStub{unblock: unblock}

	exec := &execution[*sdkpb.ExecutionResult]{
		ctx:                 t.Context(),
		capabilityResponses: make(map[int32]<-chan *sdkpb.CapabilityResponse),
		executor:            stub,
	}

	req := func(id int32) *sdkpb.CapabilityRequest {
		return &sdkpb.CapabilityRequest{CallbackId: id}
	}

	require.NoError(t, exec.callCapAsync(t.Context(), req(1)))
	require.NoError(t, exec.callCapAsync(t.Context(), req(2)))

	var awaitFinished atomic.Bool
	var awaitResp *sdkpb.AwaitCapabilitiesResponse
	var awaitErr error
	go func() {
		awaitResp, awaitErr = exec.awaitCapabilities(t.Context(), &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{1, 2}})
		awaitFinished.Store(true)
	}()

	time.Sleep(200 * time.Millisecond)
	require.False(t, awaitFinished.Load(), "awaitCapabilities returned before callback 1 was unblocked; head-of-line invariant violated")

	// Unblock callback 1 so the first channel receive in awaitCapabilities can complete.
	close(unblock)

	require.Eventually(t, func() bool { return awaitFinished.Load() }, 2*time.Second, 10*time.Millisecond,
		"awaitCapabilities did not complete after unblocking callback 1")
	require.NoError(t, awaitErr)
	require.Len(t, awaitResp.Responses, 2)
	require.Contains(t, awaitResp.Responses, int32(1))
	require.Contains(t, awaitResp.Responses, int32(2))
}
