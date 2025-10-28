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
	URL           string            `json:"url"`                 // URL to query, only http and https protocols are supported.
	Method        string            `json:"method,omitempty"`    // HTTP verb, defaults to GET.
	Headers       map[string]string `json:"headers,omitempty"`   // HTTP headers, defaults to empty.
	Body          []byte            `json:"body,omitempty"`      // HTTP request body
	TimeoutMs     uint32            `json:"timeoutMs,omitempty"` // Timeout in milliseconds
	CacheSettings CacheSettings     `json:"cacheSettings"`       // Best-effort cache control for the request

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
	ErrorMessage            string            `json:"errorMessage,omitempty"`            // error message for all errors except HTTP errors returned by external endpoints
	IsExternalEndpointError bool              `json:"isExternalEndpointError,omitempty"` // indicates whether the error is from a faulty external endpoint (e.g. timeout, response size) vs error introduced internally by gateway execution
	IsValidationError       bool              `json:"isValidationError,omitempty"`       // indicates whether the error is a validation error (e.g. blocked HTTP header, blocked IP addresses)
	StatusCode              int               `json:"statusCode,omitempty"`              // HTTP status code
	Headers                 map[string]string `json:"headers,omitempty"`                 // HTTP headers
	Body                    []byte            `json:"body,omitempty"`                    // HTTP response body
	ExternalEndpointLatency time.Duration     `json:"externalEndpointLatency,omitempty"` // Latency of the external endpoint
}
