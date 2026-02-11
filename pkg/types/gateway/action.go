package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
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
	URL    string `json:"url"`              // URL to query, only http and https protocols are supported.
	Method string `json:"method,omitempty"` // HTTP verb, defaults to GET.

	// Deprecated: Use MultiHeaders instead. Headers is a comma joined string of all values for a given header for backwards
	// compatability.
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

// ErrBothHeadersAndMultiHeaders is returned when both Headers and MultiHeaders are non-empty.
// Callers must use only one of the two for a given request or response.
var ErrBothHeadersAndMultiHeaders = errors.New("must not set both Headers and MultiHeaders; use MultiHeaders only")

// Hash generates a hash of the request for caching purposes.
// WorkflowID is not included in the hash because cached responses can be used across workflows.
// Headers are included in a deterministic order: MultiHeaders is used when non-empty, otherwise Headers.
// When using MultiHeaders, keys and values within each key are sorted for determinism.
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

	writeHeadersToHash(s, sep, req.Headers, req.MultiHeaders)

	s.Write([]byte(strconv.FormatUint(uint64(req.MaxResponseBytes), 10)))

	return hex.EncodeToString(s.Sum(nil))
}

// writeHeadersToHash writes a deterministic encoding of headers into the hash.
// MultiHeaders is used when non-empty; otherwise Headers is used. Keys and values are sorted.
func writeHeadersToHash(s hash.Hash, sep []byte, headers map[string]string, multiHeaders map[string][]string) {
	if len(multiHeaders) > 0 {
		keys := make([]string, 0, len(multiHeaders))
		for k := range multiHeaders {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			vals := multiHeaders[key]
			valsCopy := make([]string, len(vals))
			copy(valsCopy, vals)
			sort.Strings(valsCopy)
			s.Write([]byte(key))
			s.Write(sep)
			for _, v := range valsCopy {
				s.Write([]byte(v))
				s.Write(sep)
			}
		}
		return
	}
	if len(headers) > 0 {
		keys := make([]string, 0, len(headers))
		for k := range headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			s.Write([]byte(key))
			s.Write(sep)
			s.Write([]byte(headers[key]))
			s.Write(sep)
		}
	}
}

// Validate returns an error if both Headers and MultiHeaders are non-empty.
// Callers must populate only one of the two.
func (req *OutboundHTTPRequest) Validate() error {
	if len(req.Headers) > 0 && len(req.MultiHeaders) > 0 {
		return ErrBothHeadersAndMultiHeaders
	}
	return nil
}

// Validate returns an error if both Headers and MultiHeaders are non-empty.
// Callers must populate only one of the two.
func (resp *OutboundHTTPResponse) Validate() error {
	if len(resp.Headers) > 0 && len(resp.MultiHeaders) > 0 {
		return ErrBothHeadersAndMultiHeaders
	}
	return nil
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

	// Deprecated: Use MultiHeaders instead. Headers is a comma joined string of all values for a given header for backwards
	// compatability.
	Headers                 map[string]string   `json:"headers,omitempty"`                 // HTTP headers returned by the customer's endpoint
	MultiHeaders            map[string][]string `json:"multiHeaders,omitempty"`            // HTTP headers with all values preserved
	Body                    []byte              `json:"body,omitempty"`                    // HTTP response body returned by the customer's endpoint
	ExternalEndpointLatency time.Duration       `json:"externalEndpointLatency,omitempty"` // Time taken by the customer's endpoint to respond
}
