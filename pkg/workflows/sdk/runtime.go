package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var BreakErr = capabilities.ErrStopExecution

type Emitter interface {
	// Emit sends a message with the given message and labels to the configured collector.
	//
	// TODO(mstreet3): Emit and custmsg.Labeler should be context aware.  Update signature once
	// WASM can support context.
	Emit(msg string, labels map[string]any) error
}

type EmitterFunc func(msg string, labels map[string]any) error

func (f EmitterFunc) Emit(msg string, labels map[string]any) error {
	return f(msg, labels)
}

// Guest interface
type Runtime interface {
	Logger() logger.Logger
	Fetch(req FetchRequest) (FetchResponse, error)

	Emitter
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
