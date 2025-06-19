package jsonrpc2

import "encoding/json"

var (
	// ErrUnknown should be used for all non-coded errors.
	ErrUnknown = NewError(-32001, "JSON RPC unknown error")
	// ErrParse is used when invalid JSON was received by the server.
	ErrParse = NewError(-32700, "JSON RPC parse error")
	//ErrInvalidRequest is used when the JSON sent is not a valid Request object.
	ErrInvalidRequest = NewError(-32600, "JSON RPC invalid request")
	// ErrMethodNotFound should be returned by the handler when the method does
	// not exist / is not available.
	ErrMethodNotFound = NewError(-32601, "JSON RPC method not found")
	// ErrInvalidParams should be returned by the handler when method
	// parameter(s) were invalid.
	ErrInvalidParams = NewError(-32602, "JSON RPC invalid params")
	// ErrInternal is not currently returned but defined for completeness.
	ErrInternal = NewError(-32603, "JSON RPC internal error")

	//ErrServerOverloaded is returned when a message was refused due to a
	//server being temporarily unable to accept any new messages.
	ErrServerOverloaded = NewError(-32000, "JSON RPC overloaded")
)

const (
	JsonRpcVersion = "2.0"
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

func (err *WireError) Error() string {
	return err.Message
}

func NewError(code int64, message string) error {
	return &WireError{
		Code:    code,
		Message: message,
	}
}
