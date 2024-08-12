package beholder_test

import (
	"fmt"
	"time"

	beholder "github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	packageName = "beholder"
)

func ExampleConfig() {
	config := beholder.Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		PackageName:              packageName,
		// Resource
		ResourceAttributes: map[string]string{
			"package_name": packageName,
			"sender":       "beholdeclient",
		},
		// Message Emitter
		EmitterExportTimeout: 1 * time.Second,
		// Trace
		TraceSampleRate:   1,
		TraceBatchTimeout: 1 * time.Second,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// Log
		LogExportTimeout: 1 * time.Second,
	}
	fmt.Printf("%+v", config)
	// Output:
	// {InsecureConnection:true CACertFile: OtelExporterGRPCEndpoint:localhost:4317 PackageName:beholder ResourceAttributes:map[package_name:beholder sender:beholdeclient] EmitterExportTimeout:1s TraceSampleRate:1 TraceBatchTimeout:1s MetricReaderInterval:1s LogExportTimeout:1s}
}
