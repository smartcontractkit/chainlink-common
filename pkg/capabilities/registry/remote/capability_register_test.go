package remote

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

// fakeTriggerGRPC is the gRPC side of a trigger capability client.
type fakeTriggerGRPC struct {
	pb.TriggerExecutableClient
	// open stands for the connection: it returns when the stream is open.
	open func(ctx context.Context) error
	// messages is what the server sends on the stream.
	messages chan *pb.TriggerResponseMessage
	// streamCtx receives the context each stream was opened with.
	streamCtx chan context.Context
}

func (f *fakeTriggerGRPC) RegisterTrigger(ctx context.Context, _ *pb.TriggerRegistrationRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[pb.TriggerResponseMessage], error) {
	if err := f.open(ctx); err != nil {
		return nil, err
	}
	f.streamCtx <- ctx
	return &fakeTriggerStream{ctx: ctx, messages: f.messages}, nil
}

type fakeTriggerStream struct {
	grpc.ClientStream
	ctx      context.Context //nolint:containedctx // a stream keeps the context it was opened with
	messages chan *pb.TriggerResponseMessage
}

func (s *fakeTriggerStream) Recv() (*pb.TriggerResponseMessage, error) {
	select {
	case m := <-s.messages:
		return m, nil
	case <-s.ctx.Done():
		return nil, context.Cause(s.ctx)
	}
}

func newFakeTriggerClient(t *testing.T, open func(context.Context) error) (*triggerExecutableClient, *fakeTriggerGRPC) {
	f := &fakeTriggerGRPC{
		open:      open,
		messages:  make(chan *pb.TriggerResponseMessage, 10),
		streamCtx: make(chan context.Context, 10),
	}
	return &triggerExecutableClient{grpc: f, lggr: logger.Test(t), cancelFuncs: map[string]func(){}}, f
}

// A trigger stream outlives the context of the call that registered it. The call itself
// must not: when the plugin is gone the registration is never acknowledged, and the call
// has to end with its caller's context.
func TestTriggerExecutableClient_RegisterTriggerEndsWithItsContextUntilAcknowledged(t *testing.T) {
	t.Parallel()
	req := capabilities.TriggerRegistrationRequest{TriggerID: "trigger-1"}

	type result struct {
		ch  <-chan capabilities.TriggerResponse
		err error
	}
	register := func(ctx context.Context, c *triggerExecutableClient) <-chan result {
		done := make(chan result, 1)
		go func() {
			ch, err := c.RegisterTrigger(ctx, req)
			done <- result{ch, err}
		}()
		return done
	}
	waitFor := func(t *testing.T, done <-chan result) result {
		t.Helper()
		select {
		case r := <-done:
			return r
		case <-time.After(tests.WaitTimeout(t)):
			require.FailNow(t, "RegisterTrigger did not return")
			return result{}
		}
	}

	t.Run("the connection cannot be made", func(t *testing.T) {
		t.Parallel()
		// As a LOOP client does: keep trying to connect for as long as the context allows.
		client, _ := newFakeTriggerClient(t, func(ctx context.Context) error {
			<-ctx.Done()
			return context.Cause(ctx)
		})
		ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
		defer cancel()
		r := waitFor(t, register(ctx, client))
		require.ErrorIs(t, r.err, context.DeadlineExceeded)
	})

	t.Run("a context that has ended already opens no stream", func(t *testing.T) {
		t.Parallel()
		opened := make(chan struct{}, 1)
		client, _ := newFakeTriggerClient(t, func(context.Context) error {
			opened <- struct{}{}
			return nil
		})
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		r := waitFor(t, register(ctx, client))
		require.ErrorIs(t, r.err, context.Canceled)
		require.Empty(t, opened, "a stream was opened")
	})

	t.Run("the registration is not acknowledged", func(t *testing.T) {
		t.Parallel()
		client, grpcSide := newFakeTriggerClient(t, func(context.Context) error { return nil })
		ctx, cancel := context.WithCancel(t.Context())
		done := register(ctx, client)
		streamCtx := <-grpcSide.streamCtx
		cancel()
		r := waitFor(t, done)
		require.ErrorIs(t, r.err, context.Canceled)
		require.Error(t, streamCtx.Err(), "the stream of a failed registration must be closed")
	})

	t.Run("an acknowledged registration outlives the context", func(t *testing.T) {
		t.Parallel()
		client, grpcSide := newFakeTriggerClient(t, func(context.Context) error { return nil })
		grpcSide.messages <- &pb.TriggerResponseMessage{Message: &pb.TriggerResponseMessage_Ack{Ack: &emptypb.Empty{}}}
		ctx, cancel := context.WithCancel(t.Context())
		r := waitFor(t, register(ctx, client))
		require.NoError(t, r.err)
		streamCtx := <-grpcSide.streamCtx

		cancel()
		grpcSide.messages <- &pb.TriggerResponseMessage{Message: &pb.TriggerResponseMessage_Response{
			Response: pb.TriggerResponseToProto(capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: "event-1"}}),
		}}
		select {
		case resp := <-r.ch:
			assert.Equal(t, "event-1", resp.Event.ID)
		case <-time.After(tests.WaitTimeout(t)):
			require.FailNow(t, "no event after the caller's context ended")
		}
		require.NoError(t, streamCtx.Err())

		// Unregistering is what closes the stream.
		grpcSide.TriggerExecutableClient = unregisterOK{}
		require.NoError(t, client.UnregisterTrigger(t.Context(), req))
		require.Error(t, streamCtx.Err())
	})
}

type unregisterOK struct{ pb.TriggerExecutableClient }

func (unregisterOK) UnregisterTrigger(context.Context, *pb.TriggerRegistrationRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
