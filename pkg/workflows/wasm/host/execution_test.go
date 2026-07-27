package host

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

type callCapAsyncErrorStub struct {
	err error
}

func (s *callCapAsyncErrorStub) CallCapability(_ context.Context, _ *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error) {
	return nil, s.err
}

func (s *callCapAsyncErrorStub) GetSecrets(context.Context, *sdkpb.GetSecretsRequest) ([]*sdkpb.SecretResponse, error) {
	return nil, nil
}

func (s *callCapAsyncErrorStub) GetWorkflowExecutionID() string { return "" }

func (s *callCapAsyncErrorStub) GetNodeTime() time.Time { return time.Time{} }

func (s *callCapAsyncErrorStub) GetDONTime() (time.Time, error) { return time.Time{}, nil }

func (s *callCapAsyncErrorStub) EmitUserLog(string) error { return nil }

func (s *callCapAsyncErrorStub) EmitUserMetric(context.Context, *wfpb.WorkflowUserMetric) error {
	return nil
}

var _ ExecutionHelper = (*callCapAsyncErrorStub)(nil)

func TestCallCapAsync_errorSerialization(t *testing.T) {

	capErr := caperrors.NewPublicUserError(errors.New("capability failed"), caperrors.InvalidArgument)
	plainErr := errors.New("plain error")

	tests := []struct {
		name       string
		err        error
		wantErrStr string
	}{
		{
			name:       "capability error is serialized",
			err:        capErr,
			wantErrStr: capErr.SerializeToString(),
		},
		{
			name:       "non-capability error is not serialized",
			err:        plainErr,
			wantErrStr: plainErr.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			exec := newTestExec(1, &callCapAsyncErrorStub{err: tt.err})
			exec.ctx = t.Context()

			const callbackID int32 = 1
			require.NoError(t, exec.callCapAsync(t.Context(), &sdkpb.CapabilityRequest{CallbackId: callbackID}))

			awaitResp, err := exec.awaitCapabilities(t.Context(), &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{callbackID}})
			require.NoError(t, err)
			require.Len(t, awaitResp.Responses, 1)

			capResp := awaitResp.Responses[callbackID]
			require.NotNil(t, capResp)

			errResp, ok := capResp.Response.(*sdkpb.CapabilityResponse_Error)
			require.True(t, ok, "expected error response")
			require.Equal(t, tt.wantErrStr, errResp.Error)

			if tt.err == capErr {
				require.NotEqual(t, capErr.Error(), errResp.Error, "capability errors must use SerializeToString, not Error()")
			}
		})
	}
}

// TestCallCapAsync_DuplicateCallbackIdRejected proves a second callCapAsync
// call reusing a CallbackId that is still in flight is rejected, and that the
// rejection does not disturb the original in-flight call: it must still be
// awaitable and return the real response, not hang until ctx.Done().
func TestCallCapAsync_DuplicateCallbackIdRejected(t *testing.T) {
	t.Parallel()

	stub := &slowCapStub{delay: 0}
	exec := newTestExec(10, stub)

	ctx := t.Context()
	const callbackID int32 = 1

	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: callbackID}))

	err := exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: callbackID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate callback id not allowed")

	// The original call must be unaffected by the rejected duplicate.
	awaitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	resp, err := exec.awaitCapabilities(awaitCtx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{callbackID}})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 1)
	require.NotNil(t, resp.Responses[callbackID])
}

// TestCallCapAsync_DuplicateCallbackIdDoesNotLeakLimiterSlot ensures a
// rejected duplicate call frees the pool-limiter slot it acquired, so it
// doesn't starve subsequent legitimate calls.
func TestCallCapAsync_DuplicateCallbackIdDoesNotLeakLimiterSlot(t *testing.T) {
	t.Parallel()

	const max = 2
	stub := &slowCapStub{delay: 0}
	exec := newTestExec(max, stub)

	ctx := t.Context()

	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 1}))
	require.Error(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: 1}))

	assert.Eventually(t, func() bool {
		avail, err := exec.pendingCallsLimiter.Available(ctx)
		return err == nil && avail == max-1
	}, time.Second, 5*time.Millisecond,
		"rejected duplicate call leaked a pool-limiter slot")
}

// TestGetSecretsAsync_DuplicateCallbackIdRejected mirrors the callCapAsync
// duplicate-rejection behavior for the secrets path.
func TestGetSecretsAsync_DuplicateCallbackIdRejected(t *testing.T) {
	t.Parallel()

	stub := &slowCapStub{delay: 0}
	exec := newTestExec(10, stub)

	ctx := t.Context()
	const callbackID int32 = 1

	require.NoError(t, exec.getSecretsAsync(ctx, &sdkpb.GetSecretsRequest{CallbackId: callbackID}))

	err := exec.getSecretsAsync(ctx, &sdkpb.GetSecretsRequest{CallbackId: callbackID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate callback id not allowed")

	awaitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	resp, err := exec.awaitSecrets(awaitCtx, &sdkpb.AwaitSecretsRequest{Ids: []int32{callbackID}})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 1)
}

// TestCallbackId_CapAndSecretNamespacesDoNotCollide proves a capability call
// and a secrets call may reuse the same numeric CallbackId without tripping
// the duplicate check, since the two ID spaces are tracked independently.
func TestCallbackId_CapAndSecretNamespacesDoNotCollide(t *testing.T) {
	t.Parallel()

	stub := &slowCapStub{delay: 0}
	exec := newTestExec(10, stub)

	ctx := t.Context()
	const callbackID int32 = 42

	require.NoError(t, exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{CallbackId: callbackID}))
	require.NoError(t, exec.getSecretsAsync(ctx, &sdkpb.GetSecretsRequest{CallbackId: callbackID}))

	awaitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	capResp, err := exec.awaitCapabilities(awaitCtx, &sdkpb.AwaitCapabilitiesRequest{Ids: []int32{callbackID}})
	require.NoError(t, err)
	require.Len(t, capResp.Responses, 1)

	secretResp, err := exec.awaitSecrets(awaitCtx, &sdkpb.AwaitSecretsRequest{Ids: []int32{callbackID}})
	require.NoError(t, err)
	require.Len(t, secretResp.Responses, 1)
}
