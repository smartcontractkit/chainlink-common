package gateway

const (
	MethodHTTPAction = "http_action"
)

type OutboundHTTPRequest struct {
	URL       string            `json:"url"`                 // URL to query, only http and https protocols are supported.
	Method    string            `json:"method,omitempty"`    // HTTP verb, defaults to GET.
	Headers   map[string]string `json:"headers,omitempty"`   // HTTP headers, defaults to empty.
	Body      []byte            `json:"body,omitempty"`      // HTTP request body
	TimeoutMs uint32            `json:"timeoutMs,omitempty"` // Timeout in milliseconds

	// Maximum number of bytes to read from the response body.  If the gateway max response size is smaller than this value, the gateway max response size will be used.
	MaxResponseBytes uint32 `json:"maxBytes,omitempty"`
	WorkflowID       string
}

type OutboundHTTPResponse struct {
	ErrorMessage string            `json:"errorMessage,omitempty"` // error message in case of execution errors. i.e. errors before or after attempting HTTP call to external client
	StatusCode   int               `json:"statusCode,omitempty"`   // HTTP status code
	Headers      map[string]string `json:"headers,omitempty"`      // HTTP headers
	Body         []byte            `json:"body,omitempty"`         // HTTP response body
}
