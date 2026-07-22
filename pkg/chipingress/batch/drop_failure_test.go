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
		name     string
		err      error
		wantType string
		wantCode string
	}{
		{
			name:     "partial delivery publish error",
			err:      &batch.PublishError{Code: chipingress.PublishErrorCode(1), Reason: "schema not found"},
			wantType: batch.ErrorTypePartialDelivery,
			wantCode: chipingress.PublishErrorCode(1).String(),
		},
		{
			name:     "results mismatch is classified as partial delivery",
			err:      &batch.PublishError{Code: batch.ErrCodeResultsMismatch, Reason: "server returned 1 results for 2 events"},
			wantType: batch.ErrorTypePartialDelivery,
			wantCode: batch.ErrCodeResultsMismatch.String(),
		},
		{
			name:     "deadline exceeded status",
			err:      status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			wantType: batch.ErrorTypeRPCError,
			wantCode: codes.DeadlineExceeded.String(),
		},
		{
			name:     "deadline exceeded context",
			err:      context.DeadlineExceeded,
			wantType: batch.ErrorTypeRPCError,
			wantCode: codes.DeadlineExceeded.String(),
		},
		{
			name:     "unavailable gateway 502",
			err:      status.Error(codes.Unavailable, `unexpected HTTP status code received from server: 502 (Bad Gateway)`),
			wantType: batch.ErrorTypeRPCError,
			wantCode: codes.Unavailable.String(),
		},
		{
			name:     "internal publish failure",
			err:      status.Error(codes.Internal, "failed to publish events"),
			wantType: batch.ErrorTypeRPCError,
			wantCode: codes.Internal.String(),
		},
		{
			name:     "buffer full",
			err:      batch.ErrMessageBufferFull,
			wantType: batch.ErrorTypeBufferFull,
			wantCode: batch.ErrMessageBufferFull.Error(),
		},
		{
			name:     "client shutdown",
			err:      batch.ErrClientShutdown,
			wantType: batch.ErrorTypeClientError,
			wantCode: batch.ErrClientShutdown.Error(),
		},
		{
			name:     "unknown error",
			err:      errors.New("something else"),
			wantType: batch.ErrorTypeClientError,
			wantCode: batch.ErrorTypeClientError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errorType, errorCode := batch.ErrorCodeFor(tt.err)
			assert.Equal(t, tt.wantType, errorType)
			assert.Equal(t, tt.wantCode, errorCode)
		})
	}
}

func TestErrorCodeFor_nil(t *testing.T) {
	t.Parallel()
	errorType, errorCode := batch.ErrorCodeFor(nil)
	require.Empty(t, errorType)
	require.Empty(t, errorCode)
}
