package http

import (
	"bytes"
	"io"
	netHttp "net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestUnrestrictedClient(t *testing.T) {
	t.Parallel()

	client := NewUnrestrictedClient()
	assert.True(t, client.Transport.(*netHttp.Transport).DisableCompression)
	client.Transport = newMockTransport()

	netReq, err := netHttp.NewRequestWithContext(t.Context(), "GET", "http://localhost", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	req := &Request{
		Client:  client,
		Request: netReq,
		Config:  RequestConfig{SizeLimit: 1000},
		Logger:  logger.Nop(),
	}

	response, statusCode, headers, err := req.SendRequest()
	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, `{"foo":123}`, string(response))
}

type mockTransport struct{}

func newMockTransport() netHttp.RoundTripper {
	return &mockTransport{}
}

func (t *mockTransport) RoundTrip(req *netHttp.Request) (*netHttp.Response, error) {
	// Create mocked http.Response
	response := &netHttp.Response{
		Header:     make(netHttp.Header),
		Request:    req,
		StatusCode: netHttp.StatusOK,
	}
	response.Header.Set("Content-Type", "application/json")

	responseBody := `{"foo":123}`
	response.Body = io.NopCloser(strings.NewReader(responseBody))
	return response, nil
}

func TestSendRequestReader_WithSizeLimit(t *testing.T) {
	t.Parallel()

	client := NewUnrestrictedClient()
	client.Transport = newMockTransport()

	netReq, err := netHttp.NewRequestWithContext(t.Context(), "GET", "http://localhost", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	req := &Request{
		Client:  client,
		Request: netReq,
		Config:  RequestConfig{SizeLimit: 100},
		Logger:  logger.Nop(),
	}

	reader, statusCode, headers, err := req.SendRequestReader()
	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))

	// Verify reader works
	body, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":123}`, string(body))
	assert.NoError(t, reader.Close())
}
