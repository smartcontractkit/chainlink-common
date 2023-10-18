package types

import (
	"context"

	"github.com/smartcontractkit/libocr/commontypes"
)

type Telemetry interface {
	Send(ctx context.Context, contractID string, telemetryType string, network string, chainID string, payload []byte) error
}

// MonitoringEndpointGenerator almost identical to synchronization.MonitoringEndpointGenerator except for the telemetry type
type MonitoringEndpointGenerator interface {
	GenMonitoringEndpoint(contractID string, telemType string, network string, chainID string) commontypes.MonitoringEndpoint
}
