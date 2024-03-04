package loop_test

import (
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
)

func TestHCLogLogger(t *testing.T) {
	lggr, ol := logger.TestObserved(t, zapcore.ErrorLevel)
	loggerTest := &test.GRPCPluginLoggerTest{Logger: lggr}
	cc := loggerTest.ClientConfig()
	cc.Cmd = NewHelperProcessCommand(test.PluginLoggerTestName, false)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	_, err := c.Client()
	require.Error(t, err)

	// Some logs should come through with plugin-side names
	require.NotEmpty(t, ol.Filter(func(entry observer.LoggedEntry) bool {
		return entry.LoggerName == test.LoggerTestName
	}), ol.All())
}

func TestHCLogLoggerPanic(t *testing.T) {
	lggr, ol := logger.TestObserved(t, zapcore.ErrorLevel)
	loggerTest := &test.GRPCPluginLoggerTestPanic{Logger: lggr}
	cc := loggerTest.ClientConfig()
	cc.Cmd = NewHelperProcessCommand(test.PluginLoggerTestPanicName, false)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	_, err := c.Client()
	require.NoError(t, err)
	time.Sleep(time.Second * 2) //wait for panic

	entrys := ol.All()
	require.Len(t, entrys, 2)
	require.Equal(t, entrys[0].Message, "panic: test panic")
	require.Equal(t, entrys[0].Level, zapcore.ErrorLevel)

}
