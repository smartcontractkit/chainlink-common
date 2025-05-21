package core

import (
	"context"
	"net/url"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/gateway"
)

// GatewayConnector is a component run by Nodes to connect to a set of Gateways.
type GatewayConnector interface {
	services.Service
	// TODO: check if this can be removed
	ConnectionInitiator

	AddHandler(methods []string, handler GatewayConnectorHandler) error
	// SendToGateway takes a signed message as argument and sends it to the specified gateway
	SendToGateway(ctx context.Context, gatewayID string, msg *gateway.Message) error
	// SignAndSendToGateway signs the message and sends the message to the specified gateway
	SignAndSendToGateway(ctx context.Context, gatewayID string, msg *gateway.MessageBody) error
	// GatewayIDs returns the list of Gateway IDs
	GatewayIDs() []string
	// DonID returns the DON ID
	DonID() string
	AwaitConnection(ctx context.Context, gatewayID string) error
}

// GatewayConnector user (node) implements application logic in the Handler interface.
type GatewayConnectorHandler interface {
	Start(context.Context) error
	Close() error
	HandleGatewayMessage(ctx context.Context, gatewayID string, msg *gateway.Message)
}

// The handshake works as follows:
//
//	 Client (Initiator)                  Server (Acceptor)
//
//	NewAuthHeader()
//	            -------auth header-------->
//	                                      StartHandshake()
//	            <-------challenge----------
//
// ChallengeResponse()
//
//	---------response--------->
//	                        FinalizeHandshake()
type ConnectionInitiator interface {
	// Generate authentication header value specific to node and gateway
	NewAuthHeader(url *url.URL) ([]byte, error)

	// Sign challenge to prove identity.
	ChallengeResponse(url *url.URL, challenge []byte) ([]byte, error)
}
