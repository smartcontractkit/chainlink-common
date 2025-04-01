package client

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"testing"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

type fakeClient struct{}

func TestClient(t *testing.T) {

	// Create new client
	client := &chipIngressClient{
		log:    zap.NewNop(),
		client: &fakeClient{},
	}

	t.Run("NewClient", func(t *testing.T) {
		// Create new client
		client, err := NewChipIngressClient("localhost:8080")
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Publish", func(t *testing.T) {

		// Create new event
		testProto := pb.PingResponse{Message: "testing"}
		event, err := NewEvent("some-domain_here", "platform.on_chain.forwarder.ReportProcessed", &testProto)
		require.NoError(t, err)

		// Publish event
		_, err = client.Publish(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("Publish errors when validation fails", func(t *testing.T) {

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

func TestNewEvent(t *testing.T) {

	// Create new event
	testProto := pb.PingResponse{Message: "testing"}
	event, err := NewEvent("some-domain_here", "platform.on_chain.forwarder.ReportProcessed", &testProto)
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

func (f *fakeClient) Ping(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "Pong"}, nil
}

func (f *fakeClient) Publish(ctx context.Context, in *cepb.CloudEvent, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
	return &pb.PublishResponse{}, nil
}

func (f *fakeClient) PublishBatch(ctx context.Context, in *pb.CloudEventBatch, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
	return &pb.PublishResponse{}, nil
}
