package beholder_test

import (
	"fmt"
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	packageName = "beholder"
)

func ExampleConfig() {
	config := beholder.Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		OtelExporterHTTPEndpoint: "localhost:4318",
		// Resource
		ResourceAttributes: []otelattr.KeyValue{
			otelattr.String("package_name", packageName),
			otelattr.String("sender", "beholderclient"),
		},
		// Message Emitter
		EmitterExportTimeout:  1 * time.Second,
		EmitterBatchProcessor: true,
		// OTel message log exporter retry config
		LogRetryConfig: nil,
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		// OTel trace exporter retry config
		TraceRetryConfig: nil,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// OTel metric exporter retry config
		MetricRetryConfig: nil,
		// Log
		LogExportTimeout:  1 * time.Second,
		LogBatchProcessor: true,
	}
	fmt.Printf("%+v\n", config)
	config.LogRetryConfig = &beholder.RetryConfig{
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  1 * time.Minute, // Set to zero to disable retry
	}
	fmt.Printf("%+v\n", *config.LogRetryConfig)
	// Output:
	// {InsecureConnection:true CACertFile: OtelExporterGRPCEndpoint:localhost:4317 OtelExporterHTTPEndpoint:localhost:4318 ResourceAttributes:[{Key:package_name Value:{vtype:4 numeric:0 stringly:beholder slice:<nil>}} {Key:sender Value:{vtype:4 numeric:0 stringly:beholderclient slice:<nil>}}] EmitterExportTimeout:1s EmitterBatchProcessor:true LogRetryConfig:<nil> TraceSampleRatio:1 TraceBatchTimeout:1s TraceSpanExporter:<nil> TraceRetryConfig:<nil> MetricReaderInterval:1s MetricRetryConfig:<nil> LogExportTimeout:1s LogBatchProcessor:true}
	// {InitialInterval:5s MaxInterval:30s MaxElapsedTime:1m0s}
}
