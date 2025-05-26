package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/gateway"
)

// GatewayConnector is a component run by Nodes to connect to a set of Gateways.
type GatewayConnector interface {
	// Start starts the GatewayConnector
	Start(context.Context) error
	// Close closes the GatewayConnector
	Close() error
	// AddHandler adds a handler to the GatewayConnector
	AddHandler(methods []string, handler GatewayConnectorHandler) error
	// SendToGateway takes a signed message as argument and sends it to the specified gateway
	SendToGateway(ctx context.Context, gatewayID string, msg *gateway.Message) error
	// SignAndSendToGateway signs the message and sends the message to the specified gateway
	SignAndSendToGateway(ctx context.Context, gatewayID string, msg *gateway.MessageBody) error
	// GatewayIDs returns the list of Gateway IDs
	GatewayIDs() ([]string, error)
	// DonID returns the DON ID
	DonID() (string, error)
	AwaitConnection(ctx context.Context, gatewayID string) error
}

// GatewayConnector user (node) implements application logic in the Handler interface.
type GatewayConnectorHandler interface {
	// Start starts the GatewayConnectorHandler
	Start(context.Context) error
	// Close closes the GatewayConnectorHandler
	Close() error
	// TODO: revisit interface
	Info() (GatewayConnectorHandlerInfo, error)
	HandleGatewayMessage(ctx context.Context, gatewayID string, msg *gateway.Message) error
}

type GatewayConnectorHandlerInfo struct {
	// ID is the unique identifier of the handler
	ID string
}
