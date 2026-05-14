package diskmonitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestNewDiskMonitor(t *testing.T) {
	dm, err := NewDiskMonitor(logger.Test(t), t.TempDir(), "tmp_disk_usage_bytes", time.Second)
	require.NoError(t, err)
	assert.NotNil(t, dm)
	assert.Equal(t, time.Second, dm.tickInterval)
}

type mockGauge struct {
	gotValue int64
}

func (m *mockGauge) Record(ctx context.Context, value int64, options ...metric.RecordOption) {
	m.gotValue = value
}

func TestDiskMonitor_emitDirSizeMetric(t *testing.T) {
	lggr, observed := logger.TestObserved(t, zapcore.DebugLevel)

	dm := &DiskMonitor{
		sizeOfDir: func() (int64, error) {
			return 42, nil
		},
		gauge: &mockGauge{},
		lggr:  lggr,
	}

	dm.emitDirSizeMetric(t.Context())
	assert.Equal(t, int64(42), dm.gauge.(*mockGauge).gotValue)

	assert.Len(t, observed.FilterMessage("Emitting directory size metric").All(), 1)
}

func TestDiskMonitor_emitDirSizeMetric_error(t *testing.T) {
	lggr, observed := logger.TestObserved(t, zapcore.DebugLevel)

	dm := &DiskMonitor{
		sizeOfDir: func() (int64, error) {
			return 0, errors.New("disk read error")
		},
		gauge: &mockGauge{},
		lggr:  lggr,
	}

	dm.emitDirSizeMetric(t.Context())
	assert.Equal(t, int64(0), dm.gauge.(*mockGauge).gotValue)

	assert.Len(t, observed.FilterMessage("Failed to measure directory size").All(), 1)
}
