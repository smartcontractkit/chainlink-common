package servicetest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type Runnable interface {
	Start(context.Context) error
	Close() error
}

type TestingT interface {
	require.TestingT
	Helper()
	Cleanup(func())
}

// Run fails tb if the service fails to start or close.
func Run[R Runnable](tb TestingT, r R) R {
	tb.Helper()
	RunCfg{}.run(tb, r)
	return r
}

// RunHealthy fails tb if the service fails to start, close, is never ready, or is ever unhealthy (based on periodic checks).
//   - after starting, readiness will always be checked at least once, before closing
//   - if ever ready, then health will be checked at least once, before closing
func RunHealthy[S services.Service](tb TestingT, s S) S {
	tb.Helper()
	RunCfg{Healthy: true}.Run(tb, s)
	return s
}

// RunCfg specifies a test configuration for running a service.
// By default, health checks are not enforced, but Start/Close timeout are.
type RunCfg struct {
	// Healthy includes extra checks for whether the service is never ready, or is ever unhealthy (based on periodic checks).
	//   - after starting, readiness will always be checked at least once, before closing
	//   - if ever ready, then health will be checked at least once, before closing
	Healthy bool
	// WaitForReady blocks returning until after Ready() returns nil, after calling Start().
	WaitForReady bool
	// StartTimeout sets a limit for Start which results in an error if exceeded.
	StartTimeout time.Duration
	// StartTimeout sets a limit for Close which results in an error if exceeded.
	CloseTimeout time.Duration
}

func (cfg RunCfg) Run(tb TestingT, s services.Service) {
	tb.Helper()

	cfg.run(tb, s)

	if cfg.WaitForReady {
		ctx := tests.Context(tb)
		cfg.waitForReady(tb, s, ctx.Done())
	}

	if cfg.Healthy {
		cfg.healthCheck(tb, s)
	}
}

func (cfg RunCfg) run(tb TestingT, s Runnable) {
	tb.Helper()
	//TODO remove....set from built-ins? or disallow unbounded, so exceptions must be explicit?
	if cfg.StartTimeout == 0 {
		cfg.StartTimeout = time.Second
	}
	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = time.Second
	}

	start := time.Now()
	require.NoError(tb, s.Start(tests.Context(tb)), "service failed to start: %T", s)
	if elapsed := time.Since(start); cfg.StartTimeout > 0 && elapsed > cfg.StartTimeout {
		tb.Errorf("slow service start: %T.Start() took %s", s, elapsed)
	}

	tb.Cleanup(func() {
		tb.Helper()
		start := time.Now()
		assert.NoError(tb, s.Close(), "error closing service: %T", s)
		if elapsed := time.Since(start); cfg.CloseTimeout > 0 && elapsed > cfg.CloseTimeout {
			tb.Errorf("slow service close: %T.Close() took %s", s, elapsed)
		}
	})
}

func (cfg RunCfg) waitForReady(tb TestingT, s services.Service, done <-chan struct{}) {
	for err := s.Ready(); err != nil; err = s.Ready() {
		select {
		case <-done:
			assert.NoError(tb, err, "service never ready")
			return
		case <-time.After(time.Second):
		}
	}
}

func (cfg RunCfg) healthCheck(tb TestingT, s services.Service) {
	tb.Helper()

	done := make(chan struct{})
	tb.Cleanup(func() {
		done <- struct{}{}
		<-done
	})
	go func() {
		defer close(done)
		hp := func() (err error) {
			for k, v := range s.HealthReport() {
				if v != nil {
					err = errors.Join(err, fmt.Errorf("%s: %w", k, v))
				}
			}
			return
		}
		if !cfg.WaitForReady {
			cfg.waitForReady(tb, s, done)
			assert.NoError(tb, hp(), "service unhealthy")
		}
		for {
			select {
			case <-done:
				assert.NoError(tb, hp(), "service unhealthy")
				return
			case <-time.After(time.Second):
				assert.NoError(tb, hp(), "service unhealthy")
			}
		}
	}()
}
