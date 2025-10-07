package core

// StandardCapabilitiesServices contains all the services required for capability initialization.
// We use a struct to evolve the interface without requiring updates to all implementors.
type StandardCapabilitiesServices struct {
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
