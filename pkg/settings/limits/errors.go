package limits

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// ErrMissingTenant is returned when a scoped limiter is used without the tenant that its
// scope requires in ctx. Unlike a settings read failure, where the limiter still resolves a
// usable value alongside the error, no lookup is attempted here, so callers must not treat
// the returned value as a limit. It signals that the CRE context was never populated, which
// is a programming error rather than a degraded system.
type ErrMissingTenant struct {
	Scope settings.Scope
}

func (e ErrMissingTenant) Is(target error) bool {
	var errorMissingTenant ErrMissingTenant
	return errors.As(target, &errorMissingTenant)
}

func (e ErrMissingTenant) Error() string {
	return fmt.Sprintf("missing tenant for scope: %s", e.Scope)
}

// IsErrRecoverable reports whether err still left the caller a usable value. Limiters fall back
// to the compiled default when a settings read fails, so those errors are advisory and the
// returned limit can still be enforced. It reports false when no value was resolved and
// enforcing the zero value would be wrong.
//
// Unknown errors are treated as recoverable, so that a degraded settings service does not
// cause callers to drop work. Match on this rather than on specific error types, so callers
// pick up future non-recoverable cases automatically.
func IsErrRecoverable(err error) bool {
	return !errors.Is(err, ErrMissingTenant{})
}

// LimitError is implemented by errors returned when a limit is exceeded.
// Use [errors.As] to identify limit errors, for example:
//
//	var limitErr LimitError
//	if errors.As(err, &limitErr) { ... }
type LimitError interface {
	error
	limitError()
}

type ErrorRateLimited struct {
	Key string

	Scope  settings.Scope
	Tenant string

	N int

	Err error
}

func (ErrorRateLimited) limitError() {}

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
	msg := fmt.Sprintf("%srate limited%s", which, who)
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

func (ErrorResourceLimited[N]) limitError() {}

func (e ErrorResourceLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorResourceLimited[N]) Is(target error) bool {
	_, ok := target.(ErrorResourceLimited[N]) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorResourceLimited[N]) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%sresource limited%s: cannot use %v, already using %v/%v", which, who, e.Amount, e.Used, e.Limit)
}

type ErrorTimeLimited struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Timeout time.Duration
}

func (ErrorTimeLimited) limitError() {}

func (e ErrorTimeLimited) GRPCStatus() *status.Status {
	return status.New(codes.DeadlineExceeded, e.Error())
}

func (e ErrorTimeLimited) Is(target error) bool {
	_, ok := target.(ErrorTimeLimited)
	return ok
}

func (e ErrorTimeLimited) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%stime limited%s to %s", which, who, e.Timeout)
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

func (ErrorBoundLimited[N]) limitError() {}

func (e ErrorBoundLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorBoundLimited[N]) Is(target error) bool {
	_, ok := target.(ErrorBoundLimited[N]) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorBoundLimited[N]) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: cannot use %v, limit is %v", which, who, e.Amount, e.Limit)
}

type ErrorRangeLimited[N Number] struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Limit  settings.Range[N]
	Amount N
}

func (ErrorRangeLimited[N]) limitError() {}

func (e ErrorRangeLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorRangeLimited[N]) Is(target error) bool {
	_, ok := target.(ErrorRangeLimited[N]) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorRangeLimited[N]) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: cannot use %v, limited to range %v", which, who, e.Amount, e.Limit)
}

type ErrorQueueFull struct {
	Key string

	Scope  settings.Scope
	Tenant string

	Limit int
}

func (ErrorQueueFull) limitError() {}

func (e ErrorQueueFull) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorQueueFull) Is(target error) bool {
	_, ok := target.(ErrorQueueFull) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorQueueFull) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: queue of %d is full", which, who, e.Limit)
}

var ErrQueueEmpty = errors.New("queue is empty")

type ErrorNotAllowed struct {
	Key string

	Scope  settings.Scope
	Tenant string
}

func (ErrorNotAllowed) limitError() {}

func (e ErrorNotAllowed) GRPCStatus() *status.Status {
	return status.New(codes.PermissionDenied, e.Error())
}

func (e ErrorNotAllowed) Is(target error) bool {
	_, ok := target.(ErrorNotAllowed) //nolint:errcheck // implementing errors.Is
	return ok
}

func (e ErrorNotAllowed) Error() string {
	which, who := errArgs(e.Key, e.Scope, e.Tenant)
	return fmt.Sprintf("%slimited%s: not allowed", which, who)
}
