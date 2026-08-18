package diskmonitor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestDiskMonitor_emitDirSizeMetric_realDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), []byte("xy"), 0o644))
	sub := filepath.Join(root, "sub")
	require.NoError(t, os.Mkdir(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "c.txt"), []byte("z"), 0o644))

	dm := &DiskMonitor{
		sizeOfDir: func() (int64, error) {
			return totalRegularFileSizeBytes(root)
		},
		gauge: &mockGauge{},
		lggr:  logger.Test(t),
	}

	dm.emitDirSizeMetric(t.Context())
	assert.Equal(t, int64(8), dm.gauge.(*mockGauge).gotValue)
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

func TestDiskMonitor_emitDirSizeMetric_prometheusGauge(t *testing.T) {
	dm := &DiskMonitor{
		sizeOfDir: func() (int64, error) {
			return 99, nil
		},
		gauge:     &mockGauge{},
		promGauge: &mockPromGauge{},
		lggr:      logger.Test(t),
	}

	dm.emitDirSizeMetric(t.Context())
	assert.Equal(t, int64(99), dm.gauge.(*mockGauge).gotValue)
	assert.Equal(t, float64(99), dm.promGauge.(*mockPromGauge).gotValue)
}

type mockPromGauge struct {
	gotValue float64
}

func (m *mockPromGauge) Set(v float64) {
	m.gotValue = v
}
