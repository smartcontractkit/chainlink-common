package telemetry

import "github.com/smartcontractkit/libocr/commontypes"

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
	e.client.Send(&TelemetryRequest{
		Telemetry: log,
		Address:   e.namespace,
	})
}
