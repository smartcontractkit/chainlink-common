package jsonrpc2

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	// ErrLimitExceeded is returned when a message was refused due to user exceeding
	// a limit (e.g. number of requests per second).
	ErrLimitExceeded int64 = -32002

	// ErrConflict is returned when a request conflicts with an existing request.
	ErrConflict int64 = -32003
)

// Wrapping/unwrapping Message objects into JSON RPC ones folllowing https://www.jsonrpc.org/specification
type Request[Params any] struct {
	Version string  `json:"jsonrpc"`
	ID      string  `json:"id"`
	Method  string  `json:"method"`
	Params  *Params `json:"params"`
	// Auth is used to store the JWT token for the request. It is not part of the JSON RPC specification.
	// JWT token can be part of the request payload or attached to the request header.
	Auth string `json:"auth,omitempty"`
}

func (r *Request[Params]) ServiceName() string {
	return strings.Split(r.Method, ".")[0]
}

// Digest returns a digest of the request. This is used for signature verification.
// The digest is a SHA256 hash of the JSON string of the request.
func (r *Request[Params]) Digest() (string, error) {
	canonicalJSONBytes, err := json.Marshal(Request[Params]{
		Version: r.Version,
		ID:      r.ID,
		Method:  r.Method,
		Params:  r.Params,
		// Auth is intentionally excluded from the digest
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(canonicalJSONBytes)
	digestBytes := hasher.Sum(nil)

	return hex.EncodeToString(digestBytes), nil
}

type Response[Result any] struct {
	Version string     `json:"jsonrpc"`
	ID      string     `json:"id"`
	Method  string     `json:"method"`
	Result  *Result    `json:"result,omitempty"`
	Error   *WireError `json:"error,omitempty"`
}

func (r *Response[Result]) Digest() (string, error) {
	canonicalJSONBytes, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	hasher := sha256.New()
	if _, err := hasher.Write(canonicalJSONBytes); err != nil {
		return "", fmt.Errorf("error writing to hasher: %w", err)
	}
	digestBytes := hasher.Sum(nil)

	return hex.EncodeToString(digestBytes), nil
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
