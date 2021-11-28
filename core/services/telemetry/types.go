package telemetry

import "github.com/smartcontractkit/chainlink-relay/core/services/telemetry/generated"

type Client interface {
	Send(*generated.TelemetryRequest)
}

type Service interface {
	Start() (Client, error)
	Stop()
}
