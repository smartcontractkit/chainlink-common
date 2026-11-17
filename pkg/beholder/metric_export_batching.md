# OTel metric export batching

The OpenTelemetry SDK can split one metric collection into multiple exporter
calls through the Chainlink Core configuration:

```text
[Telemetry]
MetricExportBatchSize = <positive data-point count>
```

This is backed by the OpenTelemetry SDK `WithMaxExportBatchSize` reader option
(sdk/metric v1.47+). The Beholder client applies it when constructing each
`PeriodicReader`; the value limits metric data points per exporter call and is
not a serialized-byte limit. Large attributes, histograms, or a single
oversized data point can still exceed a collector receive limit.

A zero or negative setting preserves the default unbatched behavior. Configure
the Core TOML before node startup.

LOOP plugin subprocesses do not read `beholder.Config` directly; Core forwards
its resolved value to each plugin via the `CL_TELEMETRY_METRIC_EXPORT_BATCH_SIZE`
environment variable (see `pkg/loop/config.go`).

Earlier SDK releases (through v1.46) exposed this setting only via the
experimental `OTEL_GO_X_METRIC_EXPORT_BATCH_SIZE` environment variable; that
variable was removed in v1.47 and is no longer read.
