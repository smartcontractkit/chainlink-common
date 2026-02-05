package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"time"
)

const (
	// Note: any addition to this list must be reflected in the handler's Methods() function.
	MethodHTTPAction = "http_action"
)

// CacheSettings defines cache control options for outbound HTTP requests.
type CacheSettings struct {
	MaxAgeMs int32 `json:"maxAgeMs,omitempty"` // Maximum age of a cached response in milliseconds.
	Store    bool  `json:"store,omitempty"`    // If true, cache the response.
	// Deprecated: positive MaxAgeMs implies ReadFromCache is true
	ReadFromCache bool `json:"readFromCache,omitempty"` // If true, attempt to read a cached response for the request
}

// OutboundHTTPRequest represents an HTTP request to be sent from workflow node to the gateway.
type OutboundHTTPRequest struct {
	URL           string              `json:"url"`                    // URL to query, only http and https protocols are supported.
	Method        string              `json:"method,omitempty"`       // HTTP verb, defaults to GET.
	Headers       map[string]string   `json:"headers,omitempty"`      // HTTP headers, defaults to empty.
	MultiHeaders  map[string][]string `json:"multiHeaders,omitempty"` // HTTP headers with all values preserved
	Body          []byte              `json:"body,omitempty"`         // HTTP request body
	TimeoutMs     uint32              `json:"timeoutMs,omitempty"`    // Timeout in milliseconds
	CacheSettings CacheSettings       `json:"cacheSettings"`          // Best-effort cache control for the request

	// Maximum number of bytes to read from the response body.  If the gateway max response size is smaller than this value, the gateway max response size will be used.
	MaxResponseBytes uint32 `json:"maxBytes,omitempty"`
	WorkflowID       string `json:"workflowId"`
	WorkflowOwner    string `json:"workflowOwner"`
}

// Hash generates a hash of the request for caching purposes.
// WorkflowID is not included in the hash because cached responses can be used across workflows
func (req OutboundHTTPRequest) Hash() string {
	s := sha256.New()
	sep := []byte("/")

	s.Write([]byte(req.WorkflowOwner))
	s.Write(sep)
	s.Write([]byte(req.URL))
	s.Write(sep)
	s.Write([]byte(req.Method))
	s.Write(sep)
	s.Write(req.Body)
	s.Write(sep)

	// To ensure deterministic order, iterate headers in sorted order
	keys := make([]string, 0, len(req.Headers))
	for k := range req.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		s.Write([]byte(key))
		s.Write(sep)
		s.Write([]byte(req.Headers[key]))
		s.Write(sep)
	}

	s.Write([]byte(strconv.FormatUint(uint64(req.MaxResponseBytes), 10)))

	return hex.EncodeToString(s.Sum(nil))
}

// OutboundHTTPResponse represents the response from gateway to workflow node.
type OutboundHTTPResponse struct {
	// ErrorMessage contains error details for gateway-level errors (validation, internal errors, or external endpoint failures like timeouts).
	// This field is empty when the request successfully reaches the customer's endpoint and returns a response (regardless of HTTP status code).
	ErrorMessage string `json:"errorMessage,omitempty"`

	// IsExternalEndpointError indicates the request was sent to the customer's endpoint but failed while sending the request or receiving the response.
	// (e.g., connection timeout, response too large, unreachable host). When true, ErrorMessage contains the failure details.
	IsExternalEndpointError bool `json:"isExternalEndpointError,omitempty"`

	// IsValidationError indicates the request was blocked by the gateway BEFORE being sent to the customer's endpoint
	// due to policy violations (e.g., blocked HTTP headers, blocked IP addresses, invalid URL).
	// This is distinct from StatusCode 4xx, which would indicate the customer's endpoint received and rejected the request.
	IsValidationError bool `json:"isValidationError,omitempty"`

	// StatusCode is the HTTP status code returned by the customer's endpoint (e.g., 2xx, 4xx, 5xx).
	// This field is only populated when the request successfully reaches the customer's endpoint and the response is received.
	StatusCode int `json:"statusCode,omitempty"`

	Headers                 map[string]string   `json:"headers,omitempty"`                 // HTTP headers returned by the customer's endpoint (deprecated: use MultiHeaders, contains first value only for backward compatibility)
	MultiHeaders            map[string][]string `json:"multiHeaders,omitempty"`            // HTTP headers with all values preserved
	Body                    []byte              `json:"body,omitempty"`                    // HTTP response body returned by the customer's endpoint
	ExternalEndpointLatency time.Duration       `json:"externalEndpointLatency,omitempty"` // Time taken by the customer's endpoint to respond
}
