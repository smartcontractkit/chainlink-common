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
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// Log
		LogExportTimeout:  1 * time.Second,
		LogBatchProcessor: true,
	}
	fmt.Printf("%+v", config)
	// Output:
	// {InsecureConnection:true CACertFile: OtelExporterGRPCEndpoint:localhost:4317 OtelExporterHTTPEndpoint:localhost:4318 ResourceAttributes:[{Key:package_name Value:{vtype:4 numeric:0 stringly:beholder slice:<nil>}} {Key:sender Value:{vtype:4 numeric:0 stringly:beholderclient slice:<nil>}}] EmitterExportTimeout:1s EmitterBatchProcessor:true TraceSampleRatio:1 TraceBatchTimeout:1s TraceSpanExporter:<nil> MetricReaderInterval:1s LogExportTimeout:1s LogBatchProcessor:true AuthenticatorPublicKey:[] AuthenticatorSigner:<nil> AuthenticatorHeaders:map[]}
}
