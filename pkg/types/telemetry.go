package types

import (
	"context"

	"github.com/smartcontractkit/libocr/commontypes"
)

type Telemetry interface {
	Send(ctx context.Context, contractID string, telemetryType string, network string, chainID string, payload []byte) error
	GenMonitoringEndpoint(contractID string, telemType string, network string, chainID string) commontypes.MonitoringEndpoint
}

// MonitoringEndpointGenerator almost identical to synchronization.MonitoringEndpointGenerator except for the telemetry type being string after https://github.com/smartcontractkit/chainlink/pull/10623 gets merged
type MonitoringEndpointGenerator interface {
	GenMonitoringEndpoint(contractID string, telemType string, network string, chainID string) commontypes.MonitoringEndpoint
}
