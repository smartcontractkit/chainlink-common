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

// ErrorCodeFor returns a bounded error code string for a batch send/queue error.
// The returned code is intended for metric dimensions; callers can infer the
// error category from the code:
//   - PUBLISH_ERROR_CODE_*  -> partial_delivery
//   - "buffer_full"         -> buffer_full
//   - "client_shutdown"     -> client_error
//   - gRPC code name        -> rpc_error
//   - "client_error"        -> client_error (fallback for unrecognized client-side errors)
func ErrorCodeFor(err error) string {
	if err == nil {
		return ""
	}

	var pubErr *PublishError
	if errors.As(err, &pubErr) {
		return pubErr.Code.String()
	}

	if errors.Is(err, ErrMessageBufferFull) {
		return ErrMessageBufferFull.Error()
	}
	if errors.Is(err, ErrClientShutdown) {
		return ErrClientShutdown.Error()
	}
	if st, ok := status.FromError(err); ok {
		return st.Code().String()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return codes.DeadlineExceeded.String()
	}
	if errors.Is(err, context.Canceled) {
		return codes.Canceled.String()
	}

	return ErrorTypeClientError
}
