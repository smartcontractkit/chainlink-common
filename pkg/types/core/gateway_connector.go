package core

import (
	"context"
	"encoding/json"
	"errors"

	jsonrpc "github.com/smartcontractkit/chainlink-common/pkg/jsonrpc2"
)

// GatewayConnector is a component run by Nodes to connect to a set of Gateways.
type GatewayConnector interface {
	// AddHandler adds a handler to the GatewayConnector
	AddHandler(ctx context.Context, methods []string, handler GatewayConnectorHandler) error
	RemoveHandler(ctx context.Context, methods []string) error
	// SendToGateway takes a signed message as argument and sends it to the specified gateway
	SendToGateway(ctx context.Context, gatewayID string, resp *jsonrpc.Response[json.RawMessage]) error
	// Sign the given message and return signature
	SignMessage(ctx context.Context, msg []byte) ([]byte, error)
	// GatewayIDs returns the list of Gateway IDs
	GatewayIDs(ctx context.Context) ([]string, error)
	// DonID returns the DON ID
	DonID(ctx context.Context) (string, error)
	AwaitConnection(ctx context.Context, gatewayID string) error
}

// GatewayConnector user (node) implements application logic in the Handler interface.
type GatewayConnectorHandler interface {
	// ID returns the unique identifier for the handler
	// This ID is used for routing gRPC requests to the correct handler
	ID(ctx context.Context) (string, error)
	// HandleGatewayMessage is called when a message is received from a gateway
	HandleGatewayMessage(ctx context.Context, gatewayID string, req *jsonrpc.Request[json.RawMessage]) error
}

var _ GatewayConnector = (*UnimplementedGatewayConnector)(nil)

type UnimplementedGatewayConnector struct{}

func (u *UnimplementedGatewayConnector) AddHandler(ctx context.Context, methods []string, handler GatewayConnectorHandler) error {
	return errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) RemoveHandler(ctx context.Context, methods []string) error {
	return errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) SendToGateway(ctx context.Context, gatewayID string, resp *jsonrpc.Response[json.RawMessage]) error {
	return errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) SignMessage(ctx context.Context, msg []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) GatewayIDs(ctx context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) DonID(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) AwaitConnection(ctx context.Context, gatewayID string) error {
	return errors.New("not implemented")
}
