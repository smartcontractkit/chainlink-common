package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// BreakErr can be used inside the compute capability function to stop the execution of the workflow.
var BreakErr = capabilities.ErrStopExecution

type MessageEmitter interface {
	// Emit sends a message to the labeler's destination.
	Emit(string) error

	// With sets the labels for the message to be emitted.  Labels are passed as key-value pairs
	// and are cumulative.
	With(kvs ...string) MessageEmitter
}

// Runtime exposes external system calls to workflow authors.
// - `Logger` can be used to log messages
// - `Emitter` can be used to send messages to beholder
// - `Fetch` can be used to make external HTTP calls
type Runtime interface {
	Logger() logger.Logger
	Fetch(req FetchRequest) (FetchResponse, error)

	// Emitter sends the given message and labels to the configured collector.
	Emitter() MessageEmitter
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

// NewEventEmitter creates a new event emitter that wraps the runtime's emitter
// with additional functionality for easier logging as Beholder events.
func NewEventEmitter(runtime Runtime) *EventEmitter {
	return &EventEmitter{
		runtime: runtime,
	}
}

// EventEmitter provides a convenient way to emit messages and log events
// in workflows.
type EventEmitter struct {
	runtime Runtime
	labels  []string
}

// With adds labels to the emitter that will be included in all subsequent emits.
// Labels are passed as key-value pairs and are cumulative.
func (e *EventEmitter) With(kvs ...string) *EventEmitter {
	e.labels = append(e.labels, kvs...)
	return e
}

// Emit sends a message with the accumulated labels.
func (e *EventEmitter) Emit(message string) {
	err := e.runtime.Emitter().With(e.labels...).Emit(message)
	if err != nil {
		e.runtime.Logger().Errorf("failed to emit message: %s", message)
	}
}
