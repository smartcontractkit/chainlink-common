package chipingress

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gp "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

var (
	defaultCfg = clientConfig{
		transportCredentials:  insecure.NewCredentials(),
		insecureConnection:    true, // Default to insecure connection
		host:                  "localhost",
		perRPCCredentials:     nil, // No per-RPC credentials by default
		headerProvider:        nil,
		nopInfoHeaderProvider: nil,
	}
)

func TestClient(t *testing.T) {

	t.Run("NewClient", func(t *testing.T) {
		// Create new client
		client, err := NewClient("localhost:8080",
			WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("NewClient errors when address is empty", func(t *testing.T) {
		client, err := NewClient("")
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "invalid address format: missing port in address")
	})

	t.Run("invalid address format", func(t *testing.T) {
		// Address without port will cause net.SplitHostPort to fail
		client, err := NewClient("invalid-address-format")
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "missing port in address")
	})

	t.Run("valid address with port", func(t *testing.T) {
		client, err := NewClient("localhost:8080")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("NoopClient", func(t *testing.T) {
		client := &NoopClient{}
		assert.NotNil(t, client)

		// Test that it implements the Client interface
		var _ Client = client

		// Test Close returns no error
		err := client.Close()
		require.NoError(t, err)

		// Test Ping returns success
		pingResp, err := client.Ping(t.Context(), &pb.EmptyRequest{})
		require.NoError(t, err)
		assert.NotNil(t, pingResp)
		assert.Equal(t, "pong", pingResp.Message)

		// Test RegisterSchemas returns empty map
		schemas := []*pb.Schema{
			{Subject: "test", Schema: `{"test":"value"}`, Format: 1},
		}
		result, err := client.RegisterSchemas(t.Context(), schemas...)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("Close", func(t *testing.T) {
		rawClient, err := NewClient("localhost:8080")
		require.NoError(t, err)

		client, ok := rawClient.(*client)
		require.True(t, ok)

		require.NoError(t, client.Close())
	})

}

func TestNewEvent(t *testing.T) {

	// Create new event
	testProto := pb.PingResponse{Message: "testing"}
	protoBytes, err := proto.Marshal(&testProto)
	attributes := map[string]any{
		"datacontenttype": "application/protobuf",
		"dataschema":      "https://example.com/schema",
		"subject":         "example-subject",
		"time":            time.Now().Add(-5 * time.Second),
	}
	require.NoError(t, err)

	event, err := NewEvent("some-domain_here", "platform.on_chain.forwarder.ReportProcessed", protoBytes, attributes)
	require.NoError(t, err)

	// There should be no validation errors
	err = event.Validate()
	require.NoError(t, err)

	// Assert fields were set as expected
	assert.Equal(t, "some-domain_here", event.Source())
	assert.Equal(t, "platform.on_chain.forwarder.ReportProcessed", event.Type())
	assert.NotEmpty(t, event.ID())
	assert.Equal(t, "application/protobuf", event.DataContentType())
	assert.Equal(t, "https://example.com/schema", event.DataSchema())
	assert.Equal(t, "example-subject", event.Subject())
	assert.Equal(t, attributes["time"].(time.Time).UTC(), event.Time())
	assert.NotEmpty(t, event.Extensions()["recordedtime"])
	assert.True(t, event.Extensions()["recordedtime"].(ce.Timestamp).After(attributes["time"].(time.Time)))

	// Assert the event data was set as expected
	var resultProto pb.PingResponse
	err = proto.Unmarshal(event.Data(), &resultProto)
	require.NoError(t, err)
	assert.Equal(t, testProto.Message, resultProto.Message)
}

func TestNewEvent_IdempotencyKey(t *testing.T) {
	payload := []byte("body")

	t.Run("non-empty key is set as extension", func(t *testing.T) {
		attrs := map[string]any{IdempotencyKeyAttr: "my-key-123"}
		event, err := NewEvent("domain", "entity", payload, attrs)
		require.NoError(t, err)
		ext := event.Extensions()
		require.NotNil(t, ext)
		assert.Equal(t, "my-key-123", ext[IdempotencyKeyAttr])
	})

	t.Run("empty key is not set as extension", func(t *testing.T) {
		attrs := map[string]any{IdempotencyKeyAttr: ""}
		event, err := NewEvent("domain", "entity", payload, attrs)
		require.NoError(t, err)
		ext := event.Extensions()
		_, present := ext[IdempotencyKeyAttr]
		assert.False(t, present, "empty idempotency key must not appear as an extension")
	})

	t.Run("absent key leaves extension unset", func(t *testing.T) {
		event, err := NewEvent("domain", "entity", payload, nil)
		require.NoError(t, err)
		ext := event.Extensions()
		_, present := ext[IdempotencyKeyAttr]
		assert.False(t, present, "absent idempotency key must not appear as an extension")
	})
}

func TestSanitizeExtensionName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "snake_case", in: "chain_id", want: "chainid"},
		{name: "dotted", in: "k8s.pod.name", want: "k8spodname"},
		{name: "already valid", in: "chainid", want: "chainid"},
		{name: "upper case is lowered", in: "ChainID", want: "chainid"},
		{name: "empty", in: "", want: ""},
		{name: "all invalid characters", in: "---...", want: ""},
		{name: "mixed valid and invalid", in: "Service-Name.1", want: "servicename1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeExtensionName(tt.in))
		})
	}
}

func TestNewEventWithOpts_WithResourceAttributeExtensions(t *testing.T) {
	payload := []byte("body")

	t.Run("sanitized keys/values land on the event", func(t *testing.T) {
		attrs := map[string]string{"chain_id": "1", "k8s.pod.name": "pod-abc"}
		event, err := NewEventWithOpts("domain", "entity", payload, nil, WithResourceAttributeExtensions(attrs))
		require.NoError(t, err)
		ext := event.Extensions()
		assert.Equal(t, "1", ext["chainid"])
		assert.Equal(t, "pod-abc", ext["k8spodname"])
	})

	t.Run("empty sanitized name is dropped", func(t *testing.T) {
		attrs := map[string]string{"---": "value"}
		event, err := NewEventWithOpts("domain", "entity", payload, nil, WithResourceAttributeExtensions(attrs))
		require.NoError(t, err)
		assert.Len(t, event.Extensions(), 1) // only the always-set recordedtime extension
	})

	t.Run("reserved name is skipped", func(t *testing.T) {
		attrs := map[string]string{IdempotencyKeyAttr: "should-not-override", "subject": "should-not-override"}
		event, err := NewEventWithOpts("domain", "entity", payload, map[string]any{IdempotencyKeyAttr: "real-key"}, WithResourceAttributeExtensions(attrs))
		require.NoError(t, err)
		ext := event.Extensions()
		assert.Equal(t, "real-key", ext[IdempotencyKeyAttr])
		assert.Empty(t, event.Subject())
	})

	t.Run("duplicate sanitized names resolve deterministically to sorted-first key", func(t *testing.T) {
		attrs := map[string]string{"service.name": "from-dotted", "service_name": "from-snake"}
		event, err := NewEventWithOpts("domain", "entity", payload, nil, WithResourceAttributeExtensions(attrs))
		require.NoError(t, err)
		// sorted order: "service.name" < "service_name" ('.' < '_' in ASCII), so the dotted key wins.
		assert.Equal(t, "from-dotted", event.Extensions()["servicename"])
	})

	t.Run("omitting all opts is a no-op", func(t *testing.T) {
		event, err := NewEventWithOpts("domain", "entity", payload, nil)
		require.NoError(t, err)
		assert.Len(t, event.Extensions(), 1) // only the always-set recordedtime extension
	})
}

// TestNewEvent_UnchangedSignature is a backward-compatibility guard: NewEvent's exported
// signature must stay exactly as it was before EventOpt/NewEventWithOpts were introduced, and
// must remain equivalent to calling NewEventWithOpts with no opts.
func TestNewEvent_UnchangedSignature(t *testing.T) {
	payload := []byte("body")
	attributes := map[string]any{"subject": "example-subject"}

	viaNewEvent, err := NewEvent("domain", "entity", payload, attributes)
	require.NoError(t, err)

	viaNewEventWithOpts, err := NewEventWithOpts("domain", "entity", payload, attributes)
	require.NoError(t, err)

	assert.Equal(t, viaNewEventWithOpts.Subject(), viaNewEvent.Subject())
	assert.Equal(t, viaNewEventWithOpts.Extensions()["recordedtime"].(ce.Timestamp).Truncate(time.Second),
		viaNewEvent.Extensions()["recordedtime"].(ce.Timestamp).Truncate(time.Second))
	assert.Equal(t, viaNewEventWithOpts.Data(), viaNewEvent.Data())
}

func TestEventToProto(t *testing.T) {
	// Create a test protobuf message
	testProto := pb.PingResponse{Message: "test message"}
	protoBytes, err := proto.Marshal(&testProto)
	require.NoError(t, err)
	t.Run("successful conversion", func(t *testing.T) {
		// Create a CloudEvent
		event, err := NewEvent("test-domain", "test.event.type", protoBytes, nil)
		require.NoError(t, err)

		// Convert to proto
		eventPb, err := EventToProto(event)
		require.NoError(t, err)
		assert.NotNil(t, eventPb)

		// Verify the converted protobuf event has the expected fields
		assert.Equal(t, "test-domain", eventPb.Source)
		assert.Equal(t, "test.event.type", eventPb.Type)
		assert.NotEmpty(t, eventPb.Id)
		assert.NotNil(t, eventPb.Data)
	})

	t.Run("conversion with attributes", func(t *testing.T) {
		// Create event with custom attributes
		attributes := map[string]any{
			"subject":    "test-subject",
			"dataschema": "https://example.com/schema",
		}

		event, err := NewEvent("test-domain", "test.event.type", protoBytes, attributes)
		require.NoError(t, err)

		// Convert to proto
		eventPb, err := EventToProto(event)

		require.NoError(t, err)
		assert.NotNil(t, eventPb)

		// Verify the converted protobuf event has the expected fields
		assert.Equal(t, "test-domain", eventPb.Source)
		assert.Equal(t, "test.event.type", eventPb.Type)
		assert.NotEmpty(t, eventPb.Id)
		assert.NotNil(t, eventPb.Data)
		assert.NotNil(t, eventPb.Attributes["subject"])
		assert.NotNil(t, eventPb.Attributes["dataschema"])

		eventFromPb, err := ProtoToEvent(eventPb)
		require.NoError(t, err)
		assert.NotNil(t, eventFromPb)

		// Verify attributes were preserved
		assert.Equal(t, "test-subject", eventFromPb.Context.GetSubject())
		assert.Equal(t, "https://example.com/schema", eventFromPb.Context.GetDataSchema())
		assert.Equal(t, "application/protobuf", eventFromPb.Context.GetDataContentType())
	})

	t.Run("conversion preserves extensions", func(t *testing.T) {
		// Create event which should have recordedtime extension
		event, err := NewEvent("test-domain", "test.event.type", protoBytes, nil)
		require.NoError(t, err)

		// Convert to proto
		eventPb, err := EventToProto(event)
		require.NoError(t, err)
		assert.NotNil(t, eventPb)

		// Verify extensions are present
		assert.NotNil(t, eventPb.Attributes)

		// Check for recordedtime extension
		recordedTimeAttr, exists := eventPb.Attributes["recordedtime"]
		assert.True(t, exists, "recordedtime extension should be present")
		assert.NotNil(t, recordedTimeAttr)
	})

	t.Run("conversion with empty data", func(t *testing.T) {
		// Create event with empty data
		event, err := NewEvent("test-domain", "test.event.type", []byte{}, nil)
		require.NoError(t, err)

		// Convert to proto
		eventPb, err := EventToProto(event)
		require.NoError(t, err)
		assert.NotNil(t, eventPb)

		// Verify basic fields are still set
		assert.Equal(t, "test-domain", eventPb.Source)
		assert.Equal(t, "test.event.type", eventPb.Type)
		assert.NotEmpty(t, eventPb.Id)
	})
}

func TestProtoToEvent(t *testing.T) {
	// Create a test protobuf message
	testProto := pb.PingResponse{Message: "test message for proto conversion"}
	protoBytes, err := proto.Marshal(&testProto)
	require.NoError(t, err)

	t.Run("successful conversion from protobuf", func(t *testing.T) {
		// First create a CloudEvent and convert to proto
		originalEvent, err := NewEvent("test-domain", "test.event.type", protoBytes, nil)
		require.NoError(t, err)

		// Convert to proto
		eventPb, err := EventToProto(originalEvent)
		require.NoError(t, err)
		assert.NotNil(t, eventPb)

		// Now test ProtoToEvent conversion back
		convertedEvent, err := ProtoToEvent(eventPb)
		require.NoError(t, err)
		assert.NotNil(t, convertedEvent)

		// Verify the converted event has the expected fields
		assert.Equal(t, "test-domain", convertedEvent.Source())
		assert.Equal(t, "test.event.type", convertedEvent.Type())
		assert.NotEmpty(t, convertedEvent.ID())
		assert.NotEmpty(t, convertedEvent.Data())

		// Verify the data can be unmarshaled back to the original proto
		var resultProto pb.PingResponse
		err = proto.Unmarshal(convertedEvent.Data(), &resultProto)
		require.NoError(t, err)
		assert.Equal(t, testProto.Message, resultProto.Message)
	})

	t.Run("conversion preserves all attributes", func(t *testing.T) {
		// Create event with custom attributes
		attributes := map[string]any{
			"subject":    "test-subject",
			"dataschema": "https://example.com/schema",
		}

		originalEvent, err := NewEvent("test-domain", "test.event.type", protoBytes, attributes)
		require.NoError(t, err)

		// Convert to proto and back
		eventPb, err := EventToProto(originalEvent)
		require.NoError(t, err)

		convertedEvent, err := ProtoToEvent(eventPb)
		require.NoError(t, err)
		assert.NotNil(t, convertedEvent)

		// Verify all attributes were preserved
		assert.Equal(t, "test-subject", convertedEvent.Subject())
		assert.Equal(t, "https://example.com/schema", convertedEvent.DataSchema())
		assert.Equal(t, "application/protobuf", convertedEvent.DataContentType())

		// Verify extensions are preserved
		recordedTime, ok := convertedEvent.Extensions()["recordedtime"]
		assert.True(t, ok, "recordedtime extension should be preserved")
		assert.NotNil(t, recordedTime)
	})

	t.Run("conversion with nil protobuf event", func(t *testing.T) {
		// Test with nil input
		convertedEvent, err := ProtoToEvent(nil)
		require.Error(t, err)
		assert.Equal(t, CloudEvent{}, convertedEvent)
		assert.Contains(t, err.Error(), "could not convert proto to event")
	})

	t.Run("roundtrip conversion preserves data integrity", func(t *testing.T) {
		// Create an event with complex data
		complexProto := &pb.PublishResponse{
			Results: []*pb.PublishResult{
				{EventId: "event-1"},
				{EventId: "event-2"},
				{EventId: "event-3"},
			},
		}
		complexBytes, err := proto.Marshal(complexProto)
		require.NoError(t, err)

		// Create original event
		originalEvent, err := NewEvent("complex-domain", "complex.event.type", complexBytes, map[string]any{
			"subject": "complex-subject",
		})
		require.NoError(t, err)

		// Convert to proto and back
		eventPb, err := EventToProto(originalEvent)
		require.NoError(t, err)

		convertedEvent, err := ProtoToEvent(eventPb)
		require.NoError(t, err)

		// Verify data integrity
		var resultProto pb.PublishResponse
		err = proto.Unmarshal(convertedEvent.Data(), &resultProto)
		require.NoError(t, err)
		assert.Len(t, resultProto.Results, 3)
		assert.Equal(t, "event-1", resultProto.Results[0].EventId)
		assert.Equal(t, "event-2", resultProto.Results[1].EventId)
		assert.Equal(t, "event-3", resultProto.Results[2].EventId)

		// Verify metadata is preserved
		assert.Equal(t, "complex-domain", convertedEvent.Source())
		assert.Equal(t, "complex.event.type", convertedEvent.Type())
		assert.Equal(t, "complex-subject", convertedEvent.Subject())
	})

	t.Run("conversion with empty data", func(t *testing.T) {
		// Create event with empty data
		originalEvent, err := NewEvent("empty-domain", "empty.event.type", []byte{}, nil)
		require.NoError(t, err)

		// Convert to proto and back
		eventPb, err := EventToProto(originalEvent)
		require.NoError(t, err)

		convertedEvent, err := ProtoToEvent(eventPb)
		require.NoError(t, err)
		assert.NotNil(t, convertedEvent)

		// Verify basic fields are preserved
		assert.Equal(t, "empty-domain", convertedEvent.Source())
		assert.Equal(t, "empty.event.type", convertedEvent.Type())
		assert.NotEmpty(t, convertedEvent.ID())
	})
}

func TestEventsToBatch(t *testing.T) {
	// Create test protobuf messages
	testProto1 := pb.PingResponse{Message: "test message 1"}
	testProto2 := pb.PingResponse{Message: "test message 2"}
	protoBytes1, err := proto.Marshal(&testProto1)
	require.NoError(t, err)
	protoBytes2, err := proto.Marshal(&testProto2)
	require.NoError(t, err)

	t.Run("successful batch conversion with multiple events", func(t *testing.T) {
		// Create multiple CloudEvents
		event1, err := NewEvent("domain1", "type1", protoBytes1, nil)
		require.NoError(t, err)

		event2, err := NewEvent("domain2", "type2", protoBytes2, map[string]any{
			"subject": "test-subject",
		})
		require.NoError(t, err)

		events := []CloudEvent{event1, event2}

		// Convert to batch
		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Len(t, batch.Events, 2)

		// Verify first event in batch
		assert.Equal(t, "domain1", batch.Events[0].Source)
		assert.Equal(t, "type1", batch.Events[0].Type)
		assert.NotEmpty(t, batch.Events[0].Id)

		// Verify second event in batch
		assert.Equal(t, "domain2", batch.Events[1].Source)
		assert.Equal(t, "type2", batch.Events[1].Type)
		assert.NotEmpty(t, batch.Events[1].Id)
		assert.NotNil(t, batch.Events[1].Attributes["subject"])
	})

	t.Run("batch conversion with single event", func(t *testing.T) {
		event, err := NewEvent("single-domain", "single-type", protoBytes1, nil)
		require.NoError(t, err)

		events := []CloudEvent{event}

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Len(t, batch.Events, 1)

		// Verify the single event
		assert.Equal(t, "single-domain", batch.Events[0].Source)
		assert.Equal(t, "single-type", batch.Events[0].Type)
		assert.NotEmpty(t, batch.Events[0].Id)
	})

	t.Run("batch conversion with empty slice", func(t *testing.T) {
		events := []CloudEvent{}

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Empty(t, batch.Events)
	})

	t.Run("batch conversion preserves event data", func(t *testing.T) {
		// Create events with different protobuf payloads
		complexProto := &pb.PublishResponse{
			Results: []*pb.PublishResult{
				{EventId: "batch-event-1"},
				{EventId: "batch-event-2"},
			},
		}
		complexBytes, err := proto.Marshal(complexProto)
		require.NoError(t, err)

		event1, err := NewEvent("data-domain1", "data-type1", protoBytes1, nil)
		require.NoError(t, err)

		event2, err := NewEvent("data-domain2", "data-type2", complexBytes, nil)
		require.NoError(t, err)

		events := []CloudEvent{event1, event2}

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Len(t, batch.Events, 2)

		// Convert protobuf events back to CloudEvents to verify data
		convertedEvent1, err := ProtoToEvent(batch.Events[0])
		require.NoError(t, err)

		convertedEvent2, err := ProtoToEvent(batch.Events[1])
		require.NoError(t, err)

		// Verify first event data can be unmarshaled
		var resultProto1 pb.PingResponse
		err = proto.Unmarshal(convertedEvent1.Data(), &resultProto1)
		require.NoError(t, err)
		assert.Equal(t, testProto1.Message, resultProto1.Message)

		// Verify second event data can be unmarshaled
		var resultProto2 pb.PublishResponse
		err = proto.Unmarshal(convertedEvent2.Data(), &resultProto2)
		require.NoError(t, err)
		assert.Len(t, resultProto2.Results, 2)
		assert.Equal(t, "batch-event-1", resultProto2.Results[0].EventId)
		assert.Equal(t, "batch-event-2", resultProto2.Results[1].EventId)
	})

	t.Run("batch conversion with nil slice", func(t *testing.T) {
		var events []CloudEvent

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Empty(t, batch.Events)
	})

	t.Run("batch conversion preserves all event attributes", func(t *testing.T) {
		// Create events with various attributes
		attributes1 := map[string]any{
			"subject":    "subject1",
			"dataschema": "https://schema1.example.com",
		}

		attributes2 := map[string]any{
			"subject":    "subject2",
			"dataschema": "https://schema2.example.com",
		}

		event1, err := NewEvent("attr-domain1", "attr-type1", protoBytes1, attributes1)
		require.NoError(t, err)

		event2, err := NewEvent("attr-domain2", "attr-type2", protoBytes2, attributes2)
		require.NoError(t, err)

		events := []CloudEvent{event1, event2}

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Len(t, batch.Events, 2)

		// Verify attributes are preserved in batch
		assert.NotNil(t, batch.Events[0].Attributes["subject"])
		assert.NotNil(t, batch.Events[0].Attributes["dataschema"])
		assert.NotNil(t, batch.Events[1].Attributes["subject"])
		assert.NotNil(t, batch.Events[1].Attributes["dataschema"])

		// Verify extensions like recordedtime are preserved
		assert.NotNil(t, batch.Events[0].Attributes["recordedtime"])
		assert.NotNil(t, batch.Events[1].Attributes["recordedtime"])
	})

	t.Run("large batch conversion", func(t *testing.T) {
		// Test with a larger number of events
		const numEvents = 100
		events := make([]CloudEvent, numEvents)

		for i := range numEvents {
			event, err := NewEvent(
				fmt.Sprintf("domain-%d", i),
				fmt.Sprintf("type-%d", i),
				protoBytes1,
				map[string]any{
					"subject": fmt.Sprintf("subject-%d", i),
				},
			)
			require.NoError(t, err)
			events[i] = event
		}

		batch, err := EventsToBatch(events)
		require.NoError(t, err)
		assert.NotNil(t, batch)
		assert.Len(t, batch.Events, numEvents)

		// Verify a few random events in the batch
		assert.Equal(t, "domain-0", batch.Events[0].Source)
		assert.Equal(t, "type-0", batch.Events[0].Type)
		assert.Equal(t, "domain-50", batch.Events[50].Source)
		assert.Equal(t, "type-50", batch.Events[50].Type)
		assert.Equal(t, "domain-99", batch.Events[99].Source)
		assert.Equal(t, "type-99", batch.Events[99].Type)
	})
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

	t.Run("WithBasicAuth", func(t *testing.T) {
		config := defaultCfg
		WithBasicAuth("user", "pass")(&config)
		assert.NotNil(t, config.perRPCCredentials)
	})

	t.Run("WithTokenAuth", func(t *testing.T) {
		mockProvider := &mockHeaderProvider{
			headers: map[string]string{"Authorization": "Bearer token"},
		}
		config := defaultCfg
		WithTokenAuth(mockProvider)(&config)
		assert.NotNil(t, config.perRPCCredentials)
	})

	t.Run("WithNopLookup", func(t *testing.T) {
		config := defaultCfg
		WithNOPLookup()(&config)
		assert.NotNil(t, config.nopInfoHeaderProvider)

		headers, err := config.nopInfoHeaderProvider.Headers(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "true", headers["x-include-nop-info"])
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
	mockInvoker := func(ctx context.Context, method string, req, reply any, cc *gp.ClientConn, opts ...gp.CallOption) error {
		capturedCtx = ctx
		return nil
	}

	// Get the interceptor from the function in client.go
	interceptor := newHeaderInterceptor(mockProvider)

	// Call the interceptor
	err := interceptor(t.Context(), "testMethod", nil, nil, nil, mockInvoker)
	require.NoError(t, err)

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

func (m *mockHeaderProvider) Headers(ctx context.Context) (map[string]string, error) {
	return m.headers, nil
}

// capturingServer is a minimal ChipIngressServer that records the incoming gRPC metadata
// of the last request it handles.
type capturingServer struct {
	pb.UnimplementedChipIngressServer
	lastMD metadata.MD
}

func (s *capturingServer) Ping(ctx context.Context, _ *pb.EmptyRequest) (*pb.PingResponse, error) {
	s.lastMD, _ = metadata.FromIncomingContext(ctx)
	return &pb.PingResponse{}, nil
}

// TestClient_ChainedHeaderProviders is a regression test for the fix that switched from
// grpc.WithUnaryInterceptor (last-one-wins) to grpc.WithChainUnaryInterceptor: when both
// WithHeaderProvider and WithNOPLookup are configured, both providers' headers must reach
// the server, not just the one registered last.
func TestClient_ChainedHeaderProviders(t *testing.T) {
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer lis.Close()

	srv := gp.NewServer()
	capture := &capturingServer{}
	pb.RegisterChipIngressServer(srv, capture)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	client, err := NewClient(lis.Addr().String(),
		WithInsecureConnection(),
		WithHeaderProvider(&mockHeaderProvider{headers: map[string]string{"x-resource-attr": "chain-1"}}),
		WithNOPLookup(),
	)
	require.NoError(t, err)
	defer client.Close() //nolint:errcheck

	_, err = client.Ping(t.Context(), &EmptyRequest{})
	require.NoError(t, err)

	require.NotNil(t, capture.lastMD)
	assert.Equal(t, []string{"chain-1"}, capture.lastMD.Get("x-resource-attr"))
	assert.Equal(t, []string{"true"}, capture.lastMD.Get("x-include-nop-info"))
}

func TestWithTLS(t *testing.T) {
	serverName := "example.com"
	config := defaultCfg
	config.host = serverName // Set host for SNI
	WithTLS()(&config)
	assert.NotNil(t, config.transportCredentials)
	// Verify it's TLS credentials (we can't easily inspect the internal config)
	assert.IsType(t, credentials.NewTLS(nil), config.transportCredentials)
}

func TestWithInsecureConnection(t *testing.T) {
	config := defaultCfg
	WithInsecureConnection()(&config)
	assert.Equal(t, insecure.NewCredentials(), config.transportCredentials)
}

func TestNewClientWithTLS(t *testing.T) {
	// This test verifies the option is applied, but doesn't test actual connection
	// since we'd need a real gRPC server for that
	client, err := NewClient(
		"example.com:443",
		WithTLS(),
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

func TestClient_RegisterSchemas(t *testing.T) {
	t.Run("successfully registers schemas", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		ctx := t.Context()
		mockClient.EXPECT().RegisterSchema(
			ctx,
			&pb.RegisterSchemaRequest{
				Schemas: []*pb.Schema{
					{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
					{Subject: "schema2", Schema: `{"type":"record","name":"Test2","fields":[{"name":"field2"}]}`, Format: 2},
				},
			},
		).Return(&pb.RegisterSchemaResponse{
			Registered: []*pb.RegisteredSchema{
				{Subject: "schema1", Version: 1},
				{Subject: "schema2", Version: 2},
			},
		}, nil)

		client := &client{
			client: mockClient,
			conn:   nil,
		}

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
			{Subject: "schema2", Schema: `{"type":"record","name":"Test2","fields":[{"name":"field2"}]}`, Format: 2},
		}

		result, err := client.RegisterSchemas(ctx, schemas...)
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"schema1": 1, "schema2": 2}, result)
	})

	t.Run("returns error when registration fails", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		ctx := t.Context()
		mockClient.EXPECT().RegisterSchema(
			ctx,
			&pb.RegisterSchemaRequest{
				Schemas: []*pb.Schema{
					{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
				},
			},
		).Return(nil, fmt.Errorf("registration failed"))

		client := &client{
			client: mockClient,
			conn:   nil,
		}

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
		}

		result, err := client.RegisterSchemas(ctx, schemas...)
		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to register schema: registration failed")
	})
}

func TestTransactionEnabledProtoTypes(t *testing.T) {
	t.Run("PublishResult roundtrip with Error", func(t *testing.T) {
		original := &pb.PublishResult{
			EventId: "evt-123",
			Error: &pb.PublishError{
				ErrorCode: pb.PublishErrorCode_PUBLISH_ERROR_CODE_SCHEMA_MISSING,
				Reason:    "subject test-domain-test.Entity not found",
			},
		}
		data, err := proto.Marshal(original)
		require.NoError(t, err)

		decoded := &pb.PublishResult{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)

		assert.Equal(t, "evt-123", decoded.EventId)
		require.NotNil(t, decoded.Error)
		assert.Equal(t, pb.PublishErrorCode_PUBLISH_ERROR_CODE_SCHEMA_MISSING, decoded.Error.ErrorCode)
		assert.Equal(t, "subject test-domain-test.Entity not found", decoded.Error.Reason)
	})

	t.Run("PublishResult success has nil Error", func(t *testing.T) {
		original := &pb.PublishResult{EventId: "evt-ok"}
		data, err := proto.Marshal(original)
		require.NoError(t, err)

		decoded := &pb.PublishResult{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)

		assert.Equal(t, "evt-ok", decoded.EventId)
		assert.Nil(t, decoded.Error)
	})

	t.Run("CloudEventBatch with PublishOptions roundtrip", func(t *testing.T) {
		txn := true
		batch := &pb.CloudEventBatch{
			Events:  []*cepb.CloudEvent{{Id: "e1", Source: "src", SpecVersion: "1.0", Type: "t"}},
			Options: &pb.PublishOptions{TransactionEnabled: &txn},
		}
		data, err := proto.Marshal(batch)
		require.NoError(t, err)

		decoded := &pb.CloudEventBatch{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)

		require.NotNil(t, decoded.Options)
		assert.True(t, decoded.Options.GetTransactionEnabled())
		assert.Len(t, decoded.Events, 1)
		assert.Equal(t, "e1", decoded.Events[0].Id)
	})

	t.Run("GetTransactionEnabled edge cases", func(t *testing.T) {
		// nil PublishOptions
		var nilOpts *pb.PublishOptions
		assert.False(t, nilOpts.GetTransactionEnabled())

		// TransactionEnabled unset (nil pointer)
		emptyOpts := &pb.PublishOptions{}
		assert.False(t, emptyOpts.GetTransactionEnabled())

		// Explicitly true
		trueVal := true
		assert.True(t, (&pb.PublishOptions{TransactionEnabled: &trueVal}).GetTransactionEnabled())

		// Explicitly false
		falseVal := false
		assert.False(t, (&pb.PublishOptions{TransactionEnabled: &falseVal}).GetTransactionEnabled())
	})

	t.Run("EventsToBatchWithOpts with WithTransactionEnabled(false) emits PublishOptions{TransactionEnabled: false}", func(t *testing.T) {
		event, err := NewEvent("domain", "entity", []byte("data"), nil)
		require.NoError(t, err)

		batch, err := EventsToBatchWithOpts([]CloudEvent{event}, WithTransactionEnabled(false))
		require.NoError(t, err)
		require.NotNil(t, batch.Options, "client always emits PublishOptions to make intent explicit on the wire")
		require.NotNil(t, batch.Options.TransactionEnabled)
		assert.False(t, batch.Options.GetTransactionEnabled())
	})

	t.Run("EventsToBatchWithOpts with WithTransactionEnabled(true) emits PublishOptions{TransactionEnabled: true}", func(t *testing.T) {
		event, err := NewEvent("domain", "entity", []byte("data"), nil)
		require.NoError(t, err)

		batch, err := EventsToBatchWithOpts([]CloudEvent{event}, WithTransactionEnabled(true))
		require.NoError(t, err)
		require.NotNil(t, batch.Options)
		require.NotNil(t, batch.Options.TransactionEnabled)
		assert.True(t, batch.Options.GetTransactionEnabled())
	})

	t.Run("EventsToBatch without options defaults to PublishOptions{TransactionEnabled: false}", func(t *testing.T) {
		event, err := NewEvent("domain", "entity", []byte("data"), nil)
		require.NoError(t, err)

		batch, err := EventsToBatch([]CloudEvent{event})
		require.NoError(t, err)
		require.NotNil(t, batch.Options, "client always emits PublishOptions to make intent explicit on the wire")
		require.NotNil(t, batch.Options.TransactionEnabled)
		assert.False(t, batch.Options.GetTransactionEnabled())
	})
}
