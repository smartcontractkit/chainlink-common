package eventstore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var (
	testTime1 = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	testTime2 = time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)
)

func TestClient_InsertAndList(t *testing.T) {
	ctx := t.Context()
	client := Client{grpc: newTestGRPCClient()}

	ev := capabilities.PendingEvent{
		TriggerId: "trigger-1",
		EventId:   "event-1",
		Payload:   []byte("payload-data"),
		FirstAt:   testTime1,
		Attempts:  0,
	}

	err := client.Insert(ctx, ev)
	require.NoError(t, err)

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "trigger-1", events[0].TriggerId)
	assert.Equal(t, "event-1", events[0].EventId)
	assert.Equal(t, []byte("payload-data"), events[0].Payload)
	assert.Equal(t, testTime1, events[0].FirstAt)
	assert.True(t, events[0].LastSentAt.IsZero())
}

func TestClient_InsertWithLastSentAt(t *testing.T) {
	ctx := t.Context()
	mock := newTestGRPCClient()
	client := Client{grpc: mock}

	ev := capabilities.PendingEvent{
		TriggerId:  "trigger-1",
		EventId:    "event-1",
		Payload:    []byte("data"),
		FirstAt:    testTime1,
		LastSentAt: testTime2,
		Attempts:   3,
	}

	require.NoError(t, client.Insert(ctx, ev))

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, testTime2, events[0].LastSentAt)
	assert.Equal(t, 3, events[0].Attempts)
}

func TestClient_UpdateDelivery(t *testing.T) {
	ctx := t.Context()
	mock := newTestGRPCClient()
	client := Client{grpc: mock}

	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-1",
		EventId:   "event-1",
		Payload:   []byte("data"),
		FirstAt:   testTime1,
	}))

	err := client.UpdateDelivery(ctx, "trigger-1", "event-1", testTime2, 5)
	require.NoError(t, err)

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, testTime2, events[0].LastSentAt)
	assert.Equal(t, 5, events[0].Attempts)
}

func TestClient_DeleteEvent(t *testing.T) {
	ctx := t.Context()
	client := Client{grpc: newTestGRPCClient()}

	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-1", EventId: "event-1", Payload: []byte("a"), FirstAt: testTime1,
	}))
	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-1", EventId: "event-2", Payload: []byte("b"), FirstAt: testTime1,
	}))

	err := client.DeleteEvent(ctx, "trigger-1", "event-1")
	require.NoError(t, err)

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "event-2", events[0].EventId)
}

func TestClient_DeleteEventsForTrigger(t *testing.T) {
	ctx := t.Context()
	client := Client{grpc: newTestGRPCClient()}

	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-1", EventId: "event-1", Payload: []byte("a"), FirstAt: testTime1,
	}))
	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-1", EventId: "event-2", Payload: []byte("b"), FirstAt: testTime1,
	}))
	require.NoError(t, client.Insert(ctx, capabilities.PendingEvent{
		TriggerId: "trigger-2", EventId: "event-3", Payload: []byte("c"), FirstAt: testTime1,
	}))

	err := client.DeleteEventsForTrigger(ctx, "trigger-1")
	require.NoError(t, err)

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "trigger-2", events[0].TriggerId)
}

func TestServer_InsertAndList(t *testing.T) {
	ctx := t.Context()
	server := Server{impl: capabilities.NewMemEventStore()}

	_, err := server.Insert(ctx, &pb.InsertEventRequest{
		Event: &pb.PendingEventProto{
			TriggerId: "trigger-1",
			EventId:   "event-1",
			Payload:   []byte("payload-data"),
			FirstAt:   timestamppb.New(testTime1),
			Attempts:  0,
		},
	})
	require.NoError(t, err)

	resp, err := server.List(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Events, 1)
	assert.Equal(t, "trigger-1", resp.Events[0].TriggerId)
	assert.Equal(t, "event-1", resp.Events[0].EventId)
	assert.Equal(t, []byte("payload-data"), resp.Events[0].Payload)
	assert.Equal(t, testTime1, resp.Events[0].FirstAt.AsTime())
}

func TestServer_UpdateDelivery(t *testing.T) {
	ctx := t.Context()
	server := Server{impl: capabilities.NewMemEventStore()}

	_, err := server.Insert(ctx, &pb.InsertEventRequest{
		Event: &pb.PendingEventProto{
			TriggerId: "trigger-1",
			EventId:   "event-1",
			Payload:   []byte("data"),
			FirstAt:   timestamppb.New(testTime1),
		},
	})
	require.NoError(t, err)

	_, err = server.UpdateDelivery(ctx, &pb.UpdateDeliveryRequest{
		TriggerId:  "trigger-1",
		EventId:    "event-1",
		LastSentAt: timestamppb.New(testTime2),
		Attempts:   3,
	})
	require.NoError(t, err)

	resp, err := server.List(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Events, 1)
	assert.Equal(t, testTime2, resp.Events[0].LastSentAt.AsTime())
	assert.Equal(t, int32(3), resp.Events[0].Attempts)
}

func TestServer_DeleteEvent(t *testing.T) {
	ctx := t.Context()
	server := Server{impl: capabilities.NewMemEventStore()}

	_, err := server.Insert(ctx, &pb.InsertEventRequest{
		Event: &pb.PendingEventProto{
			TriggerId: "trigger-1", EventId: "event-1", Payload: []byte("a"), FirstAt: timestamppb.New(testTime1),
		},
	})
	require.NoError(t, err)
	_, err = server.Insert(ctx, &pb.InsertEventRequest{
		Event: &pb.PendingEventProto{
			TriggerId: "trigger-1", EventId: "event-2", Payload: []byte("b"), FirstAt: timestamppb.New(testTime1),
		},
	})
	require.NoError(t, err)

	_, err = server.DeleteEvent(ctx, &pb.DeleteEventRequest{TriggerId: "trigger-1", EventId: "event-1"})
	require.NoError(t, err)

	resp, err := server.List(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Events, 1)
	assert.Equal(t, "event-2", resp.Events[0].EventId)
}

func TestServer_DeleteEventsForTrigger(t *testing.T) {
	ctx := t.Context()
	server := Server{impl: capabilities.NewMemEventStore()}

	for _, ev := range []struct{ tid, eid string }{
		{"trigger-1", "event-1"},
		{"trigger-1", "event-2"},
		{"trigger-2", "event-3"},
	} {
		_, err := server.Insert(ctx, &pb.InsertEventRequest{
			Event: &pb.PendingEventProto{
				TriggerId: ev.tid, EventId: ev.eid, Payload: []byte("x"), FirstAt: timestamppb.New(testTime1),
			},
		})
		require.NoError(t, err)
	}

	_, err := server.DeleteEventsForTrigger(ctx, &pb.DeleteEventsForTriggerRequest{TriggerId: "trigger-1"})
	require.NoError(t, err)

	resp, err := server.List(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Events, 1)
	assert.Equal(t, "trigger-2", resp.Events[0].TriggerId)
}

func TestServer_NilImpl(t *testing.T) {
	ctx := t.Context()
	server := Server{}

	_, err := server.Insert(ctx, &pb.InsertEventRequest{Event: &pb.PendingEventProto{}})
	assert.Error(t, err)

	_, err = server.UpdateDelivery(ctx, &pb.UpdateDeliveryRequest{})
	assert.Error(t, err)

	_, err = server.List(ctx, &emptypb.Empty{})
	assert.Error(t, err)

	_, err = server.DeleteEvent(ctx, &pb.DeleteEventRequest{})
	assert.Error(t, err)

	_, err = server.DeleteEventsForTrigger(ctx, &pb.DeleteEventsForTriggerRequest{})
	assert.Error(t, err)
}

func TestRoundTrip_ClientThroughServer(t *testing.T) {
	ctx := t.Context()
	store := capabilities.NewMemEventStore()
	server := NewServer(store)
	client := Client{grpc: &serverBackedGRPCClient{server: server}}

	ev := capabilities.PendingEvent{
		TriggerId:  "trigger-rt",
		EventId:    "event-rt",
		Payload:    []byte("round-trip-payload"),
		FirstAt:    testTime1,
		LastSentAt: testTime2,
		Attempts:   2,
	}

	require.NoError(t, client.Insert(ctx, ev))

	events, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, ev.TriggerId, events[0].TriggerId)
	assert.Equal(t, ev.EventId, events[0].EventId)
	assert.Equal(t, ev.Payload, events[0].Payload)
	assert.Equal(t, ev.FirstAt, events[0].FirstAt)
	assert.Equal(t, ev.LastSentAt, events[0].LastSentAt)
	assert.Equal(t, ev.Attempts, events[0].Attempts)

	newTime := testTime2.Add(10 * time.Minute)
	require.NoError(t, client.UpdateDelivery(ctx, "trigger-rt", "event-rt", newTime, 3))

	events, err = client.List(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, newTime, events[0].LastSentAt)
	assert.Equal(t, 3, events[0].Attempts)

	require.NoError(t, client.DeleteEvent(ctx, "trigger-rt", "event-rt"))

	events, err = client.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, events)
}

// testGRPCClient is a mock pb.EventStoreClient backed by an in-memory store.
type testGRPCClient struct {
	mu     sync.Mutex
	events map[string]map[string]*pb.PendingEventProto // triggerID -> eventID -> event
}

func newTestGRPCClient() *testGRPCClient {
	return &testGRPCClient{events: make(map[string]map[string]*pb.PendingEventProto)}
}

func (t *testGRPCClient) Insert(_ context.Context, in *pb.InsertEventRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	ev := in.GetEvent()
	if t.events[ev.TriggerId] == nil {
		t.events[ev.TriggerId] = make(map[string]*pb.PendingEventProto)
	}
	t.events[ev.TriggerId][ev.EventId] = ev
	return &emptypb.Empty{}, nil
}

func (t *testGRPCClient) UpdateDelivery(_ context.Context, in *pb.UpdateDeliveryRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if trigger, ok := t.events[in.TriggerId]; ok {
		if ev, ok := trigger[in.EventId]; ok {
			ev.LastSentAt = in.LastSentAt
			ev.Attempts = in.Attempts
		}
	}
	return &emptypb.Empty{}, nil
}

func (t *testGRPCClient) List(_ context.Context, _ *emptypb.Empty, _ ...grpc.CallOption) (*pb.ListEventsResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var out []*pb.PendingEventProto
	for _, trigger := range t.events {
		for _, ev := range trigger {
			out = append(out, ev)
		}
	}
	return &pb.ListEventsResponse{Events: out}, nil
}

func (t *testGRPCClient) DeleteEvent(_ context.Context, in *pb.DeleteEventRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if trigger, ok := t.events[in.TriggerId]; ok {
		delete(trigger, in.EventId)
		if len(trigger) == 0 {
			delete(t.events, in.TriggerId)
		}
	}
	return &emptypb.Empty{}, nil
}

func (t *testGRPCClient) DeleteEventsForTrigger(_ context.Context, in *pb.DeleteEventsForTriggerRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.events, in.TriggerId)
	return &emptypb.Empty{}, nil
}

// serverBackedGRPCClient implements pb.EventStoreClient by calling the Server directly,
// simulating the full serialization round-trip without a real gRPC connection.
type serverBackedGRPCClient struct {
	server *Server
}

func (s *serverBackedGRPCClient) Insert(ctx context.Context, in *pb.InsertEventRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	return s.server.Insert(ctx, in)
}

func (s *serverBackedGRPCClient) UpdateDelivery(ctx context.Context, in *pb.UpdateDeliveryRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	return s.server.UpdateDelivery(ctx, in)
}

func (s *serverBackedGRPCClient) List(ctx context.Context, in *emptypb.Empty, _ ...grpc.CallOption) (*pb.ListEventsResponse, error) {
	return s.server.List(ctx, in)
}

func (s *serverBackedGRPCClient) DeleteEvent(ctx context.Context, in *pb.DeleteEventRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	return s.server.DeleteEvent(ctx, in)
}

func (s *serverBackedGRPCClient) DeleteEventsForTrigger(ctx context.Context, in *pb.DeleteEventsForTriggerRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	return s.server.DeleteEventsForTrigger(ctx, in)
}
