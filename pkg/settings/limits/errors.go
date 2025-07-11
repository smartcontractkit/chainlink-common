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
