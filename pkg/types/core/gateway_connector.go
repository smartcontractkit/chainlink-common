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
	// DonID returns the DON ID of the node managing this connector.
	DonID(ctx context.Context) (string, error)
	AwaitConnection(ctx context.Context, gatewayID string) error
}

// MultiGatewayConnector extends GatewayConnector with multi-gateway routing methods.
// Implementations delegate GatewayIDs to GatewayIDsForDon for backward-compatible routing.
type MultiGatewayConnector interface {
	GatewayConnector

	// GatewayIDsForDon returns gateway IDs whose per-gateway DonID (gateway DON) matches donID.
	// Empty donID is treated as PrimaryDonID().
	GatewayIDsForDon(ctx context.Context, donID string) ([]string, error)
	// DonIDForGateway returns the gateway DON configured on the gateway entry for gatewayID.
	DonIDForGateway(ctx context.Context, gatewayID string) (string, error)
	// PrimaryDonID returns the default gateway DON for outbound request routing.
	PrimaryDonID(ctx context.Context) (string, error)
}

// GatewayConnector user (node) implements application logic in the Handler interface.
type GatewayConnectorHandler interface {
	// ID returns the unique identifier for the handler
	// This ID is used for routing gRPC requests to the correct handler
	ID(ctx context.Context) (string, error)
	// HandleGatewayMessage is called when a message is received from a gateway
	HandleGatewayMessage(ctx context.Context, gatewayID string, req *jsonrpc.Request[json.RawMessage]) error
}

var (
	_ GatewayConnector         = (*UnimplementedGatewayConnector)(nil)
	_ MultiGatewayConnector = (*UnimplementedGatewayConnector)(nil)
)

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

func (u *UnimplementedGatewayConnector) GatewayIDsForDon(ctx context.Context, donID string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) DonIDForGateway(ctx context.Context, gatewayID string) (string, error) {
	return "", errors.New("not implemented")
}

func (u *UnimplementedGatewayConnector) PrimaryDonID(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}
