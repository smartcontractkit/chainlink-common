package triggers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// registerTrigger will do the following:
//
//  1. Register a trigger with the given feedIDs and triggerID
//  2. Return the trigger events channel, registerUnregisterRequest, and test context
func registerTrigger(
	ctx context.Context,
	t *testing.T,
	ts *MercuryTriggerService,
	feedIDs []string,
	triggerID string,
) (
	<-chan capabilities.TriggerResponse,
	capabilities.TriggerRegistrationRequest,
) {
	config, err := values.NewMap(map[string]interface{}{
		"feedIds":        feedIDs,
		"maxFrequencyMs": 100,
	})
	require.NoError(t, err)

	requestMetadata := capabilities.RequestMetadata{
		WorkflowID: "workflow-id-1",
	}
	registerRequest := capabilities.TriggerRegistrationRequest{
		Metadata:  requestMetadata,
		TriggerID: triggerID,
		Config:    config,
	}
	triggerEventsCh, err := ts.RegisterTrigger(ctx, registerRequest)
	require.NoError(t, err)

	return triggerEventsCh, registerRequest
}

const (
	triggerID = "streams-trigger@4.5.6"
	feedOne   = "0x1111111111111111111100000000000000000000000000000000000000000000"
	feedTwo   = "0x2222222222222222222200000000000000000000000000000000000000000000"
	feedThree = "0x3333333333333333333300000000000000000000000000000000000000000000"
	feedFour  = "0x4444444444444444444400000000000000000000000000000000000000000000"
	feedFive  = "0x5555555555555555555500000000000000000000000000000000000000000000"
)

func TestMercuryTrigger(t *testing.T) {
	ts, err := NewMercuryTriggerService(100, "", "4.5.6", logger.Nop())
	require.NoError(t, err)
	ctx := tests.Context(t)
	err = ts.Start(ctx)
	require.NoError(t, err)
	// use registerTriggerHelper to register a trigger
	callback, registerUnregisterRequest := registerTrigger(
		ctx,
		t,
		ts,
		[]string{feedOne},
		"test-id-1",
	)

	// Send events to trigger and check for them in the callback
	mfr := []datastreams.FeedReport{
		{
			FeedID:               feedOne,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(2).Bytes(),
			ObservationTimestamp: 3,
			Signatures:           [][]byte{},
		},
	}
	err = ts.ProcessReport(mfr)
	assert.NoError(t, err)
	msg := <-callback
	triggerEvent, reports := upwrapTriggerEvent(t, msg)
	assert.Equal(t, triggerID, triggerEvent.TriggerType)
	assert.Len(t, reports, 1)
	assert.Equal(t, mfr[0], reports[0])

	// Unregister the trigger and check that events no longer go on the callback
	require.NoError(t, ts.UnregisterTrigger(ctx, registerUnregisterRequest))
	err = ts.ProcessReport(mfr)
	require.NoError(t, err)
	require.Len(t, callback, 0)
	require.NoError(t, ts.Close())
}

func TestMultipleMercuryTriggers(t *testing.T) {
	ts, err := NewMercuryTriggerService(100, "", "4.5.6", logger.Nop())
	require.NoError(t, err)
	ctx := tests.Context(t)
	err = ts.Start(ctx)
	require.NoError(t, err)
	callback1, cr1 := registerTrigger(
		ctx,
		t,
		ts,
		[]string{
			feedOne,
			feedThree,
			feedFour,
		},
		"test-id-1",
	)

	callback2, cr2 := registerTrigger(
		ctx,
		t,
		ts,
		[]string{
			feedTwo,
			feedThree,
			feedFive,
		},
		"test-id-2",
	)

	// Send events to trigger and check for them in the callback
	mfr1 := []datastreams.FeedReport{
		{
			FeedID:               feedOne,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(20).Bytes(),
			ObservationTimestamp: 5,
			Signatures:           [][]byte{},
		},
		{
			FeedID:               feedThree,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(25).Bytes(),
			ObservationTimestamp: 8,
			Signatures:           [][]byte{},
		},
		{
			FeedID:               feedTwo,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(30).Bytes(),
			ObservationTimestamp: 10,
			Signatures:           [][]byte{},
		},
		{
			FeedID:               feedFour,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(40).Bytes(),
			ObservationTimestamp: 15,
			Signatures:           [][]byte{},
		},
	}

	err = ts.ProcessReport(mfr1)
	assert.NoError(t, err)

	msg := <-callback1
	triggerEvent, reports := upwrapTriggerEvent(t, msg)
	assert.Equal(t, triggerID, triggerEvent.TriggerType)
	assert.Len(t, reports, 3)
	assert.Equal(t, mfr1[0], reports[0])
	assert.Equal(t, mfr1[1], reports[1])
	assert.Equal(t, mfr1[3], reports[2])

	msg = <-callback2
	triggerEvent, reports = upwrapTriggerEvent(t, msg)
	assert.Equal(t, triggerID, triggerEvent.TriggerType)
	assert.Len(t, reports, 2)
	assert.Equal(t, mfr1[2], reports[0])
	assert.Equal(t, mfr1[1], reports[1])

	require.NoError(t, ts.UnregisterTrigger(ctx, cr1))
	mfr2 := []datastreams.FeedReport{
		{
			FeedID:               feedThree,
			FullReport:           []byte("0x1234"),
			BenchmarkPrice:       big.NewInt(50).Bytes(),
			ObservationTimestamp: 20,
		},
	}
	err = ts.ProcessReport(mfr2)
	assert.NoError(t, err)

	retryCount := 0
	for rMsg := range callback2 {
		triggerEvent, reports = upwrapTriggerEvent(t, rMsg)
		require.NoError(t, err)
		require.Len(t, reports, 2)
		require.Equal(t, triggerID, triggerEvent.TriggerType)
		price := big.NewInt(0).SetBytes(reports[1].BenchmarkPrice)
		if price.Cmp(big.NewInt(50)) == 0 {
			// expect to eventually get updated feed value
			break
		}
		require.True(t, retryCount < 100)
		retryCount++
	}

	require.NoError(t, ts.UnregisterTrigger(ctx, cr2))
	err = ts.ProcessReport(mfr1)
	assert.NoError(t, err)
	assert.Len(t, callback1, 0)
	assert.Len(t, callback2, 0)
	require.NoError(t, ts.Close())
}

func TestMercuryTrigger_RegisterTriggerErrors(t *testing.T) {
	ts, err := NewMercuryTriggerService(100, "", "4.5.6", logger.Nop())
	require.NoError(t, err)
	ctx := tests.Context(t)
	require.NoError(t, ts.Start(ctx))

	cm := map[string]interface{}{
		"feedIds":        []string{feedOne},
		"maxFrequencyMs": 90,
	}
	configWrapped, err := values.NewMap(cm)
	require.NoError(t, err)

	cr := capabilities.TriggerRegistrationRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID: "workflow-id-1",
		},
		Config:    configWrapped,
		TriggerID: "test-id-1",
	}
	_, err = ts.RegisterTrigger(ctx, cr)
	require.Error(t, err)

	cm = map[string]interface{}{
		"feedIds":        []string{feedOne},
		"maxFrequencyMs": 0,
	}
	configWrapped, err = values.NewMap(cm)
	require.NoError(t, err)
	cr.Config = configWrapped
	_, err = ts.RegisterTrigger(ctx, cr)
	require.Error(t, err)

	cm = map[string]interface{}{
		"feedIds":        []string{},
		"maxFrequencyMs": 1000,
	}
	configWrapped, err = values.NewMap(cm)
	require.NoError(t, err)
	cr.Config = configWrapped
	_, err = ts.RegisterTrigger(ctx, cr)
	require.Error(t, err)

	require.NoError(t, ts.Close())
}

func TestGetNextWaitIntervalMs(t *testing.T) {
	// getNextWaitIntervalMs args = (lastTs, tickerResolutionMs, currentTs)

	// expected cases
	assert.Equal(t, int64(900), getNextWaitIntervalMs(12000, 1000, 12100))
	assert.Equal(t, int64(200), getNextWaitIntervalMs(12000, 1000, 12800))

	// slow processing
	assert.Equal(t, int64(0), getNextWaitIntervalMs(12000, 1000, 13000))
	assert.Equal(t, int64(0), getNextWaitIntervalMs(12000, 1000, 14600))
}

func upwrapTriggerEvent(t *testing.T, req capabilities.TriggerResponse) (capabilities.TriggerEvent, []datastreams.FeedReport) {
	require.NotNil(t, req.Event.Outputs)
	mercuryReports, err := testMercuryCodec{}.Unwrap(req.Event.Outputs)
	require.NoError(t, err)
	return req.Event, mercuryReports
}

func TestMercuryTrigger_ConfigValidation(t *testing.T) {
	var newConfig = func(t *testing.T, feedIDs []string, maxFrequencyMs int) *values.Map {
		cm := map[string]interface{}{
			"feedIds":        feedIDs,
			"maxFrequencyMs": maxFrequencyMs,
		}
		configWrapped, err := values.NewMap(cm)
		require.NoError(t, err)

		return configWrapped
	}

	var newConfigSingleFeed = func(t *testing.T, feedID string) *values.Map {
		return newConfig(t, []string{feedID}, 1000)
	}

	ts, err := NewMercuryTriggerService(1000, "", "4.5.6", logger.Nop())
	require.NoError(t, err)
	rawConf := newConfigSingleFeed(t, "012345678901234567890123456789012345678901234567890123456789000000")
	conf, err := ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	rawConf = newConfigSingleFeed(t, "0x1234")
	conf, err = ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	rawConf = newConfigSingleFeed(t, "0x123zzz")
	conf, err = ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	rawConf = newConfigSingleFeed(t, "0x0001013ebd4ed3f5889FB5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292")
	conf, err = ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	passingFeedID := "0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292"
	// test maxfreq < 1
	rawConf = newConfig(t, []string{passingFeedID}, 0)
	conf, err = ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	rawConf = newConfig(t, []string{passingFeedID}, -1)
	conf, err = ts.ValidateConfig(rawConf)
	require.Error(t, err)
	require.Empty(t, conf)

	rawConf = newConfigSingleFeed(t, passingFeedID)
	conf, err = ts.ValidateConfig(rawConf)
	require.NoError(t, err)
	require.NotEmpty(t, conf)
}

func TestMercuryTrigger_WrapReports(t *testing.T) {
	S := 31   // signers
	P := 50   // feeds
	B := 1000 // report size in bytes
	meta := datastreams.Metadata{}
	for i := 0; i < S; i++ {
		meta.Signers = append(meta.Signers, randomByteArray(t, 20))
	}
	reportList := []datastreams.FeedReport{}
	for i := 0; i < P; i++ {
		signatures := [][]byte{}
		for j := 0; j < S; j++ {
			signatures = append(signatures, randomByteArray(t, 65))
		}
		reportList = append(reportList, datastreams.FeedReport{
			FeedID:               "0x" + hex.EncodeToString(randomByteArray(t, 32)),
			FullReport:           randomByteArray(t, B),
			ReportContext:        randomByteArray(t, 96),
			Signatures:           signatures,
			BenchmarkPrice:       big.NewInt(56789).Bytes(),
			ObservationTimestamp: 876543,
		})
	}
	wrapped, err := wrapReports(reportList, "event_id", 1234, meta, triggerID)
	require.NoError(t, err)
	require.NotNil(t, wrapped.Event)
	require.Len(t, wrapped.Event.Outputs.Underlying["Payload"].(*values.List).Underlying, P)
}

func randomByteArray(t *testing.T, n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	return b
}
