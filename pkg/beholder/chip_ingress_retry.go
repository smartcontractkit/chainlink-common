package beholder

import (
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// ChipIngressRetryConfig is the configuration-layer form of chipingress.RetryPolicy: a master
// switch plus optional overrides, tolerant of zero values, for hosts that build the policy
// from string-based configuration (node TOML, LOOP plugin env vars). RetryPolicy resolves it
// into the *chipingress.RetryPolicy that Config.ChipIngressRetryPolicy expects.
type ChipIngressRetryConfig struct {
	// Enabled turns gRPC-level retries on for the chip ingress client. Off by default: the
	// policy is service-wide and replays non-idempotent RPCs such as Publish (see
	// chipingress.WithRetryPolicy).
	Enabled bool
	// MaxAttempts, when > 0, overrides chipingress.DefaultRetryPolicy's MaxAttempts.
	MaxAttempts int
	// InitialBackoff, when > 0, overrides chipingress.DefaultRetryPolicy's InitialBackoff.
	InitialBackoff time.Duration
	// MaxBackoff, when > 0, overrides chipingress.DefaultRetryPolicy's MaxBackoff.
	MaxBackoff time.Duration
	// BackoffMultiplier, when > 0, overrides chipingress.DefaultRetryPolicy's
	// BackoffMultiplier.
	BackoffMultiplier float64
	// RetryableStatusCodes, when non-empty, lists gRPC status code names (per
	// chipingress.ParseStatusCodes) that are eligible for retry; empty keeps the defaults.
	RetryableStatusCodes []string
}

// RetryPolicy resolves c into a *chipingress.RetryPolicy for Config.ChipIngressRetryPolicy.
// It returns (nil, nil) when c.Enabled is false (retries stay off), and otherwise starts from
// chipingress.DefaultRetryPolicy and applies the non-zero overrides, so a config layer that
// only sets Enabled still gets a valid, safe policy rather than a zero-valued one that gRPC
// would silently discard (see chipingress' retry_policy_test.go for that failure mode).
func (c ChipIngressRetryConfig) RetryPolicy() (*chipingress.RetryPolicy, error) {
	if !c.Enabled {
		return nil, nil
	}
	p := chipingress.DefaultRetryPolicy()
	if c.MaxAttempts > 0 {
		p.MaxAttempts = c.MaxAttempts
	}
	if c.InitialBackoff > 0 {
		p.InitialBackoff = c.InitialBackoff
	}
	if c.MaxBackoff > 0 {
		p.MaxBackoff = c.MaxBackoff
	}
	if c.BackoffMultiplier > 0 {
		p.BackoffMultiplier = c.BackoffMultiplier
	}
	if len(c.RetryableStatusCodes) > 0 {
		retryable, err := chipingress.ParseStatusCodes(c.RetryableStatusCodes)
		if err != nil {
			return nil, fmt.Errorf("invalid retryable status codes: %w", err)
		}
		p.RetryableStatusCodes = retryable
	}
	return &p, nil
}
