package batch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
)

func TestErrorCodeFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "partial delivery publish error",
			err:  &batch.PublishError{Code: chipingress.PublishErrorCode(1), Reason: "schema not found"},
			want: chipingress.PublishErrorCode(1).String(),
		},
		{
			name: "deadline exceeded status",
			err:  status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			want: codes.DeadlineExceeded.String(),
		},
		{
			name: "deadline exceeded context",
			err:  context.DeadlineExceeded,
			want: codes.DeadlineExceeded.String(),
		},
		{
			name: "unavailable gateway 502",
			err:  status.Error(codes.Unavailable, `unexpected HTTP status code received from server: 502 (Bad Gateway)`),
			want: codes.Unavailable.String(),
		},
		{
			name: "internal publish failure",
			err:  status.Error(codes.Internal, "failed to publish events"),
			want: codes.Internal.String(),
		},
		{
			name: "buffer full",
			err:  batch.ErrMessageBufferFull,
			want: batch.ErrMessageBufferFull.Error(),
		},
		{
			name: "client shutdown",
			err:  batch.ErrClientShutdown,
			want: batch.ErrClientShutdown.Error(),
		},
		{
			name: "unknown error",
			err:  errors.New("something else"),
			want: batch.ErrorTypeClientError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, batch.ErrorCodeFor(tt.err))
		})
	}
}

func TestErrorCodeFor_nil(t *testing.T) {
	t.Parallel()
	require.Empty(t, batch.ErrorCodeFor(nil))
}
