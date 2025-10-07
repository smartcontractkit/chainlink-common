package core

// StandardCapabilitiesDependencies contains all the dependencies injected for capability initialization.
// We use a struct to evolve the interface without requiring updates to all implementors.
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
}
