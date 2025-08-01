package custmsg

import (
	"context"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
)

func TestBeholderLogger(t *testing.T) {
	emitter := newMockMessageEmitter(t)
	taskProc := &taskProcessor{
		emitQueue: make(chan func(context.Context), 10),
	}
	loopDone := make(chan struct{})
	go emitLoop(t, taskProc.emitQueue, loopDone)
	lggr := NewBeholderLogger(logger.Test(t), emitter, taskProc)

	emitter.EXPECT().With(LogLevelKey, "debug").Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, "message").Return(nil).Once()
	lggr.Debug("message")

	emitter.EXPECT().With(LogLevelKey, "debug").Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, ">message<").Return(nil).Once()
	lggr.Debugf(">%s<", "message")

	emitter.EXPECT().With(LogLevelKey, "debug").Return(emitter).Once()
	emitter.EXPECT().WithMapLabels(map[string]string{"keyX": "valueX"}).Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, "message").Return(nil).Once()
	lggr.Debugw("message", "keyX", "valueX")

	emitter.EXPECT().WithMapLabels(map[string]string{"keyY": "valueY"}).Return(emitter).Once()
	_ = lggr.With("keyY", "valueY")
	close(taskProc.emitQueue)
	<-loopDone
}

func TestBeholderLogger_Blocking(t *testing.T) {
	emitter := newMockMessageEmitter(t)
	taskProc := &taskProcessor{
		emitQueue: make(chan func(context.Context), 10),
	}
	loopDone := make(chan struct{})
	go emitLoop(t, taskProc.emitQueue, loopDone)
	lggr := NewBeholderLogger(logger.Test(t), emitter, taskProc)

	emitter.EXPECT().With(LogLevelKey, "debug").Return(emitter).Once()
	// simulate a blocking call
	emitter.EXPECT().Emit(matches.AnyContext, "message").RunAndReturn(func(ctx context.Context, msg string) error {
		<-ctx.Done()
		return ctx.Err()
	}).Once()
	// should eventually cancel the context
	lggr.Debug("message")
	close(taskProc.emitQueue)
	<-loopDone
}

type taskProcessor struct {
	emitQueue chan func(context.Context)
}

func (tp *taskProcessor) Enqueue(fn func(context.Context)) {
	tp.emitQueue <- fn
}

func emitLoop(t *testing.T, emitQueue chan func(context.Context), done chan struct{}) {
	for fn := range emitQueue {
		// block for 100ms to simulate a blocking call
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()
		fn(ctx)
	}
	close(done)
}
