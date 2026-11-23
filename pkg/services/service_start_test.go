package services_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// startRecorder is a sub-service that starts successfully and records whether Close ran.
type startRecorder struct {
	closed atomic.Bool
}

func (s *startRecorder) Start(context.Context) error { return nil }
func (s *startRecorder) Close() error                { s.closed.Store(true); return nil }
func (s *startRecorder) Name() string                { return "startRecorder" }
func (s *startRecorder) Ready() error                { return nil }
func (s *startRecorder) Healthy() error              { return nil }
func (s *startRecorder) HealthReport() map[string]error {
	return map[string]error{s.Name(): nil}
}

// A parent's Start hook failing after its sub-services started must roll them back:
// Close refuses from the StartFailed state, so nothing else would ever close them.
func TestStartHookFailureRollsBackSubServices(t *testing.T) {
	sub := &startRecorder{}
	parent := services.Config{
		Name:           "parent",
		NewSubServices: func(logger.Logger) []services.Service { return []services.Service{sub} },
		Start:          func(context.Context) error { return errors.New("hook failed") },
	}.NewService(logger.Test(t))

	err := parent.Start(t.Context())
	require.ErrorContains(t, err, "hook failed")
	assert.True(t, sub.closed.Load(), "the started sub-service was rolled back by the failed start")
}
