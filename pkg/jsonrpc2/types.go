package jsonrpc2

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
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

	//ErrLimitExceeded is returned when a message was refused due to
	//exceeding a limit set by the server.
	ErrLimitExceeded int64 = -32005
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
	normalizedParams, err := Normalize(r.Params)
	if err != nil {
		return "", fmt.Errorf("error normalizing params: %w", err)
	}
	digestable, err := Normalize(map[string]interface{}{
		"jsonrpc": r.Version,
		"id":      r.ID,
		"method":  r.Method,
		"params":  normalizedParams,
		// Auth is intentionally excluded from the digest
	})
	if err != nil {
		return "", fmt.Errorf("error normalizing request: %w", err)
	}
	canonicalJSONBytes, err := json.Marshal(digestable)
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
	Result  *Result    `json:"result,omitempty"`
	Error   *WireError `json:"error,omitempty"`
}

func (r *Response[Result]) Digest() (string, error) {
	normalizedResult, err := Normalize(r.Result)
	if err != nil {
		return "", fmt.Errorf("error normalizing result: %w", err)
	}
	digestable, err := Normalize(map[string]interface{}{
		"jsonrpc": r.Version,
		"id":      r.ID,
		"result":  normalizedResult,
		"error":   r.Error,
	})
	if err != nil {
		return "", fmt.Errorf("error normalizing response: %w", err)
	}

	canonicalJSONBytes, err := json.Marshal(digestable)
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

// Normalize is a utility function that normalizes a value to ensure consistent hashing.
// It recursively processes maps and slices to ensure consistent ordering and structure.
// 
// Types handled:
// - map[string]interface{}: Keys are sorted alphabetically, and values are recursively normalized.
// - []interface{}: Elements are recursively normalized.
// - Other types (e.g., structs, strings, numbers): Returned as-is without modification.
//
// Example usage:
// Input: map[string]interface{}{"b": 2, "a": 1}
// Output: map[string]interface{}{"a": 1, "b": 2}
//
// Input: []interface{}{map[string]interface{}{"b": 2, "a": 1}, 3}
// Output: []interface{}{map[string]interface{}{"a": 1, "b": 2}, 3}
//
// This function is essential for ensuring consistent hashing of complex data structures.
func Normalize(v any) (any, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		sorted := make(map[string]interface{}, len(val))
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			normVal, err := Normalize(val[k])
			if err != nil {
				return nil, err
			}
			sorted[k] = normVal
		}
		return sorted, nil
	case []interface{}:
		normalizedSlice := make([]interface{}, len(val))
		for i, e := range val {
			norm, err := Normalize(e)
			if err != nil {
				return nil, err
			}
			normalizedSlice[i] = norm
		}
		return normalizedSlice, nil
	default:
		// For structs, strings, numbers, etc.
		return val, nil
	}
}
