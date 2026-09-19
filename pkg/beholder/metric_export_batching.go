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
// reader. OTel v1.44 exposes this experimental setting only through an
// environment variable, which it reads synchronously during construction.
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
