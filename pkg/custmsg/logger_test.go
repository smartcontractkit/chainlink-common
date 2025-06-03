package custmsg

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
)

func TestBeholderLogger(t *testing.T) {
	emitter := newMockMessageEmitter(t)
	lggr := NewBeholderLogger(logger.Test(t), emitter)

	emitter.EXPECT().With("log_level", "debug").Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, "message").Return(nil).Once()
	lggr.Debug("message")

	emitter.EXPECT().With("log_level", "debug").Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, ">message<").Return(nil).Once()
	lggr.Debugf(">%s<", "message")

	emitter.EXPECT().With("log_level", "debug").Return(emitter).Once()
	emitter.EXPECT().WithMapLabels(map[string]string{"keyX": "valueX"}).Return(emitter).Once()
	emitter.EXPECT().Emit(matches.AnyContext, "message").Return(nil).Once()
	lggr.Debugw("message", "keyX", "valueX")
}
