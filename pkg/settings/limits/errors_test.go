package limits

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func TestErrorRateLimited(t *testing.T) {
	wrapped := errors.New("wrapper")
	for _, tt := range []struct {
		name string
		err  ErrorRateLimited
		exp  string
	}{
		{
			name: "full",
			err: ErrorRateLimited{
				Key:    "foo",
				Scope:  settings.ScopeWorkflow,
				Tenant: "wf",
				N:      42,
				Err:    wrapped,
			},
			exp: "foo rate limited for workflow[wf]: wrapper",
		},
		{
			name: "no-err",
			err: ErrorRateLimited{
				Key:    "foo",
				Scope:  settings.ScopeWorkflow,
				Tenant: "wf",
				N:      42,
			},
			exp: "foo rate limited for workflow[wf]",
		},
		{
			name: "no-err-tenant",
			err: ErrorRateLimited{
				Key: "foo",
				N:   42,
			},
			exp: "foo rate limited",
		},
		{
			name: "no-err-tenant-key",
			err: ErrorRateLimited{
				N: 42,
			},
			exp: "rate limited",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, error(tt.err), ErrorRateLimited{})
			assert.EqualError(t, tt.err, tt.exp)
			require.Equal(t, codes.ResourceExhausted, status.Code(tt.err))

			got := marshalUnmarshalError(t, tt.err)

			assert.ErrorContains(t, got, tt.exp)
			require.Equal(t, codes.ResourceExhausted, status.Code(got))
		})
	}
}

func TestErrorResourceLimited(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  ErrorResourceLimited[int]
		exp  string
	}{
		{
			name: "full",
			err: ErrorResourceLimited[int]{
				Key:    "foo",
				Scope:  settings.ScopeWorkflow,
				Tenant: "wf",
				Limit:  100,
				Used:   42,
				Amount: 13,
			},
			exp: "foo resource limited for workflow[wf]: cannot use 13, already using 42/100",
		},
		{
			name: "no-tenant",
			err: ErrorResourceLimited[int]{
				Key:    "foo",
				Scope:  settings.ScopeWorkflow,
				Limit:  100,
				Used:   42,
				Amount: 13,
			},
			exp: "foo resource limited: cannot use 13, already using 42/100",
		},
		{
			name: "no-tenant-key",
			err: ErrorResourceLimited[int]{
				Limit:  100,
				Used:   42,
				Amount: 13,
			},
			exp: "resource limited: cannot use 13, already using 42/100",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, error(tt.err), ErrorResourceLimited[int]{})
			assert.EqualError(t, tt.err, tt.exp)
			require.Equal(t, codes.ResourceExhausted, status.Code(tt.err))

			got := marshalUnmarshalError(t, tt.err)

			assert.ErrorContains(t, got, tt.exp)
			require.Equal(t, codes.ResourceExhausted, status.Code(got))
		})
	}
}

func TestErrorTimeLimited(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  ErrorTimeLimited
		exp  string
	}{
		{
			name: "full",
			err: ErrorTimeLimited{
				Key:     "foo",
				Scope:   settings.ScopeWorkflow,
				Tenant:  "wf",
				Timeout: time.Minute,
			},
			exp: "foo time limited for workflow[wf] to 1m0s",
		},
		{
			name: "no-tenant",
			err: ErrorTimeLimited{
				Key:     "foo",
				Scope:   settings.ScopeWorkflow,
				Timeout: time.Minute,
			},
			exp: "foo time limited to 1m0s",
		},
		{
			name: "no-tenant-key",
			err: ErrorTimeLimited{
				Timeout: time.Minute,
			},
			exp: "time limited to 1m0s",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, error(tt.err), ErrorTimeLimited{})
			assert.EqualError(t, tt.err, tt.exp)
			require.Equal(t, codes.DeadlineExceeded, status.Code(tt.err))

			got := marshalUnmarshalError(t, tt.err)

			assert.ErrorContains(t, got, tt.exp)
			require.Equal(t, codes.DeadlineExceeded, status.Code(got))
		})
	}
}

// Round-trip marshal/unmarshal to simulated grpc call
func marshalUnmarshalError(t *testing.T, err error) error {
	s, ok := status.FromError(err)
	require.True(t, ok)
	b, err := proto.Marshal(s.Proto())
	require.NoError(t, err)
	var pb spb.Status
	require.NoError(t, proto.Unmarshal(b, &pb))
	return status.FromProto(&pb).Err()
}
