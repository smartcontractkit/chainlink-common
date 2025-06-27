package chipingress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

var (
	defaultCfg = chipIngressClientConfig{
		transportCredentials: insecure.NewCredentials(),
		headerProvider:       nil,
		authority:            "",
	}
)

func TestClient(t *testing.T) {

	t.Run("NewClient", func(t *testing.T) {
		// Create new client
		client, err := NewChipIngressClient("localhost:8080",
			WithTransportCredentials(insecure.NewCredentials()))
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("NewClient errors when address is empty", func(t *testing.T) {
		client, err := NewChipIngressClient("")
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "invalid address format: missing port in address")
	})

	t.Run("invalid address format", func(t *testing.T) {
		// Address without port will cause net.SplitHostPort to fail
		client, err := NewChipIngressClient("invalid-address-format")
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "missing port in address")
	})

	t.Run("valid address with port", func(t *testing.T) {
		client, err := NewChipIngressClient("localhost:8080")
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

}

func TestNewEvent(t *testing.T) {

	// Create new event
	testProto := pb.PingResponse{Message: "testing"}
	protoBytes, err := proto.Marshal(&testProto)
	assert.NoError(t, err)
	event, err := NewEvent("some-domain_here", "platform.on_chain.forwarder.ReportProcessed", protoBytes)
	assert.NoError(t, err)

	// There should be no validation errors
	err = event.Validate()
	assert.NoError(t, err)

	// Assert fields were set as expected
	assert.Equal(t, "some-domain_here", event.Source())
	assert.Equal(t, "platform.on_chain.forwarder.ReportProcessed", event.Type())
	assert.NotEmpty(t, event.ID())

	// Assert the event data was set as expected
	var resultProto pb.PingResponse
	err = proto.Unmarshal(event.Data(), &resultProto)
	assert.NoError(t, err)
	assert.Equal(t, testProto.Message, resultProto.Message)
}
func TestOptions(t *testing.T) {

	t.Run("WithTransportCredentials", func(t *testing.T) {
		creds := insecure.NewCredentials()
		config := defaultCfg
		WithTransportCredentials(creds)(&config)
		assert.Equal(t, creds, config.transportCredentials)
	})

	t.Run("WithHeaderProvider", func(t *testing.T) {
		mockProvider := &mockHeaderProvider{
			headers: map[string]string{"key": "value"},
		}
		config := defaultCfg
		WithHeaderProvider(mockProvider)(&config)
		assert.Equal(t, mockProvider, config.headerProvider)
	})
}

func TestHeaderInterceptor(t *testing.T) {
	// Create a mock header provider
	mockHeaders := map[string]string{
		"test-header-1": "value1",
		"test-header-2": "value2",
	}
	mockProvider := &mockHeaderProvider{
		headers: mockHeaders,
	}

	// Create a mock invoker that captures the context
	var capturedCtx context.Context
	mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	// Get the interceptor from the function in client.go
	interceptor := newHeaderInterceptor(mockProvider)

	// Call the interceptor
	err := interceptor(context.Background(), "testMethod", nil, nil, nil, mockInvoker)
	assert.NoError(t, err)

	// Extract metadata from context and verify headers were added
	md, ok := metadata.FromOutgoingContext(capturedCtx)
	assert.True(t, ok, "Metadata should be in the context")

	// Verify each header was added
	for k, v := range mockHeaders {
		values := md.Get(k)
		assert.Len(t, values, 1, "Should have exactly one value for header %s", k)
		assert.Equal(t, v, values[0], "Header value mismatch for %s", k)
	}
}

// Mock header provider for testing
type mockHeaderProvider struct {
	headers map[string]string
}

func (m *mockHeaderProvider) GetHeaders() map[string]string {
	return m.headers
}

func TestWithTLSAndHTTP2(t *testing.T) {
	serverName := "example.com"
	config := defaultCfg
	WithTLSAndHTTP2(serverName)(&config)
	assert.NotNil(t, config.transportCredentials)
	// Verify it's TLS credentials (we can't easily inspect the internal config)
	assert.IsType(t, credentials.NewTLS(nil), config.transportCredentials)
}

func TestWithAuthority(t *testing.T) {
	authority := "custom-authority.example.com"
	config := defaultCfg
	WithAuthority(authority)(&config)
	assert.Equal(t, authority, config.authority)
}

func TestWithInsecureConnection(t *testing.T) {
	config := defaultCfg
	WithInsecureConnection()(&config)
	assert.Equal(t, insecure.NewCredentials(), config.transportCredentials)
}

func TestNewChipIngressClientWithTLSAndHTTP2(t *testing.T) {
	// This test verifies the option is applied, but doesn't test actual connection
	// since we'd need a real gRPC server for that
	client, err := NewChipIngressClient(
		"example.com:443",
		WithTLSAndHTTP2("example.com"),
	)
	// The client creation should succeed even if connection fails
	// We're testing the option application, not the actual connection
	if err != nil {
		// Connection errors are expected in unit tests
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NotNil(t, client)
	}
}

func TestNewChipIngressClientWithAuthority(t *testing.T) {
	client, err := NewChipIngressClient(
		"localhost:8080",
		WithAuthority("custom-authority.example.com"),
		WithInsecureConnection(),
	)
	// Connection might fail in unit tests, but option should be applied
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	} else {
		assert.NotNil(t, client)
	}
}
