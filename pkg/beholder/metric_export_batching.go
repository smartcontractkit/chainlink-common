package beholder

import (
	"os"
	"strconv"
	"sync"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const metricExportBatchSizeEnv = "OTEL_GO_X_METRIC_EXPORT_BATCH_SIZE"

var metricExportBatchSizeMu sync.Mutex

// newPeriodicReader applies the configured batch size while constructing a
// reader. OTel sdk/metric (through v1.46) exposes this experimental setting
// only through an environment variable, which it reads synchronously during
// construction. All PeriodicReader construction in this package must go
// through this function so the process-wide variable is serialized.
//
// TODO: when sdk/metric is bumped to a release with WithMaxExportBatchSize
// (v1.47+), which removes the environment variable, switch to that option and
// drop the environment handling here.
func newPeriodicReader(exporter sdkmetric.Exporter, batchSize int, opts ...sdkmetric.PeriodicReaderOption) *sdkmetric.PeriodicReader {
	metricExportBatchSizeMu.Lock()
	defer metricExportBatchSizeMu.Unlock()

	previous, wasSet := os.LookupEnv(metricExportBatchSizeEnv)
	if batchSize > 0 {
		_ = os.Setenv(metricExportBatchSizeEnv, strconv.Itoa(batchSize))
	} else {
		_ = os.Unsetenv(metricExportBatchSizeEnv)
	}
	defer func() {
		if wasSet {
			_ = os.Setenv(metricExportBatchSizeEnv, previous)
			return
		}
		_ = os.Unsetenv(metricExportBatchSizeEnv)
	}()

	return sdkmetric.NewPeriodicReader(exporter, opts...)
}
