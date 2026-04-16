package utils_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

type chanWorker struct {
	ch    chan struct{}
	delay time.Duration
}

func (t *chanWorker) Name() string {
	return "ChanWorker"
}

func (t *chanWorker) Work() {
	if t.delay != 0 {
		time.Sleep(t.delay)
	}
	t.ch <- struct{}{}
}

func TestSleeperTask_WakeupAfterStopPanics(t *testing.T) {
	t.Parallel()

	worker := &chanWorker{ch: make(chan struct{}, 1)}
	sleeper := utils.NewSleeperTask(worker)

	require.NoError(t, sleeper.Stop())

	require.Panics(t, func() {
		sleeper.WakeUp()
	})

	select {
	case <-worker.ch:
		t.Fatal("work was performed when none was expected")
	default:
	}
}

func TestSleeperTask_CallingStopTwiceFails(t *testing.T) {
	t.Parallel()

	worker := &chanWorker{}
	sleeper := utils.NewSleeperTask(worker)
	require.NoError(t, sleeper.Stop())
	require.Error(t, sleeper.Stop())
}

func TestSleeperTask_WakeupPerformsWork(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	worker := &chanWorker{ch: make(chan struct{}, 1)}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()

	select {
	case <-worker.ch:
	case <-ctx.Done():
		t.Error("timed out waiting for work to be performed")
	}

	require.NoError(t, sleeper.Stop())
}

// reflectWorker simulates the race scenario: Work() spawns goroutines that
// read the context via fmt.Sprintf (like testify's mock.callString), while
// Stop() concurrently triggers context cancellation.
type reflectWorker struct {
	wg sync.WaitGroup
}

func (w *reflectWorker) Name() string { return "ReflectWorker" }

func (w *reflectWorker) Work(ctx context.Context) {
	w.wg.Add(10)
	for range 10 {
		go func() {
			defer w.wg.Done()
			// Simulate testify's callString: fmt.Sprintf("%#v", ctx) reads context
			// internals via reflect. With the old NewCtx()/CtxCancel pattern, cancel()
			// fired concurrently, writing to context internals → DATA RACE.
			_ = fmt.Sprintf("%#v", ctx)
		}()
	}
	w.wg.Wait()
}

func TestSleeperTask_NoConcurrentContextRace(t *testing.T) {
	t.Parallel()

	// Run many iterations to increase chance of triggering a race.
	for range 50 {
		w := &reflectWorker{}
		sleeper := utils.NewSleeperTaskCtx(w)
		sleeper.WakeUp()
		// Stop concurrently with Work — this closes chStop, which previously
		// fired cancel() via a CtxCancel goroutine, racing with reflect reads.
		require.NoError(t, sleeper.Stop())
	}
}

type controllableWorker struct {
	chanWorker
	awaitWorkStarted chan struct{}
	allowResumeWork  chan struct{}
	ignoreSignals    bool
}

func (w *controllableWorker) Work() {
	if !w.ignoreSignals {
		w.awaitWorkStarted <- struct{}{}
		<-w.allowResumeWork
	}
	w.chanWorker.Work()
}

func TestSleeperTask_WakeupEnqueuesMaxTwice(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	worker := &controllableWorker{chanWorker: chanWorker{ch: make(chan struct{}, 1)}, awaitWorkStarted: make(chan struct{}), allowResumeWork: make(chan struct{})}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()
	<-worker.awaitWorkStarted
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	worker.ignoreSignals = true
	worker.allowResumeWork <- struct{}{}

	for range 2 {
		select {
		case <-worker.ch:
		case <-ctx.Done():
			t.Error("timed out waiting for work to be performed")
		}
	}

	if !t.Failed() {
		select {
		case <-worker.ch:
			t.Errorf("unexpected work performed")
		case <-time.After(time.Second):
		}
	}

	require.NoError(t, sleeper.Stop())
}

func TestSleeperTask_StopWaitsUntilWorkFinishes(t *testing.T) {
	t.Parallel()

	worker := &controllableWorker{chanWorker: chanWorker{ch: make(chan struct{}, 1)}, awaitWorkStarted: make(chan struct{}), allowResumeWork: make(chan struct{})}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()
	<-worker.awaitWorkStarted

	select {
	case <-worker.ch:
		t.Error("work was performed when none was expected")
		assert.NoError(t, sleeper.Stop())
		return
	default:
	}

	worker.allowResumeWork <- struct{}{}

	require.NoError(t, sleeper.Stop())

	select {
	case <-worker.ch:
	default:
		t.Fatal("work should have been performed")
	}

	select {
	case <-worker.ch:
		t.Fatal("extra work was performed")
	default:
	}
}
