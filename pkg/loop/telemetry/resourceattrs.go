package telemetry

import (
	"os"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
)

// Process resource attribute keys a host can enable with TelemetryConfig.TelemetryOTelResourceAttributes.
const (
	// AttrProcessExecutableName is build.Program: the main package path of this binary, so the core
	// node and each LOOP binary report different values.
	AttrProcessExecutableName = string(semconv.ProcessExecutableNameKey)
	// AttrServiceVersion is build.Version of this binary.
	AttrServiceVersion = string(semconv.ServiceVersionKey)
	// AttrProcessPID is the OS process ID. It changes on every restart, so every restart starts a new
	// set of series; enable it only where that churn is acceptable.
	AttrProcessPID = string(semconv.ProcessPIDKey)
)

// SupportedOTelResourceAttributes lists the keys ProcessResourceAttributes understands.
var SupportedOTelResourceAttributes = []string{AttrProcessExecutableName, AttrServiceVersion, AttrProcessPID}

// ProcessResourceAttributes returns the resource attributes for keys, in order, and the keys it does not
// support. Duplicate keys are added once.
func ProcessResourceAttributes(keys []string) (attrs []attribute.KeyValue, unknown []string) {
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		if seen[k] {
			continue
		}
		seen[k] = true
		switch k {
		case AttrProcessExecutableName:
			attrs = append(attrs, semconv.ProcessExecutableNameKey.String(build.Program))
		case AttrServiceVersion:
			attrs = append(attrs, semconv.ServiceVersionKey.String(build.Version))
		case AttrProcessPID:
			attrs = append(attrs, semconv.ProcessPIDKey.Int(os.Getpid()))
		default:
			unknown = append(unknown, k)
		}
	}
	return attrs, unknown
}
