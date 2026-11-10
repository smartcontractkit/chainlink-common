package beholder_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestNewOtelErrorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{name: "plain error", err: errors.New("boom")},
		{
			name: "wrapped export failure",
			err:  fmt.Errorf("failed to upload metrics: %w", errors.New("rpc error: code = Unavailable desc = data refused due to high memory usage")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lggr, logs := logger.TestObserved(t, zapcore.DebugLevel)

			beholder.NewOtelErrorHandler(lggr).Handle(tt.err)

			entries := logs.TakeAll()
			require.Len(t, entries, 1)
			assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
			assert.Equal(t, "Telemetry error", entries[0].Message)
			assert.Equal(t, tt.err.Error(), entries[0].ContextMap()["err"])
		})
	}
}
