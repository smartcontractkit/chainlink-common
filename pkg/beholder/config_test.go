package beholder_test

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	packageName = "beholder"
)

//go:embed testdata/config-example.json
var configExample string

var update = flag.Bool("update", false, "update golden test files")

func TestConfig(t *testing.T) {
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
		// true uses batched async export for custom messages.
		EmitterBatchProcessor: true,
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

	b, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	if *update {
		require.NoError(t, os.WriteFile("testdata/config-example.json", b, 0644))
	} else {
		assert.Equal(t, configExample, string(b))
	}

	config.LogRetryConfig = &beholder.RetryConfig{
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  1 * time.Minute, // Set to zero to disable retry
	}
	assert.Equal(t, "{InitialInterval:5s MaxInterval:30s MaxElapsedTime:1m0s}", fmt.Sprintf("%+v", *config.LogRetryConfig))
}
