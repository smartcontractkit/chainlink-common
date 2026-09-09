package chipingress

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// RetryPolicy configures the gRPC-level retry behavior the client installs as its default
// service config (see WithRetryPolicy). It is the Go-native, mistake-resistant equivalent of
// hand-writing a gRPC "retryPolicy" service-config JSON block, which is easy to get wrong: the
// nesting (methodConfig[].retryPolicy, not top-level), the duration encoding (protobuf JSON
// duration, e.g. "0.1s", not Go's "100ms"), and the status-code names are all silently ignored
// by gRPC's parser if malformed, rather than rejected - see the client_test.go regression test
// for what that looks like.
type RetryPolicy struct {
	// MaxAttempts is the maximum number of call attempts, including the original attempt.
	// Must be 2 or greater for gRPC to treat the policy as valid; otherwise gRPC silently
	// installs no retry policy at all for the affected method.
	MaxAttempts int
	// InitialBackoff is the base delay before the first retry attempt.
	InitialBackoff time.Duration
	// MaxBackoff caps the exponential backoff delay between attempts.
	MaxBackoff time.Duration
	// BackoffMultiplier is applied to the backoff delay after each attempt.
	BackoffMultiplier float64
	// RetryableStatusCodes is the set of gRPC status codes eligible for retry.
	RetryableStatusCodes []codes.Code
}

// RetryThrottlingPolicy configures gRPC's retry throttling, which limits how much additional
// load retries (and hedged RPCs) can place on an already-struggling server. See
// https://github.com/grpc/proposal/blob/master/A6-client-retries.md#integration-with-service-config.
type RetryThrottlingPolicy struct {
	// MaxTokens is the starting/maximum size of the retry token bucket. Must be in (0, 1000].
	MaxTokens float64
	// TokenRatio is the number of tokens restored per successful RPC. Must be > 0.
	TokenRatio float64
}

// defaultRetryPolicy is the retry policy installed when no WithRetryPolicy override is
// supplied. It matches the values the client has always intended to use (see the historical,
// malformed service-config JSON this replaced), except that the backoff durations are now
// correctly encoded so gRPC actually installs the policy.
func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:          3,
		InitialBackoff:       100 * time.Millisecond,
		MaxBackoff:           1 * time.Second,
		BackoffMultiplier:    2,
		RetryableStatusCodes: []codes.Code{codes.Unavailable, codes.ResourceExhausted},
	}
}

// defaultRetryThrottlingPolicy caps retry amplification against a degraded server. Values
// follow the example in the gRPC retry design doc (A6): a bucket of 10 tokens, refilled by 0.1
// per success, so retries taper off quickly once failures dominate.
func defaultRetryThrottlingPolicy() RetryThrottlingPolicy {
	return RetryThrottlingPolicy{
		MaxTokens:  10,
		TokenRatio: 0.1,
	}
}

// grpcServiceConfigDuration renders d in the string form gRPC's service-config parser requires
// (grpc-go's internal/serviceconfig.Duration, which follows protobuf's JSON duration encoding):
// a decimal number of seconds followed by "s", e.g. "0.1s" for 100ms. Plain Go duration strings
// such as "100ms" are NOT accepted by that parser and fail JSON unmarshaling.
func grpcServiceConfigDuration(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "s"
}

// grpcStatusCodeName returns the canonical gRPC service-config string name for a status code
// (e.g. "UNAVAILABLE"), per https://github.com/grpc/grpc/blob/master/doc/statuscodes.md. Falls
// back to the numeric code for values without a well-known name; gRPC's parser accepts either
// form.
func grpcStatusCodeName(c codes.Code) string {
	switch c {
	case codes.OK:
		return "OK"
	case codes.Canceled:
		return "CANCELLED"
	case codes.Unknown:
		return "UNKNOWN"
	case codes.InvalidArgument:
		return "INVALID_ARGUMENT"
	case codes.DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case codes.NotFound:
		return "NOT_FOUND"
	case codes.AlreadyExists:
		return "ALREADY_EXISTS"
	case codes.PermissionDenied:
		return "PERMISSION_DENIED"
	case codes.ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case codes.FailedPrecondition:
		return "FAILED_PRECONDITION"
	case codes.Aborted:
		return "ABORTED"
	case codes.OutOfRange:
		return "OUT_OF_RANGE"
	case codes.Unimplemented:
		return "UNIMPLEMENTED"
	case codes.Internal:
		return "INTERNAL"
	case codes.Unavailable:
		return "UNAVAILABLE"
	case codes.DataLoss:
		return "DATA_LOSS"
	case codes.Unauthenticated:
		return "UNAUTHENTICATED"
	default:
		return strconv.Itoa(int(c))
	}
}

// serviceConfigJSON mirrors (the small subset of) the gRPC service-config JSON schema this
// package needs: https://github.com/grpc/grpc/blob/master/doc/service_config.md. Building the
// JSON via these typed structs (instead of hand-written string literals) is the fix for the bug
// this package shipped previously: a flat, malformed JSON blob whose fields gRPC's parser
// silently discarded as unknown, installing no retry policy at all without any error.
type serviceConfigJSON struct {
	MethodConfig    []methodConfigJSON   `json:"methodConfig"`
	RetryThrottling *retryThrottlingJSON `json:"retryThrottling,omitempty"`
}

type methodConfigJSON struct {
	Name        []methodNameJSON `json:"name"`
	RetryPolicy retryPolicyJSON  `json:"retryPolicy"`
}

type methodNameJSON struct {
	Service string `json:"service"`
}

type retryPolicyJSON struct {
	MaxAttempts          int      `json:"maxAttempts"`
	InitialBackoff       string   `json:"initialBackoff"`
	MaxBackoff           string   `json:"maxBackoff"`
	BackoffMultiplier    float64  `json:"backoffMultiplier"`
	RetryableStatusCodes []string `json:"retryableStatusCodes"`
}

type retryThrottlingJSON struct {
	MaxTokens  float64 `json:"maxTokens"`
	TokenRatio float64 `json:"tokenRatio"`
}

// buildRetryServiceConfigJSON marshals policy (and, if non-nil, throttling) into the gRPC
// default-service-config JSON that grpc.WithDefaultServiceConfig expects, scoped to the
// ChipIngress gRPC service via its fully-qualified name (from the generated
// pb.ChipIngress_ServiceDesc, so it can't drift from the actual service).
func buildRetryServiceConfigJSON(policy RetryPolicy, throttling *RetryThrottlingPolicy) (string, error) {
	codeNames := make([]string, 0, len(policy.RetryableStatusCodes))
	for _, c := range policy.RetryableStatusCodes {
		codeNames = append(codeNames, grpcStatusCodeName(c))
	}

	sc := serviceConfigJSON{
		MethodConfig: []methodConfigJSON{
			{
				Name: []methodNameJSON{{Service: pb.ChipIngress_ServiceDesc.ServiceName}},
				RetryPolicy: retryPolicyJSON{
					MaxAttempts:          policy.MaxAttempts,
					InitialBackoff:       grpcServiceConfigDuration(policy.InitialBackoff),
					MaxBackoff:           grpcServiceConfigDuration(policy.MaxBackoff),
					BackoffMultiplier:    policy.BackoffMultiplier,
					RetryableStatusCodes: codeNames,
				},
			},
		},
	}
	if throttling != nil {
		sc.RetryThrottling = &retryThrottlingJSON{
			MaxTokens:  throttling.MaxTokens,
			TokenRatio: throttling.TokenRatio,
		}
	}

	b, err := json.Marshal(sc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal retry policy service config: %w", err)
	}
	return string(b), nil
}

// WithRetryPolicy overrides the client's default gRPC-level retry policy for the ChipIngress
// service. If not supplied, the client retries UNAVAILABLE and RESOURCE_EXHAUSTED failures up
// to 3 times with exponential backoff starting at 100ms and capped at 1s (see
// defaultRetryPolicy). Retry throttling (see RetryThrottlingPolicy) is always applied alongside
// whichever retry policy is in effect, so a struggling server isn't amplified by retries.
func WithRetryPolicy(policy RetryPolicy) Opt {
	return func(c *clientConfig) { c.retryPolicy = policy }
}
