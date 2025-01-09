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
	require.NoError(tb, r.Start(tests.Context(tb)), "service failed to start")
	tb.Cleanup(func() {
		tb.Helper()
		assert.NoError(tb, r.Close(), "error closing service")
	})
	return r
}

// RunHealthy fails tb if the service fails to start, close, is never ready, or is ever unhealthy (based on periodic checks).
//   - after starting, readiness will always be checked at least once, before closing
//   - if ever ready, then health will be checked at least once, before closing
func RunHealthy[S services.Service](tb TestingT, s S) S {
	tb.Helper()
	Run(tb, s)

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
		for s.Ready() != nil {
			select {
			case <-done:
				if assert.NoError(tb, s.Ready(), "service never ready") {
					assert.NoError(tb, hp(), "service unhealthy")
				}
				return
			case <-time.After(time.Second):
			}
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
	return s
}
