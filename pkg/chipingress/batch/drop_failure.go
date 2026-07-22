package batch

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrorTypePartialDelivery = "partial_delivery"
	ErrorTypeRPCError        = "rpc_error"
	ErrorTypeBufferFull      = "buffer_full"
	ErrorTypeClientError     = "client_error"
)

// ErrorCodeFor classifies a batch send/queue error for metric dimensions,
// returning both a coarse-grained error type and a bounded error code.
func ErrorCodeFor(err error) (errorType, errorCode string) {
	if err == nil {
		return "", ""
	}

	var pubErr *PublishError
	if errors.As(err, &pubErr) {
		return ErrorTypePartialDelivery, pubErr.Code.String()
	}

	if errors.Is(err, ErrMessageBufferFull) {
		return ErrorTypeBufferFull, ErrMessageBufferFull.Error()
	}
	if errors.Is(err, ErrClientShutdown) {
		return ErrorTypeClientError, ErrClientShutdown.Error()
	}
	if st, ok := status.FromError(err); ok {
		return ErrorTypeRPCError, st.Code().String()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTypeRPCError, codes.DeadlineExceeded.String()
	}
	if errors.Is(err, context.Canceled) {
		return ErrorTypeRPCError, codes.Canceled.String()
	}

	return ErrorTypeClientError, ErrorTypeClientError
}
