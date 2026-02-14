package limits

import (
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

type ErrorRateLimited struct {
	Key string

	Scope  settings.Scope
	Tenant string

	N int

	Err error
}

func (e ErrorRateLimited) Unwrap() error { return e.Err }

func (e ErrorRateLimited) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorRateLimited) Is(target error) bool {
	_, ok := target.(ErrorRateLimited) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorRateLimited) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	msg := fmt.Sprintf("%srate limited%s: request rate has exceeded the allowed limit. Please reduce request frequency or wait before retrying", which, who)
	if e.Err == nil {
		return msg
	}
	return fmt.Sprintf("%s: %v", msg, e.Err)
}

type ErrorResourceLimited[N Number] struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Used, Limit, Amount N
}

func (e ErrorResourceLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorResourceLimited[N]) Is(target error) bool {
	_, ok := target.(ErrorResourceLimited[N]) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorResourceLimited[N]) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%sresource limited%s: cannot allocate %v, already using %v of %v maximum. Free existing resources or request a limit increase", which, who, e.Amount, e.Used, e.Limit)
}

type ErrorTimeLimited struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Timeout time.Duration
}

func (e ErrorTimeLimited) GRPCStatus() *status.Status {
	return status.New(codes.DeadlineExceeded, e.Error())
}

func (e ErrorTimeLimited) Is(target error) bool {
	_, ok := target.(ErrorTimeLimited)
	return ok
}

func (e ErrorTimeLimited) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%stime limited%s: operation exceeded the maximum allowed duration of %s. Consider simplifying the operation or requesting a timeout increase", which, who, e.Timeout)
}

func errArgs(key string, scope settings.Scope, tenant string) (which, who string) {
	if key != "" {
		which = key + " "
	}
	if tenant != "" {
		who = " for " + scope.String() + "[" + tenant + "]"
	}
	return
}

type ErrorBoundLimited[N Number] struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Limit, Amount N
}

func (e ErrorBoundLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorBoundLimited[N]) Is(target error) bool {
	_, ok := target.(ErrorBoundLimited[N]) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorBoundLimited[N]) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: cannot use %v, maximum allowed is %v. Reduce usage or request a limit increase", which, who, e.Amount, e.Limit)
}

type ErrorQueueFull struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Limit int
}

func (e ErrorQueueFull) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorQueueFull) Is(target error) bool {
	_, ok := target.(ErrorQueueFull) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorQueueFull) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: queue is full (capacity: %d). Incoming items are being dropped. Consider reducing submission rate or requesting a capacity increase", which, who, e.Limit)
}

var ErrQueueEmpty = fmt.Errorf("queue is empty")

type ErrorNotAllowed struct {
	Key string

	Scope  settings.Scope
	Tenant string
}

func (e ErrorNotAllowed) GRPCStatus() *status.Status {
	return status.New(codes.PermissionDenied, e.Error())
}

func (e ErrorNotAllowed) Is(target error) bool {
	_, ok := target.(ErrorNotAllowed) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorNotAllowed) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: operation not allowed. This action is restricted by current configuration or permissions", which, who)
}
