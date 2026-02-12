package core

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/services/orgresolver"
)

// StandardCapabilitiesDependencies contains all the dependencies injected for capability initialization.
// We use a struct to evolve the interface without requiring updates to all implementors.
// i.e. Initialise(ctx context.Context, dependencies core.StandardCapabilitiesDependencies) error
type StandardCapabilitiesDependencies struct {
	Config             string
	TelemetryService   TelemetryService
	Store              KeyValueStore
	CapabilityRegistry CapabilitiesRegistry
	ErrorLog           ErrorLog
	PipelineRunner     PipelineRunnerService
	RelayerSet         RelayerSet
	OracleFactory      OracleFactory
	GatewayConnector   GatewayConnector
	P2PKeystore        Keystore
	OrgResolver        orgresolver.OrgResolver
	CRESettings        SettingsBroadcaster
	TriggerEventStore  capabilities.EventStore
}
