package beholder_test

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func ExampleConfig() {
	config := beholder.Config{
		OtelExporterGRPCEndpoint: "localhost:4317",
		PackageName:              "beholder",
	}
	fmt.Printf("%+v", config)
	// Output:
	// {OtelExporterGRPCEndpoint:localhost:4317 PackageName:beholder MessageEmitterRetryCount:0 MessageEmitterRetryDelay:0s}
}
