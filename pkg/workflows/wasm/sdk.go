package wasm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type Runtime struct {
	Logger logger.Logger
	Fetch  func(req FetchRequest) FetchResponse
}

type RuntimeConfig struct {
	MaxFetchResponseSizeBytes int64
}

const (
	defaultMaxFetchResponseSizeBytes = 5 * 1024
)

func defaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		MaxFetchResponseSizeBytes: defaultMaxFetchResponseSizeBytes,
	}
}

var _ sdk.Runtime = (*Runtime)(nil)

type FetchRequest struct {
	URL       string         `json:"url"`                 // URL to query, only http and https protocols are supported.
	Method    string         `json:"method,omitempty"`    // HTTP verb, defaults to GET.
	Headers   map[string]any `json:"headers,omitempty"`   // HTTP headers, defaults to empty.
	Body      []byte         `json:"body,omitempty"`      // HTTP request body
	TimeoutMs uint32         `json:"timeoutMs,omitempty"` // Timeout in milliseconds
}

type FetchResponse struct {
	Success    bool           `json:"success"`           // true if HTTP request was successful
	StatusCode uint8          `json:"statusCode"`        // HTTP status code
	Headers    map[string]any `json:"headers,omitempty"` // HTTP headers
	Body       []byte         `json:"body,omitempty"`    // HTTP response body
}
