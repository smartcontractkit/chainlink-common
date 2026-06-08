package host

import (
	"context"
	"errors"
	"testing"
	"time"

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
