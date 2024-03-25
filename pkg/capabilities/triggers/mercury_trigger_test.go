package triggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	feedOne   = mercury.Must(mercury.FromFeedIDString("0x1111111111111111111100000000000000000000000000000000000000000000"))
	feedTwo   = mercury.Must(mercury.FromFeedIDString("0x2222222222222222222200000000000000000000000000000000000000000000"))
	feedThree = mercury.Must(mercury.FromFeedIDString("0x3333333333333333333300000000000000000000000000000000000000000000"))
	feedFour  = mercury.Must(mercury.FromFeedIDString("0x4444444444444444444400000000000000000000000000000000000000000000"))
	feedFive  = mercury.Must(mercury.FromFeedIDString("0x5555555555555555555500000000000000000000000000000000000000000000"))
)

func TestMercuryTrigger(t *testing.T) {
	ts := NewMercuryTriggerService()
	ctx := tests.Context(t)
	require.NotNil(t, ts)

	cm := map[string]interface{}{
		"feedIds": []string{feedOne.String()},
	}

	configWrapped, err := values.NewMap(cm)
	require.NoError(t, err)

	im := map[string]interface{}{
		"triggerId": "test-id-1",
	}

	inputsWrapped, err := values.NewMap(im)
	require.NoError(t, err)

	cr := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID: "workflow-id-1",
		},
		Config: configWrapped,
		Inputs: inputsWrapped,
	}
	callback := make(chan capabilities.CapabilityResponse, 10)
	require.NoError(t, ts.RegisterTrigger(ctx, callback, cr))

	// Send events to trigger and check for them in the callback
	fr := []mercury.FeedReport{
		{
			FeedID:               feedOne,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       2,
			ObservationTimestamp: 3,
		},
	}
	err = ts.ProcessReport(fr)
	assert.NoError(t, err)
	assert.Len(t, callback, 1)
	msg := <-callback
	unwrapped, _ := mercury.Codec{}.UnwrapMercuryTriggerEvent(msg.Value)
	assert.Equal(t, "mercury", unwrapped.TriggerType)
	assert.Equal(t, GenerateTriggerEventID(fr), unwrapped.ID)
	assert.Len(t, unwrapped.Payload, 1)
	assert.Equal(t, fr[0], unwrapped.Payload[0])

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, cr))
	err = ts.ProcessReport(fr)
	assert.NoError(t, err)
	assert.Len(t, callback, 0)
}

func TestMultipleMercuryTriggers(t *testing.T) {
	ts := NewMercuryTriggerService()
	ctx := tests.Context(t)
	require.NotNil(t, ts)

	im1 := map[string]interface{}{
		"triggerId": "test-id-1",
	}

	iwrapped1, err := values.NewMap(im1)
	require.NoError(t, err)

	cm1 := map[string]interface{}{
		"feedIds": []string{
			feedOne.String(),
			feedThree.String(),
			feedFour.String(),
		},
	}

	cwrapped1, err := values.NewMap(cm1)
	require.NoError(t, err)

	cr1 := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID: "workflow-id-1",
		},
		Inputs: iwrapped1,
		Config: cwrapped1,
	}

	im2 := map[string]interface{}{
		"triggerId": "test-id-2",
	}

	iwrapped2, err := values.NewMap(im2)
	require.NoError(t, err)

	cm2 := map[string]interface{}{
		"feedIds": []string{
			feedTwo.String(),
			feedThree.String(),
			feedFive.String(),
		},
	}

	cwrapped2, err := values.NewMap(cm2)
	require.NoError(t, err)

	cr2 := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID: "workflow-id-1",
		},
		Inputs: iwrapped2,
		Config: cwrapped2,
	}

	callback1 := make(chan capabilities.CapabilityResponse, 10)
	callback2 := make(chan capabilities.CapabilityResponse, 10)

	require.NoError(t, ts.RegisterTrigger(ctx, callback1, cr1))
	require.NoError(t, ts.RegisterTrigger(ctx, callback2, cr2))

	// Send events to trigger and check for them in the callback
	fr1 := []mercury.FeedReport{
		{
			FeedID:               feedOne,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       20,
			ObservationTimestamp: 5,
		},
		{
			FeedID:               feedThree,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       25,
			ObservationTimestamp: 8,
		},
		{
			FeedID:               feedTwo,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       30,
			ObservationTimestamp: 10,
		},
		{
			FeedID:               feedFour,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       40,
			ObservationTimestamp: 15,
		},
	}

	err = ts.ProcessReport(fr1)
	assert.NoError(t, err)
	assert.Len(t, callback1, 1)
	assert.Len(t, callback2, 1)

	msg := <-callback1
	unwrapped, _ := mercury.Codec{}.UnwrapMercuryTriggerEvent(msg.Value)
	assert.Equal(t, "mercury", unwrapped.TriggerType)
	payload := make([]mercury.FeedReport, 0)
	payload = append(payload, fr1[0], fr1[1], fr1[3])
	assert.Equal(t, GenerateTriggerEventID(payload), unwrapped.ID)
	assert.Len(t, unwrapped.Payload, 3)
	assert.Equal(t, fr1[0], unwrapped.Payload[0])
	assert.Equal(t, fr1[1], unwrapped.Payload[1])
	assert.Equal(t, fr1[3], unwrapped.Payload[2])

	msg = <-callback2
	unwrapped, _ = mercury.Codec{}.UnwrapMercuryTriggerEvent(msg.Value)
	assert.Equal(t, "mercury", unwrapped.TriggerType)
	payload = make([]mercury.FeedReport, 0)
	payload = append(payload, fr1[1], fr1[2]) // Because GenerateTriggerEventID sorts the reports by feedID, this works
	assert.Equal(t, GenerateTriggerEventID(payload), unwrapped.ID)
	assert.Len(t, unwrapped.Payload, 2)
	assert.Equal(t, fr1[2], unwrapped.Payload[0])
	assert.Equal(t, fr1[1], unwrapped.Payload[1])

	require.NoError(t, ts.UnregisterTrigger(ctx, cr1))
	fr2 := []mercury.FeedReport{
		{
			FeedID:               feedThree,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       50,
			ObservationTimestamp: 20,
		},
	}
	err = ts.ProcessReport(fr2)
	assert.NoError(t, err)
	assert.Len(t, callback1, 0)
	assert.Len(t, callback2, 1)

	msg = <-callback2
	unwrapped, _ = mercury.Codec{}.UnwrapMercuryTriggerEvent(msg.Value)
	assert.Equal(t, "mercury", unwrapped.TriggerType)
	payload = make([]mercury.FeedReport, 0)
	payload = append(payload, fr2[0])
	assert.Equal(t, GenerateTriggerEventID(payload), unwrapped.ID)
	assert.Len(t, unwrapped.Payload, 1)
	assert.Equal(t, fr2[0], unwrapped.Payload[0])

	require.NoError(t, ts.UnregisterTrigger(ctx, cr2))
	err = ts.ProcessReport(fr1)
	assert.NoError(t, err)
	assert.Len(t, callback1, 0)
	assert.Len(t, callback2, 0)
}
