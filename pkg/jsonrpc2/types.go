package jsonrpc2

import (
	"encoding/json"
	"strings"
)

const (
	JsonRpcVersion = "2.0"

	// ErrUnknown should be used for all non-coded errors.
	ErrUnknown int64 = -32001

	// ErrParse is used when invalid JSON was received by the server.
	ErrParse int64 = -32700

	//ErrInvalidRequest is used when the JSON sent is not a valid Request object.
	ErrInvalidRequest int64 = -32600

	// ErrMethodNotFound should be returned by the handler when the method does
	// not exist / is not available.
	ErrMethodNotFound int64 = -32601

	// ErrInvalidParams should be returned by the handler when method
	// parameter(s) were invalid.
	ErrInvalidParams int64 = -32602

	// ErrInternal is not currently returned but defined for completeness.
	ErrInternal int64 = -32603

	//ErrServerOverloaded is returned when a message was refused due to a
	//server being temporarily unable to accept any new messages.
	ErrServerOverloaded int64 = -32000
)

// Wrapping/unwrapping Message objects into JSON RPC ones folllowing https://www.jsonrpc.org/specification
type Request struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	// Auth is used to store the JWT token for the request. It is not part of the JSON RPC specification.
	// JWT token can be part of the request payload or attached to the request header.
	Auth string `json:"auth,omitempty"`
}

func (r *Request) ServiceName() string {
	return strings.Split(r.Method, ".")[0]
}

type Response struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *WireError      `json:"error,omitempty"`
}

// WireError represents a structured error in a Response.
type WireError struct {
	// Code is an error code indicating the type of failure.
	Code int64 `json:"code"`
	// The Message is a short description of the error.
	Message string `json:"message"`
	// Data is optional structured data containing additional information about the error.
	Data *json.RawMessage `json:"data,omitempty"`
}

func (w *WireError) Error() string {
	return w.Message
}
