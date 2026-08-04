package host

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// A malicious or buggy guest can ask the host to log at "panic"/"fatal" level. The host
// must never let the guest crash or exit the host process, so these levels are downgraded
// to error level instead of calling logger.Panicw/Fatalw.
func TestLogRawMessage_PanicAndFatalLevelsDoNotCrashHost(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"panic", "fatal"} {
		t.Run(level, func(t *testing.T) {
			t.Parallel()

			lggr, logs := logger.TestObserved(t, zapcore.DebugLevel)
			require.NotPanics(t, func() {
				require.NoError(t, logRawMessage(lggr, fmt.Appendf(nil, `{"level":%q,"msg":"boom"}`, level)))
			})

			entries := logs.All()
			require.Len(t, entries, 1)
			require.Equal(t, zapcore.ErrorLevel, entries[0].Level)
			require.Equal(t, "boom", entries[0].Message)
		})
	}
}
