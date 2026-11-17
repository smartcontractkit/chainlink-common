package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// newPeriodicReader applies the configured batch size while constructing a
// reader using the WithMaxExportBatchSize option (sdk/metric v1.47+).
// A zero or negative batch size preserves the default unbatched behavior.
// All PeriodicReader construction in this package must go through this
// function so the batch size is applied consistently.
func newPeriodicReader(exporter sdkmetric.Exporter, batchSize int, opts ...sdkmetric.PeriodicReaderOption) *sdkmetric.PeriodicReader {
	if batchSize > 0 {
		opts = append(opts, sdkmetric.WithMaxExportBatchSize(batchSize))
	}
	return sdkmetric.NewPeriodicReader(exporter, opts...)
}
