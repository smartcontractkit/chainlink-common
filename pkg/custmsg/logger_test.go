package custmsg

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
)

func TestBeholderLogger(t *testing.T) {
	emitter := newMockMessageEmitter(t)
	lggr := NewBeholderLogger(logger.Test(t), emitter)

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
}

func TestBeholderLogger_Blocking(t *testing.T) {
	emitter := newMockMessageEmitter(t)
	lggr := NewBeholderLogger(logger.Test(t), emitter)

	emitter.EXPECT().With(LogLevelKey, "debug").Return(emitter).Once()
	// simulate a blocking call
	emitter.EXPECT().Emit(matches.AnyContext, "message").RunAndReturn(func(ctx context.Context, msg string) error {
		<-ctx.Done()
		return ctx.Err()
	}).Once()
	// should eventually cancel the context
	lggr.Debug("message")
}
