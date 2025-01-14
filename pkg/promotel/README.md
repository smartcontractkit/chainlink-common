# Package Overview
The package provides components for performing Prometheus to OTel metrics conversion. 

Main components: MetricsReceiver, MetricsExporter 

## Receiver
- Wraps [prometheusreceiver](github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver)
- Fetches prometheus metrics data via `prometheus.Gatherer` (same process memory, no HTTP calls)
- Uses custom implementation of `prometheus.scraper` (from here https://github.com/pkcll/prometheus/pull/1) to shortcut HTTP request calls and fetch data from `prometheus.Gatherer`
- Converts Prometheus metrics into OTel format using [prometheusreceiver](github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver)
- Passes OTel metrics data to downstream OTel [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)

## Exporter
- Wraps [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)
- Receives metric data from the receiver
- Export OTel metrics data to otel collector endpoint via [otlpexporter](go.opentelemetry.io/collector/exporter/otlpexporter)

## OTel collector prometheusreceiver

[prometheusreceiver](github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver). 
is a component of otel-collector which collects metrics from Prometheus endpoints. It scrapes the metrics at regular intervals and converts them into a format that can be processed by the rest of the collector pipeline.

`promotel` is a wrapper around `prometheusreceiver` which provides a simple API to start and stop the receiver and process the metrics data.

`promotel` uses `prometheusreceiver` factory to create an instance of the receiver via `factory.CreateMetrics` with provided configuration. It also provides a callback function which is called every time new metrics data is received. The metrics data is a `pmetric.Metrics` object which contains the metrics data received from the Prometheus endpoint.

`promotel/inernal` contains implementations for `consumer.Metrics`, `component.Host`, `receiver.Settings`, `component.TelemetrySettings` which are dependencies required for `factory.CreateMetrics`.

`metrics.Consumer` is an interface which is used to process the metrics data. The `prometheusreceiver` calls `Consumer.ConsumeMetrics` function every time new metrics data is received. 

`prometheusreceiver` has Start and Shutdown methods. 

`github.com/pkcll/prometheus v0.54.1-promotel` fork overrides the `prometheus` package to provide a way to scrape metrics directly from `prometheus.DefaultGatherer` without making HTTP requests to the Prometheus endpoint. This is useful when the Prometheus endpoint is not accessible from the collector.

Example configuration:


```yaml
receivers:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'example'
          static_configs:
            - targets: ['localhost:9090']

```

## OTel collector otlpexporter

[otlpexporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) is a component of the OpenTelemetry Collector that exports telemetry data (metrics, logs, and traces) using the OpenTelemetry Protocol (OTLP). It supports both gRPC and HTTP transport protocols.

Example configuration:

```yaml
exporters:
	otlp:
		endpoint: "localhost:4317"
		tls:
			insecure: true
		retry_on_failure:
			enabled: true
			initial_interval: 5s
			max_interval: 30s
			max_elapsed_time: 300s
		sending_queue:
			enabled: true
			queue_size: 5000
```

### `promotel` usage example:

```go
import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel"
)

func main() {
	exporterConfig, _ := promotel.NewDefaultExporterConfig()
	exporter, _ := promotel.NewMetricExporter(exporterConfig, logger) 
	receiverConfig, _ := promotel.NewDefaultReceiverConfig()
    // Fetches metrics data directly from DefaultGatherer without making HTTP requests to 127.0.0.1:8888
	receiver, _ := promotel.NewMetricReceiver(receiverConfig, prometheus.DefaultGatherer, exporter.Consumer().ConsumeMetrics, logger)
	fmt.Println("Starting promotel pipeline")
	exporter.Start(context.Background())
	receiver.Start(context.Background())
	defer receiver.Close()
	defer exporter.Close()
	time.Sleep(1 * time.Minute)
}
```

### Debug Metric Receiver 

`DebugMetricReceiver` is an implementation of `metrics.Consumer` which prints formatted metrics data to stdout. It is useful for testing purposes.

### `Debug Metric Receiver` usage example:

```go
...
	// Debug metric receiver prints fetched metrics to stdout
	receiver, err := promotel.NewDebugMetricReceiver(config, prometheus.DefaultGatherer, logger)
	// Start metric receiver 
	receiver.Start(context.Background())
...
```

Output example

```
NumberDataPoints #0
StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2025-01-02 18:38:28.905 +0000 UTC
Value: 44.000000
Metric #18
Descriptor:
     -> Name: otelcol_exporter_sent_metric_points
     -> Description: Number of metric points successfully sent to destination.
     -> Unit: 
     -> DataType: Sum
     -> IsMonotonic: true
     -> AggregationTemporality: Cumulative
NumberDataPoints #0
Data point attributes:
     -> exporter: Str(debug)
     -> service_version: Str(0.108.1)
StartTimestamp: 2025-01-02 18:38:05.905 +0000 UTC
Timestamp: 2025-01-02 18:38:28.905 +0000 UTC
Value: 137.000000
NumberDataPoints #1
Data point attributes:
     -> exporter: Str(otlphttp)
     -> service_version: Str(0.108.1)
StartTimestamp: 2025-01-02 18:38:05.905 +0000 UTC
Timestamp: 2025-01-02 18:38:28.905 +0000 UTC
Value: 137.000000
Metric #19
Descriptor:
     -> Name: otelcol_process_cpu_seconds
     -> Description: Total CPU user and system time in seconds
     -> Unit: 
     -> DataType: Sum
     -> IsMonotonic: true
     -> AggregationTemporality: Cumulative
NumberDataPoints #0
Data point attributes:
     -> service_version: Str(0.108.1)
StartTimestamp: 2025-01-02 18:38:05.905 +0000 UTC
Timestamp: 2025-01-02 18:38:28.905 +0000 UTC
Value: 0.930000
```