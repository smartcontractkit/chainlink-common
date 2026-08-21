# OTel metric export batching

The OpenTelemetry SDK can split one metric collection into multiple exporter
calls by setting this environment variable before the node constructs its
Beholder meter provider:

```text
OTEL_GO_X_METRIC_EXPORT_BATCH_SIZE=<positive data-point count>
```

This is an experimental, process-wide SDK setting. It is read when each
`PeriodicReader` is constructed, so changing the environment afterward does
not reconfigure an existing reader. The value limits the number of metric
data points per exporter call; it is not a serialized-byte limit. Large
attributes, histograms, or a single oversized data point can still exceed a
collector receive limit.

An unset, invalid, zero, or negative value preserves the default unbatched
behavior. Configure this at deployment time; library code must not mutate the
process environment.
