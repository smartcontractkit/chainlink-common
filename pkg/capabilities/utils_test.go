package capabilities_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
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

		req := capabilities.CapabilityRequest{ConfigPayload: a, Payload: a}

		resp, err := capabilities.Execute(context.Background(), req, &pb.TriggerEvent{}, &pb.TriggerEvent{}, func(_ context.Context, _ capabilities.RequestMetadata, i, c *pb.TriggerEvent) (*pb.TriggerEvent, error) {
			return &pb.TriggerEvent{Id: "out"}, nil
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Payload)
		assert.Nil(t, resp.Value)
	})

	t.Run("with value", func(t *testing.T) {
		msg := &pb.TriggerEvent{Id: "input"}
		wrapped, err := values.WrapMap(msg)
		require.NoError(t, err)

		req := capabilities.CapabilityRequest{Inputs: wrapped, Config: wrapped}

		resp, err := capabilities.Execute(context.Background(), req, &pb.TriggerEvent{}, &pb.TriggerEvent{}, func(_ context.Context, _ capabilities.RequestMetadata, i, c *pb.TriggerEvent) (*pb.TriggerEvent, error) {
			assert.Equal(t, "input", i.Id)
			assert.Equal(t, "input", c.Id)
			return &pb.TriggerEvent{Id: "out"}, nil
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Value)
		assert.Nil(t, resp.Payload)
	})
}

func TestRegisterTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := anypb.New(&pb.TriggerEvent{Id: "reg"})
	require.NoError(t, err)
	req := capabilities.TriggerRegistrationRequest{
		Metadata: capabilities.RequestMetadata{WorkflowID: "workflow-id"},
		Payload:  a,
	}

	respCh, err := capabilities.RegisterTrigger[*pb.TriggerEvent, *pb.TriggerEvent](
		ctx,
		"type",
		req,
		&pb.TriggerEvent{},
		func(_ context.Context, triggerID string, m capabilities.RequestMetadata, r *pb.TriggerEvent) (<-chan capabilities.TriggerAndId[*pb.TriggerEvent], error) {
			ch := make(chan capabilities.TriggerAndId[*pb.TriggerEvent], 1)
			ch <- capabilities.TriggerAndId[*pb.TriggerEvent]{
				Trigger: &pb.TriggerEvent{Id: "id"},
				Id:      "trigger-id",
			}
			assert.Equal(t, "workflow-id", m.WorkflowID)
			assert.Equal(t, "reg", r.Id)
			return ch, nil
		},
	)
	require.NoError(t, err)
	resp := <-respCh
	assert.Equal(t, "trigger-id", resp.Event.ID)
	assert.Equal(t, "type", resp.Event.TriggerType)
}
