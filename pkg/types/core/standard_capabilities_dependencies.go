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
	// CapabilityDonID is the on-chain DON ID of the capability DON this plugin
	// process was spawned for, resolved authoritatively by the host before
	// Initialise is called. Plugins should use this as the source of truth for
	// their own DON identity (e.g. when emitting events that need to carry the
	// *sending* DON ID, distinct from the consumer workflow's DON ID).
	//
	// Zero means the host did not provide one — either a legacy core node that
	// pre-dates this field, or a boot path that has not yet been updated to
	// populate it. Plugins SHOULD fall back to resolving via the capability
	// registry in that case, but the fallback path cannot disambiguate when
	// the local node belongs to multiple DONs running the same capability.
	CapabilityDonID uint32
}
