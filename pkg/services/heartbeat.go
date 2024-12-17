package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// Heartbeat is a usage of Engine for application specific heartbeats,
// used in the core node and for loops. It accepts a named logger,
// a beat, a setup func to initialize resources used on each beat,
// a beat function to define the behavior on each beat, and a close func
// for resource teardown
type Heartbeat struct {
	Service
	eng *Engine

	beat time.Duration
	lggr logger.Logger
}

func NewHeartbeat(
	lggr logger.Logger,
	beat time.Duration,
	setupFn func(ctx context.Context) error,
	beatFn func(bCtx context.Context),
	closeFn func() error,
) Heartbeat {
	h := Heartbeat{
		beat: beat,
		lggr: lggr,
	}
	startFn := func(ctx context.Context) error {
		err := setupFn(ctx)
		if err != nil {
			return fmt.Errorf("setting up heartbeat: %w", err)
		}

		// consistent tick period
		constantTickFn := func() time.Duration {
			return h.beat
		}

		// TODO allow for override of tracer provider in engine
		// TODO wrap beatFn in engine trace
		h.eng.GoTick(timeutil.NewTicker(constantTickFn), beatFn)
		return nil
	}

	h.Service, h.eng = Config{
		Name:  fmt.Sprintf("%s.%s", lggr.Name(), "heartbeat"),
		Start: startFn,
		Close: closeFn,
	}.NewServiceEngine(lggr)
	return h
}
