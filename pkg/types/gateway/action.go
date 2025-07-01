package gateway

const (
	MethodHTTPAction = "http_action"
)

// CacheSettings defines cache control options for outbound HTTP requests.
type CacheSettings struct {
	ReadFromCache bool  `json:"readFromCache,omitempty"` // If true, attempt to read a cached response for the request
	StoreInCache  bool  `json:"storeInCache,omitempty"`  // If true, store the response in cache for the given TTL
	Ttlms         int32 `json:"ttlMs,omitempty"`         // Time-to-live for the cache entry in milliseconds. Only applicable if StoreInCache is true.
}

// OutboundHTTPRequest represents an HTTP request to be sent from workflow node to the gateway.
type OutboundHTTPRequest struct {
	URL           string            `json:"url"`                     // URL to query, only http and https protocols are supported.
	Method        string            `json:"method,omitempty"`        // HTTP verb, defaults to GET.
	Headers       map[string]string `json:"headers,omitempty"`       // HTTP headers, defaults to empty.
	Body          []byte            `json:"body,omitempty"`          // HTTP request body
	TimeoutMs     uint32            `json:"timeoutMs,omitempty"`     // Timeout in milliseconds
	CacheSettings CacheSettings     `json:"cacheSettings,omitempty"` // Best-effort cache control for the request

	// Maximum number of bytes to read from the response body.  If the gateway max response size is smaller than this value, the gateway max response size will be used.
	MaxResponseBytes uint32 `json:"maxBytes,omitempty"`
	WorkflowID       string `json:"workflowId"`
}

// OutboundHTTPResponse represents the response from gateway to workflow node.
type OutboundHTTPResponse struct {
	ErrorMessage string            `json:"errorMessage,omitempty"` // error message in case of execution errors. i.e. errors before or after attempting HTTP call to external client
	StatusCode   int               `json:"statusCode,omitempty"`   // HTTP status code
	Headers      map[string]string `json:"headers,omitempty"`      // HTTP headers
	Body         []byte            `json:"body,omitempty"`         // HTTP response body
}
