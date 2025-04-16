package chipingress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials/insecure"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestClient(t *testing.T) {

	t.Run("NewClient", func(t *testing.T) {
		// Create new client
		client, err := NewChipIngressClient(
			"localhost:8080",
			WithLogger(zap.NewNop()),
			WithTransportCredentials(insecure.NewCredentials()))
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("NewClient errors when address is empty", func(t *testing.T) {
		client, err := NewChipIngressClient("", WithLogger(zap.NewNop()))
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "is empty")
	})

	t.Run("Publish", func(t *testing.T) {

		mockClient := &mocks.ChipIngressClient{}

		mockClient.
			On("Publish", mock.Anything, mock.Anything).
			Return(&pb.PublishResponse{}, nil)

		client := &chipIngressClient{
			log:    zap.NewNop(),
			client: mockClient,
		}

		// Create new event
		testProto := pb.PingResponse{Message: "testing"}
		protoBytes, err := proto.Marshal(&testProto)
		require.NoError(t, err)
		event, err := NewEvent("some-domain_here", "platform.on_chain.forwarder.ReportProcessed", protoBytes)
		require.NoError(t, err)

		// Publish event
		_, err = client.Publish(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("Publish errors when validation fails", func(t *testing.T) {

		client := &chipIngressClient{
			log: zap.NewNop(),
		}

		event := ce.NewEvent()
		event.SetExtension("hello", "world")

		_, err := client.Publish(context.Background(), event)
		assert.ErrorContains(t, err, "validation failed")
	})

}

func TestValidateEvents(t *testing.T) {
	// Should fail
	event1 := ce.NewEvent()
	event1.SetExtension("hello-1", "world1")

	// Should fail
	event2 := ce.NewEvent()
	event2.SetExtension("hello-2", "world2")

	// Should pass
	event3 := ce.NewEvent()
	event3.SetExtension("hello3", "world3")
	event3.SetID("id")
	event3.SetType("type")
	event3.SetSource("source")

	events := []ce.Event{event1, event2, event3}
	err := validateEvents(events)
	assert.Error(t, err)

	assert.ErrorContains(t, err, "validation failed for 2 of 3 events")
	assert.ErrorContains(t, err, "Event ID  (index 0)")
	assert.ErrorContains(t, err, "Event ID  (index 1)")
	assert.NotContains(t, err.Error(), "Event ID  (index 2)")
}

func TestPing(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Ping", mock.Anything, &pb.EmptyRequest{}).
			Return(&pb.PingResponse{Message: "Pong"}, nil)

		chipIngressClient := &chipIngressClient{
			log:    zap.NewNop(),
			client: clientMock,
		}

		resp, err := chipIngressClient.Ping(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, "Pong", resp)
	})

	t.Run("errors when ping fails", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Ping", mock.Anything, &pb.EmptyRequest{}).
			Return(nil, assert.AnError)

		chipIngressClient, err := NewChipIngressClient("test")
		assert.NoError(t, err)

		resp, err := chipIngressClient.Ping(t.Context())
		assert.Error(t, err)
		assert.Empty(t, resp)
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
	t.Run("WithLogger", func(t *testing.T) {
		logger := zap.NewNop()
		config := defaultConfig()
		WithLogger(logger)(&config)
		assert.Equal(t, logger, config.log)
	})

	t.Run("WithTransportCredentials", func(t *testing.T) {
		creds := insecure.NewCredentials()
		config := defaultConfig()
		WithTransportCredentials(creds)(&config)
		assert.Equal(t, creds, config.transportCredentials)
	})
}
func TestPublishBatch(t *testing.T) {
	t.Run("successful batch publish", func(t *testing.T) {
		mockClient := &mocks.ChipIngressClient{}

		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&pb.PublishResponse{}, nil)

		client := &chipIngressClient{
			log:    zap.NewNop(),
			client: mockClient,
		}

		// Create test events
		testProto1 := pb.PingResponse{Message: "testing1"}
		protoBytes1, err := proto.Marshal(&testProto1)
		require.NoError(t, err)
		event1, err := NewEvent("domain1", "entity.event1", protoBytes1)
		require.NoError(t, err)

		testProto2 := pb.PingResponse{Message: "testing2"}
		protoBytes2, err := proto.Marshal(&testProto2)
		require.NoError(t, err)
		event2, err := NewEvent("domain2", "entity.event2", protoBytes2)
		require.NoError(t, err)

		events := []ce.Event{event1, event2}

		// Publish events in batch
		_, err = client.PublishBatch(context.Background(), events)
		assert.NoError(t, err)
	})

	t.Run("errors when validation fails", func(t *testing.T) {
		client := &chipIngressClient{
			log: zap.NewNop(),
		}

		// Create invalid events
		event1 := ce.NewEvent() // Missing required fields
		event2 := ce.NewEvent() // Missing required fields

		events := []ce.Event{event1, event2}

		_, err := client.PublishBatch(context.Background(), events)
		assert.ErrorContains(t, err, "validation failed")
		assert.ErrorContains(t, err, "Event ID  (index 0)")
		assert.ErrorContains(t, err, "Event ID  (index 1)")
	})

	t.Run("errors when publish batch fails", func(t *testing.T) {
		mockClient := &mocks.ChipIngressClient{}

		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		client := &chipIngressClient{
			log:    zap.NewNop(),
			client: mockClient,
		}

		// Create valid events
		testProto := pb.PingResponse{Message: "testing"}
		protoBytes, err := proto.Marshal(&testProto)
		require.NoError(t, err)
		event, err := NewEvent("domain", "entity.event", protoBytes)
		require.NoError(t, err)

		events := []ce.Event{event}

		// Publish events should return error
		_, err = client.PublishBatch(context.Background(), events)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
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

	// Create the interceptor function
	cfg := defaultConfig()
	cfg.headerProvider = mockProvider

	// Call the interceptor directly
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add dynamic headers from provider if available
		if cfg.headerProvider != nil {
			for k, v := range cfg.headerProvider.GetHeaders() {
				ctx = metadata.AppendToOutgoingContext(ctx, k, v)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}

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
