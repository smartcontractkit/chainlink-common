package telemetry

import (
	"github.com/smartcontractkit/chainlink-relay/core/services/telemetry/generated"
	"github.com/smartcontractkit/libocr/commontypes"
)

type endpoint struct {
	client    Client
	namespace string
}

func MakeOCREndpoint(client Client, namespace string) commontypes.MonitoringEndpoint {
	return &endpoint{
		client,
		namespace,
	}
}

func (e *endpoint) SendLog(log []byte) {
	e.client.Send(&generated.TelemetryRequest{
		Telemetry: log,
		Address:   e.namespace,
	})
}
