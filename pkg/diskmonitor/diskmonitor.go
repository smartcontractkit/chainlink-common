package diskmonitor

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type int64Gauge interface {
	Record(ctx context.Context, value int64, options ...metric.RecordOption)
}

// totalRegularFileSizeBytes sums on-disk file sizes for every non-directory under dirPath (recursive), using filepath.WalkDir.
func totalRegularFileSizeBytes(dirPath string) (int64, error) {
	var totalSize int64
	walkErr := filepath.WalkDir(dirPath, func(_ string, d fs.DirEntry, ierr error) error {
		if ierr != nil || d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		totalSize += fi.Size()
		return nil
	})
	return totalSize, walkErr
}

// DiskMonitor measures dirPath on a fixed interval and records total file bytes to gaugeName.
type DiskMonitor struct {
	services.Service

	eng          *services.Engine
	tickInterval time.Duration
	lggr         logger.Logger
	sizeOfDir    func() (int64, error)
	gauge        int64Gauge
}

// NewDiskMonitor returns a [DiskMonitor] for dirPath using gaugeName at tickInterval.
func NewDiskMonitor(lggr logger.Logger, dirPath string, gaugeName string, tickInterval time.Duration) (*DiskMonitor, error) {
	g, err := beholder.GetMeter().Int64Gauge(gaugeName)
	if err != nil {
		return nil, fmt.Errorf("int64 gauge %q: %w", gaugeName, err)
	}

	dm := &DiskMonitor{
		gauge:        g,
		tickInterval: tickInterval,
		sizeOfDir: func() (int64, error) {
			return totalRegularFileSizeBytes(dirPath)
		},
	}

	dm.Service, dm.eng = services.Config{
		Name:  "DiskMonitor",
		Start: dm.start,
	}.NewServiceEngine(logger.With(
		lggr,
		"dirPath", dirPath,
		"gaugeName", gaugeName,
	))
	dm.lggr = dm.eng.SugaredLogger
	return dm, nil
}

func (dm *DiskMonitor) start(ctx context.Context) error {
	ticker := services.TickerConfig{}.NewTicker(dm.tickInterval)
	dm.eng.GoTick(ticker, dm.emitDirSizeMetric)
	return nil
}

func (dm *DiskMonitor) emitDirSizeMetric(ctx context.Context) {
	totalSize, err := dm.sizeOfDir()
	if err != nil {
		dm.lggr.Errorw("Failed to measure directory size", "error", err)
		return
	}

	dm.lggr.Debugw("Emitting directory size metric", "sizeBytes", totalSize)
	dm.gauge.Record(ctx, totalSize)
}
