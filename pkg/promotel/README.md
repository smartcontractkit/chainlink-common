# Package Overview
The package provides components for performing Prometheus to OTel metrics conversion. 

Main components: MetricsReceiver, MetricsExporter 

## Receiver
- Wraps [prometheusreceiver](github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver)
- Fetches prometheus metrics data via `prometheus.Gatherer` (same process memory, no HTTP calls)
- Uses custom implementation of `prometheus.scraper` (from here https://github.com/pkcll/prometheus/pull/1) to shortcut HTTP request calls and fetch data from `prometheus.Gatherer`
- Converts Prometheus metrics into OTel format using [prometheusreceiver](github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver)
- Passes OTel metrics data to downstream OTel [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)

## Exporter
- Wraps [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)
- Receives metric data from the receiver
- Export OTel metrics data to otel collector endpoint via [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)



### Usage

```go
    ...
	forwarder, err := promotel.NewForwarder(g, r, lggr, promotel.ForwarderOptions{
		Endpoint:    srv.URL,
		TLSInsecure: true,
		Interval:    interval,
	})
	err = forwarder.Start(ctx)
	defer forwarder.Close()
	...
```
