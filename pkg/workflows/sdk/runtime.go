package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var BreakErr = capabilities.ErrStopExecution

type EmitLabeler interface {
	// Emit sends a message to the labeler's destination.
	Emit(string) error

	// With sets the labels for the message to be emitted.  Labels are passed as key-value pairs
	// and are cumulative.
	With(kvs ...string) EmitLabeler
}

// Guest interface
type Runtime interface {
	Logger() logger.Logger
	Fetch(req FetchRequest) (FetchResponse, error)

	// Emitter sends the given message and labels to the configured collector.
	Emitter() EmitLabeler
}

type FetchRequest struct {
	URL       string         `json:"url"`                 // URL to query, only http and https protocols are supported.
	Method    string         `json:"method,omitempty"`    // HTTP verb, defaults to GET.
	Headers   map[string]any `json:"headers,omitempty"`   // HTTP headers, defaults to empty.
	Body      []byte         `json:"body,omitempty"`      // HTTP request body
	TimeoutMs uint32         `json:"timeoutMs,omitempty"` // Timeout in milliseconds
}

type FetchResponse struct {
	ExecutionError bool           `json:"executionError"`         // true if there were non-HTTP errors. false if HTTP request was sent regardless of status (2xx, 4xx, 5xx)
	ErrorMessage   string         `json:"errorMessage,omitempty"` // error message in case of failure
	StatusCode     uint8          `json:"statusCode"`             // HTTP status code
	Headers        map[string]any `json:"headers,omitempty"`      // HTTP headers
	Body           []byte         `json:"body,omitempty"`         // HTTP response body
}
