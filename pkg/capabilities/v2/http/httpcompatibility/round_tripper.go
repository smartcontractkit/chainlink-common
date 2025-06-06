package httpcompatibility

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	chttp "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/http"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type RoundTripper struct {
	NodeRuntime sdk.NodeRuntime
}

func NewConsensusRoundTripper[T, C any](
	wcx *sdk.WorkflowContext[C],
	runtime sdk.Runtime,
	fn func(wcx *sdk.WorkflowContext[C], roundTripper http.RoundTripper) (T, error),
	ca sdk.ConsensusAggregation[T]) sdk.Promise[T] {
	return sdk.RunInNodeMode[C, T](wcx, runtime, func(wcx *sdk.WorkflowContext[C], nodeRuntime sdk.NodeRuntime) (T, error) {
		roundTripper := &RoundTripper{NodeRuntime: nodeRuntime}
		return fn(wcx, roundTripper)
	}, ca)
}

func (r RoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	client := chttp.Client{}
	var method chttp.Method
	switch request.Method {
	case http.MethodGet:
		method = chttp.Method_GET
	case http.MethodPost:
		method = chttp.Method_POST
	case http.MethodPut:
		method = chttp.Method_PUT
	case http.MethodDelete:
		method = chttp.Method_DELETE
	case http.MethodPatch:
		method = chttp.Method_PATCH
	default:
		return nil, fmt.Errorf("unsupported http method: %s", request.Method)
	}

	// TODO headers can be repeated...
	headers := map[string]string{}
	for name, values := range request.Header {
		for _, value := range values {
			headers[name] = value
		}
	}

	var body []byte
	if request.Body != nil {
		var err error
		body, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	response, err := client.SendRequest(r.NodeRuntime, &chttp.Request{
		Url:     request.URL.String(),
		Method:  method,
		Headers: headers,
		Body:    body,
		// TODO?
		TimeoutMs: 0,
	}).Await()

	if err != nil {
		return nil, err
	}

	// TODO headers can be repeated...
	responseHeaders := http.Header{}
	for name, value := range response.Headers {
		responseHeaders[name] = []string{value}
	}

	return &http.Response{
		Status:     http.StatusText(int(response.StatusCode)),
		StatusCode: int(response.StatusCode),
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     responseHeaders,
		Body:       io.NopCloser(bytes.NewReader(response.Body)),
		// TODO verify I should be setting this given the other field's values.
		ContentLength: int64(len(response.Body)),
		// TransferEncoding: nil,
		// Close:            false,
		// Uncompressed:     false,
	}, nil
}
