package telemetry_test

const (
	chainID    = "chain-id"
	contractID = "contract-id"
	network    = "solana"
	telemType  = "mercury"
)

var (
	payload = []byte("oops")

	DefaultStaticTelemetryConfig = StaticTelemetryConfig{
		ChainID:    chainID,
		ContractID: contractID,
		Network:    network,
		Payload:    payload,
		TelemType:  telemType,
	}

	DefaultStaticTelemetry = StaticTelemetry{
		StaticTelemetryConfig: DefaultStaticTelemetryConfig,
	}
)
