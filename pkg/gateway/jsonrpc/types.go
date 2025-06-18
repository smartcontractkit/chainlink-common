package jsonrpc

import "encoding/json"

type Request struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	// Auth is used to store the JWT token for the request. It is not part of the JSON RPC specification.
	// JWT token can be part of the request payload or attached to the request header.
	Auth string `json:"auth,omitempty"`
}

type Response struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// JSON-RPC error can only be sent to users. It is not used for messages between Gateways and Nodes.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}
