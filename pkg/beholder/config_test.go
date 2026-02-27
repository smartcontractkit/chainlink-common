package beholder_test

import (
	"fmt"
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"

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
		EmitterExportTimeout:      1 * time.Second,
		EmitterExportMaxBatchSize: 512,
		EmitterExportInterval:     1 * time.Second,
		EmitterMaxQueueSize:       2048,
		EmitterBatchProcessor:     true,
		// OTel message log exporter retry config
		LogRetryConfig: nil,
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		TraceCompressor:   "gzip",
		// OTel trace exporter retry config
		TraceRetryConfig: nil,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		MetricCompressor:     "gzip",
		// OTel metric exporter retry config
		MetricRetryConfig: nil,
		// Log
		LogExportTimeout:      1 * time.Second,
		LogExportMaxBatchSize: 512,
		LogExportInterval:     1 * time.Second,
		LogMaxQueueSize:       2048,
		LogBatchProcessor:     true,
		LogStreamingEnabled:   false,             // Disable streaming logs by default
		LogLevel:              zapcore.InfoLevel, // Default log level
		LogCompressor:         "gzip",
		// Auth
		AuthPublicKeyHex: "",
		AuthHeaders:      map[string]string{},
		AuthKeySigner:    nil,
		AuthHeadersTTL:   0,
	}
	fmt.Printf("%+v\n", config)
	config.LogRetryConfig = &beholder.RetryConfig{
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  1 * time.Minute, // Set to zero to disable retry
	}
	fmt.Printf("%+v\n", *config.LogRetryConfig)
	// Output:
	// {InsecureConnection:true CACertFile: OtelExporterGRPCEndpoint:localhost:4317 OtelExporterHTTPEndpoint:localhost:4318 ResourceAttributes:[{Key:package_name Value:{vtype:4 numeric:0 stringly:beholder slice:<nil>}} {Key:sender Value:{vtype:4 numeric:0 stringly:beholderclient slice:<nil>}}] EmitterExportTimeout:1s EmitterExportInterval:1s EmitterExportMaxBatchSize:512 EmitterMaxQueueSize:2048 EmitterBatchProcessor:true TraceSampleRatio:1 TraceBatchTimeout:1s TraceSpanExporter:<nil> TraceRetryConfig:<nil> TraceCompressor:gzip MetricReaderInterval:1s MetricRetryConfig:<nil> MetricViews:[] MetricCompressor:gzip ChipIngressEmitterEnabled:false ChipIngressEmitterGRPCEndpoint: ChipIngressInsecureConnection:false ChipIngressBufferSize:0 ChipIngressMaxBatchSize:0 ChipIngressSendInterval:0s ChipIngressSendTimeout:0s LogExportTimeout:1s LogExportInterval:1s LogExportMaxBatchSize:512 LogMaxQueueSize:2048 LogBatchProcessor:true LogRetryConfig:<nil> LogStreamingEnabled:false LogLevel:info LogCompressor:gzip AuthHeaders:map[] AuthHeadersTTL:0s AuthKeySigner:<nil> AuthPublicKeyHex:}
	// {InitialInterval:5s MaxInterval:30s MaxElapsedTime:1m0s}
}
