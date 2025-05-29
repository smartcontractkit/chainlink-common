package services

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/stretchr/testify/require"
)

type mockSvc struct {
	Service
	eng *Engine
}

// Test_EngineGo verifies that a function launced by Engine.Go exists until the
// parent context is done and the engine's stop channel is closed.
func Test_EngineGo(t *testing.T) {
	lggr := logger.Test(t)
	started := make(chan struct{})
	closed := make(chan struct{})

	m := mockSvc{}

	blocker := func(ctx context.Context) {
		<-started
		<-ctx.Done()
		<-m.eng.StopChan
		close(closed)
	}

	start := func(ctx context.Context) error {
		close(started)
		m.eng.Go(blocker)
		return nil
	}

	close := func() error {
		<-closed
		return nil
	}

	svc, eng := Config{
		Name:  "test-service",
		Start: start,
		Close: close,
	}.NewServiceEngine(lggr)

	m.eng = eng

	err := svc.Start(t.Context())
	require.NoError(t, err)

	err = svc.Close()
	require.NoError(t, err)
}
