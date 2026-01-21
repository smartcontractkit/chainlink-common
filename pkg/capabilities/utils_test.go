package capabilities_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

func TestFromValueOrAny(t *testing.T) {
	t.Run("from any", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "from-any"}
		a, err := anypb.New(msg)
		require.NoError(t, err)

		var out pb.TriggerEvent
		migrated, err := capabilities.FromValueOrAny(nil, a, &out)
		require.NoError(t, err)
		assert.True(t, migrated)
		assert.Equal(t, msg.Id, out.Id)
	})

	t.Run("from values", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "from-map"}
		wrapped, err := values.WrapMap(msg)
		require.NoError(t, err)

		var out pb.TriggerEvent
		migrated, err := capabilities.FromValueOrAny(wrapped, nil, &out)
		require.NoError(t, err)
		assert.False(t, migrated)
		assert.Equal(t, msg.Id, out.Id)
	})

	t.Run("with neither any nor map", func(t *testing.T) {
		var out pb.TriggerEvent
		_, err := capabilities.FromValueOrAny(nil, nil, &out)
		require.Error(t, err)
		require.ErrorIs(t, err, capabilities.ErrNeitherValueNorAny)
	})

	t.Run("with nil map", func(t *testing.T) {
		var out pb.TriggerEvent
		req := capabilities.TriggerRegistrationRequest{}
		_, err := capabilities.FromValueOrAny(req.Config, req.Payload, &out)
		require.Error(t, err)
		require.ErrorIs(t, err, capabilities.ErrNeitherValueNorAny)
	})

	t.Run("with nil any other values", func(t *testing.T) {
		var out pb.TriggerEvent
		_, err := capabilities.FromValueOrAny(new(values.Int64), nil, &out)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to transform value to proto")
	})

	t.Run("emptybp works", func(t *testing.T) {
		msg := &emptypb.Empty{}
		a, err := anypb.New(msg)
		require.NoError(t, err)

		var out emptypb.Empty
		migrated, err := capabilities.FromValueOrAny(nil, a, &out)
		require.NoError(t, err)
		assert.True(t, migrated)
	})
}

func TestUnwrapRequest(t *testing.T) {
	t.Run("with any", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "req-any"}
		a, err := anypb.New(msg)

		cfg := &pb.TriggerEvent{Id: "req-any-cfg"}
		require.NoError(t, err)
		ac, err := anypb.New(cfg)
		require.NoError(t, err)

		req := capabilities.CapabilityRequest{ConfigPayload: a, Payload: ac}
		var input pb.TriggerEvent
		var config pb.TriggerEvent
		migrated, err := capabilities.UnwrapRequest(req, &config, &input)
		require.NoError(t, err)
		assert.True(t, migrated)
		assert.Equal(t, "req-any", config.Id)
		assert.Equal(t, "req-any-cfg", input.Id)
	})

	t.Run("with values", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "req-map"}
		wrapped, err := values.WrapMap(msg)
		require.NoError(t, err)

		cfg := &pb.TriggerEvent{Id: "req-map-cfg"}
		cwrapped, err := values.WrapMap(cfg)
		require.NoError(t, err)

		req := capabilities.CapabilityRequest{Config: wrapped, Inputs: cwrapped}
		var input pb.TriggerEvent
		var config pb.TriggerEvent
		migrated, err := capabilities.UnwrapRequest(req, &config, &input)
		require.NoError(t, err)
		assert.False(t, migrated)
		assert.Equal(t, "req-map", config.Id)
		assert.Equal(t, "req-map-cfg", input.Id)
	})
}

func TestUnwrapResponse(t *testing.T) {
	t.Run("with any", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "resp-any"}
		a, err := anypb.New(msg)
		require.NoError(t, err)

		resp := capabilities.CapabilityResponse{Payload: a}
		var out pb.TriggerEvent
		migrated, err := capabilities.UnwrapResponse(resp, &out)
		require.NoError(t, err)
		assert.True(t, migrated)
		assert.Equal(t, "resp-any", out.Id)
	})

	t.Run("with values", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "resp-map"}
		wrapped, err := values.WrapMap(msg)
		require.NoError(t, err)

		resp := capabilities.CapabilityResponse{Value: wrapped}
		var out pb.TriggerEvent
		migrated, err := capabilities.UnwrapResponse(resp, &out)
		require.NoError(t, err)
		assert.False(t, migrated)
		assert.Equal(t, "resp-map", out.Id)
	})
}

func TestSetResponse(t *testing.T) {
	t.Run("set with any", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "val-any"}
		resp := capabilities.CapabilityResponse{}
		err := capabilities.SetResponse(&resp, true, msg)
		require.NoError(t, err)
		assert.NotNil(t, resp.Payload)
		assert.Nil(t, resp.Value)
	})

	t.Run("set with value", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "val-map"}
		resp := capabilities.CapabilityResponse{}
		err := capabilities.SetResponse(&resp, false, msg)
		require.NoError(t, err)
		assert.NotNil(t, resp.Value)
		assert.Nil(t, resp.Payload)
	})
}

func TestExecute(t *testing.T) {
	t.Run("with any", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "input"}
		a, err := anypb.New(msg)
		require.NoError(t, err)

		meteringNodeDetail := capabilities.MeteringNodeDetail{
			Peer2PeerID: "node-1",
			SpendUnit:   "1",
			SpendValue:  "0.00001",
		}
		req := capabilities.CapabilityRequest{ConfigPayload: a, Payload: a}

		resp, err := capabilities.Execute(context.Background(), req, &pb.TriggerEvent{}, &pb.TriggerEvent{},
			func(_ context.Context, _ capabilities.RequestMetadata, i, c *pb.TriggerEvent) (*pb.TriggerEvent, capabilities.ResponseMetadata, error) {
				return &pb.TriggerEvent{Id: "out"}, capabilities.ResponseMetadata{
					Metering: []capabilities.MeteringNodeDetail{
						meteringNodeDetail,
					},
				}, nil
			})
		require.NoError(t, err)
		assert.NotNil(t, resp.Payload)
		assert.Nil(t, resp.Value)
		assert.Len(t, resp.Metadata.Metering, 1)
		assert.Equal(t, meteringNodeDetail, resp.Metadata.Metering[0])
	})

	t.Run("with value", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "input"}
		wrapped, err := values.WrapMap(msg)
		require.NoError(t, err)

		meteringNodeDetail := capabilities.MeteringNodeDetail{
			Peer2PeerID: "node-1",
			SpendUnit:   "1",
			SpendValue:  "0.00001",
		}
		req := capabilities.CapabilityRequest{Inputs: wrapped, Config: wrapped}

		resp, err := capabilities.Execute(context.Background(), req, &pb.TriggerEvent{}, &pb.TriggerEvent{},
			func(_ context.Context, _ capabilities.RequestMetadata, i, c *pb.TriggerEvent) (*pb.TriggerEvent, capabilities.ResponseMetadata, error) {
				assert.Equal(t, "input", i.Id)
				assert.Equal(t, "input", c.Id)
				return &pb.TriggerEvent{Id: "out"}, capabilities.ResponseMetadata{
					Metering: []capabilities.MeteringNodeDetail{
						meteringNodeDetail,
					},
				}, nil
			})
		require.NoError(t, err)
		assert.NotNil(t, resp.Value)
		assert.Nil(t, resp.Payload)
		assert.NotNil(t, resp.Metadata)
		assert.Len(t, resp.Metadata.Metering, 1)
		assert.Equal(t, meteringNodeDetail, resp.Metadata.Metering[0])
	})
}

func TestRegisterTrigger(t *testing.T) {
	// Validate that if context is not canceled and stop remains open that
	// all sent events are received and transformed.
	t.Run("OK channel drained before context cancels", func(t *testing.T) {
		ctx := t.Context()
		stop := make(chan struct{})
		t.Cleanup(func() { close(stop) })

		a, err := anypb.New(&pb.TriggerEvent{Id: "reg"})
		require.NoError(t, err)
		req := capabilities.TriggerRegistrationRequest{
			Metadata: capabilities.RequestMetadata{WorkflowID: "workflow-id"},
			Payload:  a,
		}

		wantEvents := 50
		eventCh := make(chan capabilities.TriggerAndId[*pb.TriggerEvent])

		go func() {
			defer close(eventCh)
			for i := range wantEvents {
				select {
				case eventCh <- capabilities.TriggerAndId[*pb.TriggerEvent]{
					Trigger: &pb.TriggerEvent{Id: fmt.Sprintf("id-%d", i+1)},
					Id:      "trigger-id",
				}:
				case <-ctx.Done():
					return
				}
			}
		}()

		respCh, err := capabilities.RegisterTrigger(
			ctx,
			stop,
			"type",
			req,
			&pb.TriggerEvent{},
			func(_ context.Context, triggerID string, m capabilities.RequestMetadata, r *pb.TriggerEvent) (<-chan capabilities.TriggerAndId[*pb.TriggerEvent], caperrors.Error) {
				assert.Equal(t, "workflow-id", m.WorkflowID)
				assert.Equal(t, "reg", r.Id)
				return eventCh, nil
			},
		)
		require.NoError(t, err)

		gotEvents := 0
		for resp := range respCh {
			gotEvents++

			assert.Equal(t, "trigger-id", resp.Event.ID)
			assert.Equal(t, "type", resp.Event.TriggerType)

			var gotTrigger pb.TriggerEvent
			require.NoError(t, resp.Event.Payload.UnmarshalTo(&gotTrigger))
			require.Equal(t, fmt.Sprintf("id-%d", gotEvents), gotTrigger.Id)
		}
		require.Equal(t, wantEvents, gotEvents, fmt.Sprintf("expected %d events, got %d", wantEvents, gotEvents))
	})

	// Validate that if the original context is canceled, all events are sent
	// and transformed because the stop channel remains open.
	t.Run("OK context canceled stop channel open and source channel drained", func(t *testing.T) {
		// Cancel the context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stop := make(chan struct{})
		t.Cleanup(func() { close(stop) })

		wantEvents := 50

		a, err := anypb.New(&pb.TriggerEvent{Id: "reg"})
		require.NoError(t, err)
		req := capabilities.TriggerRegistrationRequest{
			Metadata: capabilities.RequestMetadata{WorkflowID: "workflow-id"},
			Payload:  a,
		}

		eventCh := make(chan capabilities.TriggerAndId[*pb.TriggerEvent])
		go func() {
			defer close(eventCh)
			for i := range wantEvents {
				select {
				case eventCh <- capabilities.TriggerAndId[*pb.TriggerEvent]{
					Trigger: &pb.TriggerEvent{Id: fmt.Sprintf("id-%d", i+1)},
					Id:      "trigger-id",
				}:
				case <-t.Context().Done():
					return
				}
			}
		}()

		respCh, err := capabilities.RegisterTrigger(
			ctx,
			stop,
			"type",
			req,
			&pb.TriggerEvent{},
			func(_ context.Context, triggerID string, m capabilities.RequestMetadata, r *pb.TriggerEvent) (<-chan capabilities.TriggerAndId[*pb.TriggerEvent], caperrors.Error) {
				assert.Equal(t, "workflow-id", m.WorkflowID)
				assert.Equal(t, "reg", r.Id)
				return eventCh, nil
			},
		)
		require.NoError(t, err)

		gotEvents := 0
		for resp := range respCh {
			gotEvents++

			assert.Equal(t, "trigger-id", resp.Event.ID)
			assert.Equal(t, "type", resp.Event.TriggerType)

			var gotTrigger pb.TriggerEvent
			require.NoError(t, resp.Event.Payload.UnmarshalTo(&gotTrigger))
			require.Equal(t, fmt.Sprintf("id-%d", gotEvents), gotTrigger.Id)
		}
		require.Equal(t, wantEvents, gotEvents, fmt.Sprintf("expected %d events, got %d", wantEvents, gotEvents))
	})

	// Validate that if the context is canceled while calling the passed function
	// that the error is returned.
	t.Run("NOK context cancels when calling fn", func(t *testing.T) {
		// Cancel the context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stop := make(chan struct{})
		t.Cleanup(func() { close(stop) })

		a, err := anypb.New(&pb.TriggerEvent{Id: "reg"})
		require.NoError(t, err)
		req := capabilities.TriggerRegistrationRequest{
			Metadata: capabilities.RequestMetadata{WorkflowID: "workflow-id"},
			Payload:  a,
		}

		_, err = capabilities.RegisterTrigger(
			ctx,
			stop,
			"type",
			req,
			&pb.TriggerEvent{},
			func(ctx context.Context, triggerID string, m capabilities.RequestMetadata, r *pb.TriggerEvent) (<-chan capabilities.TriggerAndId[*pb.TriggerEvent], caperrors.Error) {
				return nil, caperrors.NewPublicSystemError(ctx.Err(), caperrors.Internal)
			},
		)
		require.Error(t, err)
	})

	// Validate that closing the stop channel prevents all events from being
	// sent.
	t.Run("OK stop prevents draining channel", func(t *testing.T) {
		ctx := t.Context()

		sentEvents := 50
		wantEvents := 10
		stop := make(chan struct{})

		a, err := anypb.New(&pb.TriggerEvent{Id: "reg"})
		require.NoError(t, err)
		req := capabilities.TriggerRegistrationRequest{
			Metadata: capabilities.RequestMetadata{WorkflowID: "workflow-id"},
			Payload:  a,
		}

		eventCh := make(chan capabilities.TriggerAndId[*pb.TriggerEvent])

		go func() {
			defer close(eventCh)
			for i := range sentEvents {
				select {
				case eventCh <- capabilities.TriggerAndId[*pb.TriggerEvent]{
					Trigger: &pb.TriggerEvent{Id: fmt.Sprintf("id-%d", i+1)},
					Id:      "trigger-id",
				}:
				case <-t.Context().Done():
					return
				}
			}
		}()

		respCh, err := capabilities.RegisterTrigger(
			ctx,
			stop,
			"type",
			req,
			&pb.TriggerEvent{},
			func(ctx context.Context, triggerID string, m capabilities.RequestMetadata, r *pb.TriggerEvent) (<-chan capabilities.TriggerAndId[*pb.TriggerEvent], caperrors.Error) {
				assert.Equal(t, "workflow-id", m.WorkflowID)
				assert.Equal(t, "reg", r.Id)
				if ctx.Err() != nil {
					return nil, caperrors.NewPublicSystemError(ctx.Err(), caperrors.Internal)
				} else {
					return eventCh, nil
				}
			},
		)
		require.NoError(t, err)

		gotEvents := 0
		for resp := range respCh {
			gotEvents++

			// close the stop channel to cancel transforming
			if gotEvents == wantEvents {
				close(stop)
				require.Eventually(t, func() bool {
					_, isOpen := <-respCh
					return !isOpen
				}, 1*time.Second, 100*time.Millisecond)
				require.True(t, gotEvents < sentEvents)
				return
			}

			assert.Equal(t, "trigger-id", resp.Event.ID)
			assert.Equal(t, "type", resp.Event.TriggerType)

			var gotTrigger pb.TriggerEvent
			require.NoError(t, resp.Event.Payload.UnmarshalTo(&gotTrigger))
			require.Equal(t, fmt.Sprintf("id-%d", gotEvents), gotTrigger.Id)
		}
	})
}
